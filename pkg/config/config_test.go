package config_test

import (
	"testing"

	"github.com/nikore/gocast/pkg/config"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestConfig(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Gocast Config Suite")
}

var _ = Describe("Config", func() {
	It("should set defaults", func() {
		cfg, err := config.New([]byte(``))
		Expect(err).ShouldNot(HaveOccurred())
		Expect(cfg.Agent.ListenAddr).To(Equal("8080"))
		Expect(cfg.Agent.MonitorInterval.String()).To(Equal("10s"))
		Expect(cfg.Agent.CleanupTimer.String()).To(Equal("15m0s"))
	})
	It("should read config", func() {
		yaml := []byte(`
agent:
  listen_addr: 9090
  monitor_interval: 1s
  cleanup_timer: 5m
  
bgp:
  local_as: 1234
  remote_as: 4321
  peer_ip: 10.10.10.10
  communities:
    - asn:nnnn1
    - asn:nnnn2
  origin: igp

apps:
  - name: app1
    vip: 1.1.1.1/32
    vip_config:
      bgp_communities: [ aaaa:bbbb ]
    monitors: 
      - port:tcp:5000
`)
		cfg, err := config.New(yaml)
		Expect(err).ShouldNot(HaveOccurred())
		Expect(cfg.Agent.ListenAddr).To(Equal("9090"))
		Expect(cfg.Agent.MonitorInterval.String()).To(Equal("1s"))
		Expect(cfg.Agent.CleanupTimer.String()).To(Equal("5m0s"))

		Expect(cfg.Bgp.LocalAS).To(Equal(1234))
		Expect(cfg.Bgp.PeerAS).To(Equal(4321))
		Expect(cfg.Bgp.PeerIP).To(Equal("10.10.10.10"))
		Expect(cfg.Bgp.Communities).Should(HaveLen(2))
		Expect(cfg.Bgp.Origin).To(Equal("igp"))

		Expect(cfg.Apps).Should(HaveLen(1))
		Expect(cfg.Apps[0].Name).To(Equal("app1"))
		Expect(cfg.Apps[0].Vip).To(Equal("1.1.1.1/32"))
		Expect(cfg.Apps[0].VipConfig.BgpCommunities[0]).To(Equal("aaaa:bbbb"))
		Expect(cfg.Apps[0].Monitors[0]).To(Equal("port:tcp:5000"))
	})
})
