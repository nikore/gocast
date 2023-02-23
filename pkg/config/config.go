package config

import (
	"bytes"
	"time"

	"github.com/spf13/viper"
)

type AgentConfig struct {
	ListenAddr      string        `mapstructure:"listen_addr"`
	MonitorInterval time.Duration `mapstructure:"monitor_interval"`
	CleanupTimer    time.Duration `mapstructure:"cleanup_timer"`
}

type BgpConfig struct {
	LocalAS     int      `mapstructure:"local_as"`
	PeerAS      int      `mapstructure:"remote_as"`
	LocalIP     string   `mapstructure:"lcoal_ip"`
	PeerIP      string   `mapstructure:"peer_ip"`
	Communities []string `mapstructure:"communities"`
	Origin      string   `mapstructure:"origin"`
}

type VipConfig struct {
	// per VIP BGP communities to announce. This is in addition to the
	// global config
	BgpCommunities []string `mapstructure:"bgp_communities"`
}

type AppConfig struct {
	Name      string    `mapstructure:"name"`
	Vip       string    `mapstructure:"vip"`
	VipConfig VipConfig `mapstructure:"vip_config"`
	Monitors  []string  `mapstructure:"monitors"`
}

type Config struct {
	Agent AgentConfig `mapstructure:"agent"`
	Bgp   BgpConfig   `mapstructure:"bgp"`
	Apps  []AppConfig `mapstructure:"apps"`
}

func New(configData []byte) (Config, error) {
	viper.SetConfigType("yaml")
	viper.AutomaticEnv()
	viper.SetEnvPrefix("GC")

	viper.SetDefault("agent.listen_addr", "8080")
	viper.SetDefault("agent.monitor_interval", "10s")
	viper.SetDefault("agent.cleanup_timer", "15m")

	c := Config{}

	if err := viper.ReadConfig(bytes.NewBuffer(configData)); err != nil {
		return c, err
	}

	if err := viper.Unmarshal(&c); err != nil {
		return c, err
	}

	return c, nil
}
