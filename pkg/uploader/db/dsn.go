package db

import (
	"bytes"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/door2doc/d2d-uploader/pkg/uploader/dlog"
)

type ConnectionData struct {
	Driver   string
	Host     string
	Port     string
	Instance string
	Database string
	Username string
	Password string
	Params   string
}

func (c ConnectionData) IsValid() bool {
	return c.Driver != "" && c.Host != ""
}

// Implementation of encoding.TextUnmarshaler
func (c *ConnectionData) UnmarshalText(bs []byte) error {
	if bytes.HasPrefix(bs, []byte("postgres://")) {
		c.Driver = "postgres"
		return c.unmarshalPostgres(bs)
	}

	c.Driver = "sqlserver"
	return c.unmarshalSqlServer(bs)
}

func (c *ConnectionData) unmarshalPostgres(bs []byte) error {
	u, err := url.Parse(string(bs))
	if err != nil {
		return err
	}
	c.Host = u.Hostname()
	c.Port = u.Port()
	c.Database = strings.TrimPrefix(u.Path, "/")
	c.Username = u.User.Username()
	c.Password, _ = u.User.Password()
	c.Params = u.RawQuery
	return nil
}

func (c *ConnectionData) unmarshalSqlServer(bs []byte) error {
	parts := strings.Split(string(bs), ";")
	for _, p := range parts {
		keyVal := strings.Split(strings.TrimSpace(p), "=")
		if len(keyVal) != 2 {
			return errors.New("invalid connection string")
		}
		key, val := keyVal[0], keyVal[1]
		switch strings.ToLower(key) {
		case "server":
			serverInstance := strings.Split(val, "\\")
			switch len(serverInstance) {
			case 1:
				c.Host = serverInstance[0]
			case 2:
				c.Host = serverInstance[0]
				c.Instance = serverInstance[1]
			default:
				return errors.New("invalid server")
			}
		case "database":
			c.Database = val
		case "user id":
			c.Username = val
		case "password":
			c.Password = val
		case "integrated security":
			c.Username = ""
			c.Password = ""
		default:
			if len(c.Params) != 0 {
				c.Params += ";"
			}
			c.Params += key + "=" + val
		}
		_ = val
	}

	return nil
}

// Implementation of encoding.TextMarshaler
func (c ConnectionData) MarshalText() ([]byte, error) {
	if c.Driver == "postgres" {
		return c.marshalPostgres()
	}

	return c.marshalSqlServer()
}

func (c ConnectionData) marshalPostgres() ([]byte, error) {
	u := &url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(c.Username, c.Password),
		Host:     c.Host,
		Path:     "/" + c.Database,
		RawQuery: c.Params,
	}
	if c.Port != "" {
		u.Host = u.Host + ":" + c.Port
	}
	return []byte(u.String()), nil
}

func (c *ConnectionData) marshalSqlServer() ([]byte, error) {
	var parts []string
	server := c.Host
	if c.Instance != "" {
		server += "\\" + c.Instance
	}
	parts = append(parts, fmt.Sprintf("server=%s", server))
	if c.Database != "" {
		parts = append(parts, fmt.Sprintf("database=%s", c.Database))
	}
	if c.Username != "" {
		parts = append(parts, fmt.Sprintf("user id=%s", c.Username))
	}
	if c.Password != "" {
		parts = append(parts, fmt.Sprintf("password=%s", c.Password))
	}
	if c.Username == "" && c.Password == "" {
		parts = append(parts, "integrated security=SSPI")
	}
	if c.Params != "" {
		parts = append(parts, c.Params)
	}

	dsn := strings.Join(parts, "; ")
	return []byte(dsn), nil
}

func (c *ConnectionData) DSN() string {
	res, err := c.MarshalText()
	if err != nil {
		dlog.Error("failed to convert DSN: %v", err)
		return ""
	}
	return string(res)
}
