package web

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/gin-gonic/gin"
	"harnsgateway/cmd/gateway/config"
	"harnsgateway/cmd/gateway/options"
	"harnsgateway/pkg/device"
	"harnsgateway/pkg/gateway"
	"harnsgateway/pkg/generic"
	"k8s.io/klog/v2"
	"net/http"
)

type Server struct {
	*generic.Server
	*config.Config
}

func NewServer(router *gin.Engine, o *options.Options, config *config.Config) (*Server, error) {
	allowMethods := []string{http.MethodPost, http.MethodGet, http.MethodDelete, http.MethodPut, http.MethodPatch}

	s := &generic.Server{
		Router:  router,
		Port:    o.Port,
		Methods: allowMethods,
	}

	server := &Server{
		Server: s,
		Config: config,
	}

	server.InstallHandlers()

	return server, nil

}

func (s *Server) InstallHandlers() {
	v1 := s.Router.Group("/api/v1")
	device.InstallHandler(v1, s.Config.DeviceMgr)
	gateway.InstallHandler(v1, s.Config.GatewayMgr)
}

func (s *Server) Serve() (func(ctx context.Context), error) {
	var srv *http.Server
	if len(s.Config.CertFile) != 0 && len(s.Config.KeyFile) != 0 {
		x509KeyPair, err := tls.LoadX509KeyPair(s.Config.CertFile, s.Config.KeyFile)
		if err != nil {
			return nil, err
		}
		c := &tls.Config{
			Certificates: []tls.Certificate{x509KeyPair},
		}

		srv = &http.Server{
			Addr:      fmt.Sprintf(":%s", s.Port),
			Handler:   s.Router,
			TLSConfig: c,
		}
		go func() {
			klog.Error(srv.ListenAndServeTLS("", ""))
		}()
	} else {
		srv = &http.Server{
			Addr:    fmt.Sprintf(":%s", s.Port),
			Handler: s.Router,
		}
		go func() {
			klog.Error(srv.ListenAndServe())
		}()
	}

	return func(ctx context.Context) {
		srv.SetKeepAlivesEnabled(false)
		if err := s.Config.DeviceMgr.Shutdown(ctx); err != nil {
			klog.Error(err)
		}
		if err := srv.Shutdown(ctx); err != nil {
			klog.Error(err)
		}
	}, nil
}
