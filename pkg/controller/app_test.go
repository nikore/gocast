package controller

import (
	"testing"

	"github.com/nikore/gocast/pkg/config"
	"github.com/stretchr/testify/assert"
)

func TestAppParsing(t *testing.T) {
	a := assert.New(t)
	app1, err := NewApp("app1", "1.1.1.1/32", config.VipConfig{}, []string{"port:tcp:123"}, []string{}, "")
	a.Nil(err)
	app2, err := NewApp("app1", "1.1.1.1/32", config.VipConfig{BgpCommunities: []string{"111:222"}}, []string{"port:tcp:123"}, []string{}, "")
	a.Nil(err)
	app3, err := NewApp("app3", "2.2.2.2/32", config.VipConfig{}, []string{"exec:/bin/testme"}, []string{}, "")
	a.Nil(err)

	a.Equal("1.1.1.1/32", app1.Vip.Net.String())
	a.Equal(Monitor_PORT, app1.Monitors[0].Type)
	a.Equal("123", app1.Monitors[0].Port)
	a.Equal("tcp", app1.Monitors[0].Protocol)
	a.Equal(config.VipConfig{}, app1.VipConfig)

	a.Equal(true, app1.Equal(app2))

	a.Equal("111:222", app2.Vip.Communities[0])

	a.Equal(Monitor_EXEC, app3.Monitors[0].Type)
	a.Equal("/bin/testme", app3.Monitors[0].Cmd)

	// test errors
	_, err = NewApp("app4", "4.4.4.4", config.VipConfig{}, []string{}, []string{}, "")
	a.NotNil(err)

	_, err = NewApp("app4", "4.4.4.4/32", config.VipConfig{}, []string{"port:abcd::1023"}, []string{}, "")
	a.NotNil(err)
}
