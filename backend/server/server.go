package server

import (
	"context"
	"crawlab/app/master/config"
	"crawlab/embed/raftserver/membership"
	"crawlab/pkg/core/broker"
	"crawlab/runtime"
	"crawlab/server/master"
	"errors"
	"fmt"
	"github.com/apex/log"
	"github.com/gin-gonic/gin"
	"github.com/lni/dragonboat/v3"
	config2 "github.com/lni/dragonboat/v3/config"
	"github.com/urfave/cli/v2"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"
)

func watchModeChange() error {
	sub, err := runtime.Bus.Subscribe(runtime.ModeChange, func(event broker.Event) error {
		message := event.Message()
		body := message.Body.(runtime.EnvMode)
		switch body {
		case runtime.Development:
			gin.SetMode(gin.DebugMode)
		case runtime.Production:
			gin.SetMode(gin.ReleaseMode)
		case runtime.Test:
			gin.SetMode(gin.TestMode)
		default:
			panic("unknown mode")
		}
		return nil
	})
	if err != nil {
		return err
	}
	defer sub.Unsubscribe()
	return nil
}

type AppLauncher struct {
	config *config.ApplicationConfig
}

func (b *AppLauncher) before() (err error) {
	err = watchModeChange()
	if err != nil {
		return err
	}
	return nil
}
func (b *AppLauncher) instanceMasterNode() *http.Server {
	return master.NewMasterNode(b.config)
}
func (b *AppLauncher) instanceWorkerNode() *http.Server {
	//return WorkerNode(b.config)
	return master.NewMasterNode(b.config)

}
func (b *AppLauncher) selectServer() (srv *http.Server, err error) {
	switch b.config.NodeType {
	case runtime.Master:
		return b.instanceMasterNode(), nil
	case runtime.Worker:
		return b.instanceWorkerNode(), nil
	}
	return nil, errors.New("error")
}
func (b *AppLauncher) F() error {

	//if err:= nh.StartOnDiskCluster()
	return nil
}
func (b *AppLauncher) Run() error {
	err := b.before()
	if err != nil {
		return err
	}
	runtime.SetMode(b.config.Mode)
	address := net.JoinHostPort(b.config.Server.Host, strconv.Itoa(b.config.Server.Port))

	srv, err := b.selectServer()
	if err != nil {
		return err
	}
	srv.Addr = address
	go func() {
		if err := srv.ListenAndServe(); err != nil {
			if err != http.ErrServerClosed {
				log.Error("run server error:" + err.Error())
			} else {
				log.Info("server graceful down")
			}
		}
	}()
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	ctx2, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx2); err != nil {
		log.Error("run server error:" + err.Error())
	}
	return nil
}
func Launcher(config *config.ApplicationConfig) (*AppLauncher, error) {
	return &AppLauncher{config: config}, nil
}
