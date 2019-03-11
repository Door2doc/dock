package uploader

import (
	"context"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/kardianos/service"
	"github.com/publysher/d2d-uploader/pkg/uploader/config"
	"github.com/publysher/d2d-uploader/pkg/uploader/dlog"
	"github.com/publysher/d2d-uploader/pkg/uploader/web"
)

// Service contains the definition to run the application as a service.
type Service struct {
	dev      bool
	version  string
	shutdown context.CancelFunc
	srv      *http.Server
	cfg      *config.Configuration
}

// NewService creates a new Service instance.
func NewService(development bool, version string) *Service {
	return &Service{
		dev:     development,
		version: version,
	}
}

// Start starts running the service. It will return as soon as possible.
func (s *Service) Start(svc service.Service) error {
	// load configuration
	s.cfg = config.NewConfiguration()
	if err := s.cfg.Reload(); err != nil {
		return err
	}

	// create HTTP server for configuration purposes
	handler, err := web.NewServeMux(s.dev, s.version, s.cfg)
	if err != nil {
		return err
	}

	mux := http.DefaultServeMux
	mux.Handle("/", handler)
	s.srv = &http.Server{
		Addr:    ":17226",
		Handler: mux,
	}

	// start listening, and fail service start if port is occupied
	ln, err := net.Listen("tcp", s.srv.Addr)
	if err != nil {
		return err
	}

	// create service context
	ctx, cancel := context.WithCancel(context.Background())
	s.shutdown = cancel

	// run HTTP server
	go func() {
		addr := s.srv.Addr
		if strings.HasPrefix(addr, ":") {
			addr = "localhost" + addr
		}
		addr = "http://" + addr

		dlog.Info("Starting configuration server on %s", addr)
		err := s.srv.Serve(ln)

		if err != nil && err != http.ErrServerClosed {
			dlog.Error("While running configuration server: %v", err)
		}
	}()

	// run service
	go func() {
		err := s.run(ctx)
		switch {
		case err == context.Canceled:
			// result of shutdown, we're OK
			dlog.Info("Service stopped")
		case err != nil:
			dlog.Error("While running service: %v", err)
			_ = svc.Stop()
		}
	}()

	return nil
}

// Stop stops the currently running service. It will return as soon as possible.
func (s *Service) Stop(svc service.Service) error {
	// call global context shutdown
	if s.shutdown != nil {
		s.shutdown()
	}

	// shutdown HTTP server
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	err := s.srv.Shutdown(ctx)
	if err != nil {
		dlog.Error("Failed to stop configuration server: %v", err)
		return err
	}

	dlog.Info("Configuration server stopped")
	return nil
}

func (s *Service) run(ctx context.Context) error {
	dlog.Info("Starting service")

	for {
		// run the upload, IF the configuration is active
		sleep := time.Second
		if s.cfg.Active() {
			if err := Upload(ctx, s.cfg); err != nil {
				dlog.Error("While processing upload: %v", err)
			}
			sleep = s.cfg.Interval()
		}

		// sleep until the next iteration
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(sleep):
		}
	}
}
