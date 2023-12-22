package db

import (
	"errors"
	"fmt"
	"github.com/door2doc/d2d-uploader/pkg/uploader/password"
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
	Password password.Password
	Params   string
}

func (c ConnectionData) IsValid() bool {
	return c.Driver != "" && c.Host != ""
}

func FromDSN(s string) (ConnectionData, error) {
	if strings.HasPrefix(s, "postgres://") {
		return fromPostgresDSN(s)
	}

	return fromSqlServerDSN(s)
}

func fromPostgresDSN(s string) (ConnectionData, error) {
	var c ConnectionData
	c.Driver = "postgres"

	u, err := url.Parse(s)
	if err != nil {
		return c, err
	}
	c.Host = u.Hostname()
	c.Port = u.Port()
	c.Database = strings.TrimPrefix(u.Path, "/")
	c.Username = u.User.Username()

	pwd, _ := u.User.Password()
	c.Password = password.Password(pwd)
	c.Params = u.RawQuery
	return c, nil
}

func fromSqlServerDSN(s string) (ConnectionData, error) {
	var c ConnectionData
	c.Driver = "sqlserver"
	parts := strings.Split(s, ";")
	for _, p := range parts {
		keyVal := strings.Split(strings.TrimSpace(p), "=")
		if len(keyVal) != 2 {
			return c, errors.New("invalid connection string")
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
				return c, errors.New("invalid server")
			}
		case "database":
			c.Database = val
		case "user id":
			c.Username = val
		case "password":
			c.Password = password.Password(val)
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

	return c, nil
}

func (c ConnectionData) toDSN() (string, error) {
	if c.Driver == "postgres" {
		return c.toPostgresDSN()
	}

	return c.toSqlServerDSN()
}

func (c ConnectionData) toPostgresDSN() (string, error) {
	u := &url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(c.Username, c.Password.PlainText()),
		Host:     c.Host,
		Path:     "/" + c.Database,
		RawQuery: c.Params,
	}
	if c.Port != "" {
		u.Host = u.Host + ":" + c.Port
	}
	return u.String(), nil
}

func (c ConnectionData) toSqlServerDSN() (string, error) {
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
		parts = append(parts, fmt.Sprintf("password=%s", c.Password.PlainText()))
	}
	if c.Username == "" && c.Password == "" {
		parts = append(parts, "integrated security=SSPI")
	}
	if c.Params != "" {
		parts = append(parts, c.Params)
	}

	dsn := strings.Join(parts, "; ")
	return dsn, nil
}

func (c ConnectionData) DSN() string {
	res, err := c.toDSN()
	if err != nil {
		dlog.Error("failed to convert DSN: %v", err)
		return ""
	}
	return string(res)
}

func (c ConnectionData) String() string {
	pwd := "[unset]"
	if c.Password != "" {
		pwd = "redacted"
	}

	return fmt.Sprintf("{driver: %q, database: %q, username: %q, password: %q, params: %q}", c.Driver, c.Database, c.Username, pwd, c.Params)
}
