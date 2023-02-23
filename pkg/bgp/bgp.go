package bgp

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/golang/protobuf/ptypes/any"
	"github.com/nikore/gocast/pkg/config"
	api "github.com/osrg/gobgp/v3/api"
	gobgp "github.com/osrg/gobgp/v3/pkg/server"
	"google.golang.org/protobuf/types/known/anypb"
)

type Controller interface {
	AddPeer(peer string) error
	Announce(route *Route) error
	Withdraw(route *Route) error
	PeerInfo() (*api.Peer, error)
	Shutdown() error
}

type Route struct {
	Net         *net.IPNet
	Communities []string
}

type controller struct {
	peerAS          int
	localIP, peerIP net.IP
	communities     []string
	origin          uint32
	multiHop        bool
	bgpServer       *gobgp.BgpServer
}

func NewController(cfg config.BgpConfig) (Controller, error) {
	bgpServer := gobgp.NewBgpServer()
	go bgpServer.Serve()
	if err := bgpServer.StartBgp(context.Background(), &api.StartBgpRequest{
		Global: &api.Global{
			Asn:        uint32(cfg.LocalAS),
			RouterId:   cfg.LocalIP,
			ListenPort: -1, // gobgp won't listen on tcp:179
		},
	}); err != nil {
		return nil, fmt.Errorf("unable to start bgp: %v", err)
	}
	return &controller{
		peerIP:      net.ParseIP(cfg.PeerIP),
		localIP:     net.ParseIP(cfg.LocalIP),
		communities: cfg.Communities,
		peerAS:      cfg.PeerAS,
		multiHop: func() bool {
			if cfg.PeerAS != cfg.LocalAS {
				return true
			}
			return false
		}(),
		origin: func() uint32 {
			switch cfg.Origin {
			case "igp":
				return 0
			case "egp":
				return 1
			default:
				return 2
			}
		}(),
		bgpServer: bgpServer,
	}, nil
}

func (c *controller) AddPeer(peer string) error {
	n := &api.Peer{
		Conf: &api.PeerConf{
			NeighborAddress: peer,
			PeerAsn:         uint32(c.peerAS),
		},
	}
	if c.multiHop {
		n.EbgpMultihop = &api.EbgpMultihop{Enabled: true, MultihopTtl: uint32(255)}
	}
	return c.bgpServer.AddPeer(context.Background(), &api.AddPeerRequest{Peer: n})
}

func (c *controller) getApiPath(route *Route) *api.Path {
	afi := api.Family_AFI_IP
	if route.Net.IP.To4() == nil {
		afi = api.Family_AFI_IP6
	}
	prefixlen, _ := route.Net.Mask.Size()
	nlri, _ := anypb.New(&api.IPAddressPrefix{
		Prefix:    route.Net.IP.String(),
		PrefixLen: uint32(prefixlen),
	})
	a1, _ := anypb.New(&api.OriginAttribute{
		Origin: c.origin,
	})
	a2, _ := anypb.New(&api.NextHopAttribute{
		NextHop: c.localIP.String(),
	})
	var communities []uint32
	for _, comm := range append(c.communities, route.Communities...) {
		communities = append(communities, convertCommunity(comm))
	}
	a3, _ := anypb.New(&api.CommunitiesAttribute{
		Communities: communities,
	})
	attrs := []*any.Any{a1, a2, a3}
	return &api.Path{
		Family: &api.Family{Afi: afi, Safi: api.Family_SAFI_UNICAST},
		Nlri:   nlri,
		Pattrs: attrs,
	}
}

func (c *controller) Announce(route *Route) error {
	var found bool
	err := c.bgpServer.ListPeer(context.Background(), &api.ListPeerRequest{}, func(p *api.Peer) {
		if p.Conf.NeighborAddress == c.peerIP.String() {
			found = true
		}
	})
	if err != nil {
		return err
	}
	if !found {
		if err := c.AddPeer(c.peerIP.String()); err != nil {
			return err
		}
	}
	_, err = c.bgpServer.AddPath(context.Background(), &api.AddPathRequest{Path: c.getApiPath(route)})
	return err
}

func (c *controller) Withdraw(route *Route) error {
	return c.bgpServer.DeletePath(context.Background(), &api.DeletePathRequest{Path: c.getApiPath(route)})
}

func (c *controller) PeerInfo() (*api.Peer, error) {
	var peer *api.Peer
	err := c.bgpServer.ListPeer(context.Background(), &api.ListPeerRequest{}, func(p *api.Peer) {
		if p.Conf.NeighborAddress == c.peerIP.String() {
			peer = p
		}
	})
	if err != nil {
		return nil, err
	}
	return peer, nil
}

func (c *controller) Shutdown() error {
	if err := c.bgpServer.ShutdownPeer(context.Background(), &api.ShutdownPeerRequest{
		Address: c.peerIP.String(),
	}); err != nil {
		return err
	}
	if err := c.bgpServer.StopBgp(context.Background(), &api.StopBgpRequest{}); err != nil {
		return err
	}
	return nil
}

func convertCommunity(comm string) uint32 {
	parts := strings.Split(comm, ":")
	first, _ := strconv.ParseUint(parts[0], 10, 32)
	second, _ := strconv.ParseUint(parts[1], 10, 32)
	return uint32(first)<<16 | uint32(second)
}
