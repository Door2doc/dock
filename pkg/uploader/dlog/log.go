// Package dlog provides a door2doc logger that can be integrated with service logging.
package dlog

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/kardianos/service"
)

var (
	svc      service.Logger
	username string
)

// SetService uses the service log for all subsequent logging.
func SetService(s service.Service) {
	logger, err := s.Logger(nil)
	if err != nil {
		log.Panicf("Failed to initialize service log: %v", err)
	}
	svc = logger
}

// SetUsername sets the username scope for all subsequent logging
func SetUsername(s string) {
	username = s
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
	msg := fmt.Sprintf(pattern, args...)

	if svc != nil {
		go submitError(username, msg)
	}

	if svc == nil {
		log.Println(msg)
		return
	}

	err := svc.Error(msg)
	if err != nil {
		log.Println(msg)
		log.Println(err)
	}
}

func submitError(user string, msg string) {
	req, err := http.NewRequest(http.MethodPost, "https://integration.door2doc.net/services/v3/upload/feedback", nil)
	if err != nil {
		return
	}
	q := req.URL.Query()
	q.Add("username", user)
	q.Add("message", msg)
	req.URL.RawQuery = q.Encode()

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return
	}
	_, _ = io.Copy(ioutil.Discard, res.Body)
	_ = res.Body.Close()
}

func Close(c io.Closer) {
	err := c.Close()
	if err != nil {
		Error("Failed to close %v: %v", c, err)
	}
}
