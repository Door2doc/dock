package main

import (
	"fmt"
	"log"
	"os"

	"github.com/denisenkom/go-mssqldb"
	"github.com/kardianos/service"
	"github.com/publysher/d2d-uploader/pkg/uploader"
)

var _ mssql.Driver

func main() {
	config := &service.Config{
		Name:        "Door2docUploader",
		DisplayName: "Door2doc Upload Service",
		Description: "This service takes care of regular uploads to door2doc",
	}
	svc := uploader.NewService()
	s, err := service.New(svc, config)
	if err != nil {
		log.Fatalf("Failed to construct service: %v", err)
	}

	if len(os.Args) > 1 {
		switch action := os.Args[1]; action {
		case "install", "uninstall", "start", "stop", "restart":
			err := service.Control(s, action)
			if err != nil {
				log.Fatalf("Failed to %s service: %v", action, err)
			}
			var past string
			switch action {
			case "stop":
				past = "stopped"
			default:
				past = action + "ed"
			}

			log.Printf("Successfully %s service", past)
		case "help":
			PrintVersion(os.Stdout)
			fmt.Println("Valid options are:")
			fmt.Println()
			fmt.Println("\tinstall     Install the service")
			fmt.Println("\tuninstall   Uninstall the service")
			fmt.Println("\tstart       Start the service")
			fmt.Println("\tstop        Stop the service")
			fmt.Println("\trestart     Restart the service")
			return
		}

		return
	}

	lg, err := s.Logger(nil)
	if err != nil {
		log.Fatalf("Failed to register logger: %v", err)
	}

	err = s.Run()
	if err != nil {
		_ = lg.Errorf("Failed to run service: %v", err)
	}
}
