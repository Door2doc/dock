// A windows service for uploading Chipsoft extracts to door2doc.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/denisenkom/go-mssqldb"
	"github.com/door2doc/d2d-uploader/pkg/uploader"
	"github.com/door2doc/d2d-uploader/pkg/uploader/dlog"
	"github.com/kardianos/service"
	"github.com/lib/pq"
)

var _ mssql.Driver
var _ pq.Driver

var (
	DevelopmentMode = flag.Bool("dev", false, "Run in development mode")
)

func main() {
	flag.Parse()

	config := &service.Config{
		Name:        "Door2docUploader",
		DisplayName: "Door2doc Upload Service",
		Description: "This service takes care of regular uploads to door2doc",
	}
	svc := uploader.NewService(*DevelopmentMode, Version)
	s, err := service.New(svc, config)
	if err != nil {
		log.Fatalf("Failed to construct service: %v", err)
	}

	if flag.NArg() == 1 {
		switch action := flag.Arg(0); action {
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

	dlog.SetService(s)
	err = s.Run()
	if err != nil {
		dlog.Error("Failed to run service: %v", err)
	}
}
