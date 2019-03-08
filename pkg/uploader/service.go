package uploader

import (
	"context"
	"net"
	"net/http"
	"time"

	"github.com/kardianos/service"
	"github.com/publysher/d2d-uploader/pkg/uploader/config"
	"github.com/publysher/d2d-uploader/pkg/uploader/web"
)

// Service contains the definition to run the application as a service.
type Service struct {
	dev      bool
	version  string
	shutdown context.CancelFunc
	log      service.Logger
	srv      *http.Server
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
	// obtain logger
	logger, err := svc.Logger(nil)
	if err != nil {
		return err
	}
	s.log = logger

	// create HTTP server for configuration purposes
	handler, err := web.NewServeMux(s.log, s.dev, s.version)
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
		defer func() {
			if err := ln.Close(); err != nil {
				_ = logger.Errorf("While closing listener: %v", err)
			}
		}()

		_ = s.log.Info("Starting configuration server")
		err := s.srv.Serve(ln)

		if err != nil && err != http.ErrServerClosed {
			_ = logger.Errorf("While running configuration server: %v", err)
		}
	}()

	// run service
	go func() {
		err := s.run(ctx)
		switch {
		case err == context.Canceled:
			// result of shutdown, we're OK
			_ = logger.Info("Service stopped")
		case err != nil:
			_ = logger.Errorf("While running service: %v", err)
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
	if err == nil {
		_ = s.log.Info("Configuration server stopped")
	} else {
		logger, lErr := svc.Logger(nil)
		if lErr != nil {
			_ = logger.Errorf("Failed to stop configuration server: %v", err)
		}
	}

	return nil
}

func (s *Service) run(ctx context.Context) error {
	// obtain latest version of the configuration
	cfg, err := config.Load(ctx)
	if err != nil {
		_ = s.log.Errorf("Failed to load configuration: %v", err)
		return err
	}
	_ = s.log.Infof("Starting service")

	for {
		// refresh configuration
		if err := cfg.Refresh(); err != nil {
			_ = s.log.Errorf("Failed to refresh configuration: %v", err)
		}

		// run the upload, IF the configuration is active
		sleep := time.Second
		if cfg.Active {
			if err := Upload(ctx, s.log, cfg); err != nil {
				_ = s.log.Errorf("While processing upload: %v", err)
			}
			sleep = cfg.Interval
		}

		// sleep until the next iteration
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(sleep):
		}
	}
}
