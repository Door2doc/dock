package db

import (
	"reflect"
	"testing"
)

func TestConnectionDataMarshal(t *testing.T) {
	for name, test := range map[string]struct {
		Given         string
		Want          *ConnectionData
		WantCanonical string
	}{
		"postgres URL": {
			Given: "postgres://pguser:pwd@localhost:5436/pgdb?sslmode=disable",
			Want: &ConnectionData{
				Driver:   "postgres",
				Host:     "localhost",
				Port:     "5436",
				Database: "pgdb",
				Username: "pguser",
				Password: "pwd",
				Params:   "sslmode=disable",
			},
			WantCanonical: "postgres://pguser:pwd@localhost:5436/pgdb?sslmode=disable",
		},
		"MSSQL ADO with instance": {
			Given: "server=localhost\\SQLExpress;user id=sa;database=master;app name=MyAppName",
			Want: &ConnectionData{
				Driver:   "sqlserver",
				Host:     "localhost",
				Instance: "SQLExpress",
				Database: "master",
				Username: "sa",
				Params:   "app name=MyAppName",
			},
			WantCanonical: "server=localhost\\SQLExpress; database=master; user id=sa; app name=MyAppName",
		},
		"MSSQL ADO with password": {
			Given: "Server=127.0.0.1; Database=myDB; User Id=username; Password=password ",
			Want: &ConnectionData{
				Driver:   "sqlserver",
				Host:     "127.0.0.1",
				Database: "myDB",
				Username: "username",
				Password: "password",
			},
			WantCanonical: "server=127.0.0.1; database=myDB; user id=username; password=password",
		},
	} {
		t.Run(name, func(t *testing.T) {
			c := new(ConnectionData)
			if err := c.UnmarshalText([]byte(test.Given)); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(c, test.Want) {
				t.Errorf("UnmarshalText() == \n\t%#v, got \n\t%#v", test.Want, c)
			}
			gotCanonical, err := c.MarshalText()
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(string(gotCanonical), test.WantCanonical) {
				t.Errorf("MarshalText() == \n\t%s, got \n\t%s", test.WantCanonical, gotCanonical)
			}
		})
	}
}
