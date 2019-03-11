// Package dlog provides a door2doc logger.
package dlog

import (
	"io"
	"log"

	"github.com/kardianos/service"
)

var (
	svc service.Logger
)

// SetService uses the service log for all subsequent logging.
func SetService(s service.Service) {
	logger, err := s.Logger(nil)
	if err != nil {
		log.Panicf("Failed to initialize service log: %v", err)
	}
	svc = logger
}

func Info(pattern string, args ...interface{}) {
	if svc == nil {
		log.Printf(pattern, args...)
		return
	}

	err := svc.Infof(pattern, args...)
	if err != nil {
		log.Printf(pattern, args...)
		log.Println(err)
	}
}

func Error(pattern string, args ...interface{}) {
	if svc == nil {
		log.Printf(pattern, args...)
		return
	}

	err := svc.Errorf(pattern, args...)
	if err != nil {
		log.Printf(pattern, args...)
		log.Println(err)
	}
}

func Close(c io.Closer) {
	err := c.Close()
	if err != nil {
		Error("Failed to close %v: %v", c, err)
	}
}
