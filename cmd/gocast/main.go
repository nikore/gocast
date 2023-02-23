package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/golang/glog"
	c "github.com/nikore/gocast/pkg/config"
	"github.com/nikore/gocast/pkg/controller"
	"github.com/nikore/gocast/pkg/server"
	log "github.com/sirupsen/logrus"
)

var (
	config = flag.String("config", "", "Path to config file")
)

func main() {
	flag.Parse()
	if glog.V(4) {
		log.SetLevel(log.DebugLevel)
	}
	conf := c.GetConfig(*config)
	mon := controller.NewMonitor(conf)
	srv := server.NewServer(conf.Agent.ListenAddr, mon)

	ctx, cancel := context.WithCancel(context.Background())
	// catch interrupt
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGHUP, syscall.SIGTERM)
	go func() {
		for {
			sig := <-signalChan
			if sig == os.Interrupt || sig == syscall.SIGTERM {
				mon.CloseAll()
				cancel()
				return
			}
		}
	}()
	srv.Serve(ctx)
}
