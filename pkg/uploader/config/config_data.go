package config

import (
	"github.com/door2doc/d2d-uploader/pkg/uploader/db"
	"github.com/door2doc/d2d-uploader/pkg/uploader/password"
	"time"
)

type DataV1 struct {
	Username        string        `json:"username"`
	Password        string        `json:"password"`
	Proxy           string        `json:"proxy"`
	Connection      string        `json:"dsn"`
	Timeout         time.Duration `json:"timeout"`
	VisitorQuery    string        `json:"query"`
	RadiologieQuery string        `json:"radiologie"`
	LabQuery        string        `json:"lab"`
	ConsultQuery    string        `json:"consult"`
	AccessUsername  string        `json:"accessUsername"`
	AccessPassword  string        `json:"accessPassword"`
}

func (v DataV1) ToV2() (DataV2, error) {
	connectionData, err := db.FromDSN(v.Connection)
	if err != nil {
		return DataV2{}, err
	}

	return DataV2{
		Version:         1,
		Username:        v.Username,
		Password:        password.Password(v.Password),
		Proxy:           v.Proxy,
		Connection:      connectionData,
		Timeout:         v.Timeout,
		VisitorQuery:    v.VisitorQuery,
		RadiologieQuery: v.RadiologieQuery,
		LabQuery:        v.LabQuery,
		ConsultQuery:    v.ConsultQuery,
		AccessUsername:  v.AccessUsername,
		AccessPassword:  password.Password(v.AccessPassword),
	}, nil
}

type DataV2 struct {
	Version         int
	Username        string            `json:"username"`
	Password        password.Password `json:"password"`
	Proxy           string            `json:"proxy"`
	Connection      db.ConnectionData `json:"database"`
	Timeout         time.Duration     `json:"timeout"`
	VisitorQuery    string            `json:"query"`
	RadiologieQuery string            `json:"radiologie"`
	LabQuery        string            `json:"lab"`
	ConsultQuery    string            `json:"consult"`
	AccessUsername  string            `json:"access_username"`
	AccessPassword  password.Password `json:"access_password"`
}
