package config

import (
	"github.com/door2doc/d2d-uploader/pkg/uploader/db"
	"time"
)

type ConfigDataV1 struct {
	Username        string            `json:"username"`
	Password        string            `json:"password"`
	Proxy           string            `json:"proxy"`
	Connection      db.ConnectionData `json:"dsn"`
	Timeout         time.Duration     `json:"timeout"`
	VisitorQuery    string            `json:"query"`
	RadiologieQuery string            `json:"radiologie"`
	LabQuery        string            `json:"lab"`
	ConsultQuery    string            `json:"consult"`
	AccessUsername  string            `json:"accessUsername"`
	AccessPassword  string            `json:"accessPassword"`
}
