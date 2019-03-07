package uploader

import (
	"context"
	"time"

	"github.com/kardianos/service"
)

// Service contains the definition to run the application as a service.
type Service struct {
	shutdown context.CancelFunc
	log      service.Logger
}

// NewService creates a new Service instance.
func NewService() *Service {
	return &Service{}
}

// Start starts running the service. It will return as soon as possible.
func (s *Service) Start(svc service.Service) error {
	logger, err := svc.Logger(nil)
	if err != nil {
		return err
	}
	s.log = logger

	ctx, cancel := context.WithCancel(context.Background())
	s.shutdown = cancel

	go func() {
		err := s.run(ctx)
		switch {
		case err == context.Canceled:
			// result of shutdown, we're OK
			_ = logger.Info("Service stopped")
		case err != nil:
			_ = logger.Errorf("While running service: %v", err)
		}
		if err != nil {
			_ = logger.Errorf("While running service: %v", err)
		}
	}()
	return nil
}

func (s *Service) run(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.Tick(time.Second):
			_ = s.log.Infof("tick")
		}
	}
}

// Stop stops the currently running service. It will return as soon as possible.
func (s *Service) Stop(svc service.Service) error {
	if s.shutdown != nil {
		s.shutdown()
	}
	return nil
}
