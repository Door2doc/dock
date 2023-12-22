package config

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/door2doc/d2d-uploader/pkg/uploader/db"
	_ "github.com/lib/pq"
)

const (
	TestUser     = "test-user"
	TestPassword = "test-password"
)

var (
	TestConnection = db.ConnectionData{
		Driver:   "postgres",
		Host:     "localhost",
		Port:     "5436",
		Database: "pgdb",
		Username: "pguser",
		Password: "pwd",
		Params:   "sslmode=disable",
	}
)

func DummyHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username, password, ok := r.BasicAuth()
		if !ok || username != TestUser || password != TestPassword {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		if r.URL.Path == PathPing {
			w.WriteHeader(http.StatusOK)
			return
		}

		w.WriteHeader(http.StatusNotFound)
	})
}

func TestConfiguration_Validate(t *testing.T) {
	srv := httptest.NewServer(DummyHandler())
	defer srv.Close()

	failing := httptest.NewServer(nil)
	failing.Close()

	for name, test := range map[string]struct {
		Given            func(cfg *Configuration)
		Want             *ValidationResult
		WantValid        bool
		RequiresDatabase bool
	}{
		"unconfigured, no access": {
			Given: func(cfg *Configuration) {
				Server = failing.URL
			},
			Want: &ValidationResult{
				DatabaseConnection: ErrDatabaseNotConfigured,
				VisitorQuery:       ErrQueryNotConfigured,
				RadiologieQuery:    ErrQueryNotConfigured,
				LabQuery:           ErrQueryNotConfigured,
				ConsultQuery:       ErrQueryNotConfigured,
				D2DConnection:      ErrD2DConnectionFailed,
				D2DCredentials:     ErrD2DCredentialsNotConfigured,
				Access:             ErrAccessNotConfigured,
			},
		},
		"unconfigured with endpoint access": {
			Given: func(cfg *Configuration) {
			},
			Want: &ValidationResult{
				DatabaseConnection: ErrDatabaseNotConfigured,
				VisitorQuery:       ErrQueryNotConfigured,
				RadiologieQuery:    ErrQueryNotConfigured,
				LabQuery:           ErrQueryNotConfigured,
				ConsultQuery:       ErrQueryNotConfigured,
				D2DCredentials:     ErrD2DCredentialsNotConfigured,
				Access:             ErrAccessNotConfigured,
			},
		},
		"correct credentials": {
			Given: func(cfg *Configuration) {
				cfg.SetCredentials(TestUser, TestPassword)
			},
			Want: &ValidationResult{
				DatabaseConnection: ErrDatabaseNotConfigured,
				VisitorQuery:       ErrQueryNotConfigured,
				RadiologieQuery:    ErrQueryNotConfigured,
				LabQuery:           ErrQueryNotConfigured,
				ConsultQuery:       ErrQueryNotConfigured,
				Access:             ErrAccessNotConfigured,
			},
		},
		"incorrect credentials": {
			Given: func(cfg *Configuration) {
				cfg.SetCredentials("konijn", TestPassword)
			},
			Want: &ValidationResult{
				DatabaseConnection: ErrDatabaseNotConfigured,
				VisitorQuery:       ErrQueryNotConfigured,
				RadiologieQuery:    ErrQueryNotConfigured,
				LabQuery:           ErrQueryNotConfigured,
				ConsultQuery:       ErrQueryNotConfigured,
				D2DCredentials:     ErrD2DCredentialsInvalid,
				Access:             ErrAccessNotConfigured,
			},
		},
		"correct database": {
			Given: func(cfg *Configuration) {
				cfg.SetConnection(TestConnection)
			},
			Want: &ValidationResult{
				D2DCredentials:  ErrD2DCredentialsNotConfigured,
				VisitorQuery:    ErrQueryNotConfigured,
				RadiologieQuery: ErrQueryNotConfigured,
				LabQuery:        ErrQueryNotConfigured,
				ConsultQuery:    ErrQueryNotConfigured,
				Access:          ErrAccessNotConfigured,
			},
			RequiresDatabase: true,
		},
		"incorrect user": {
			Given: func(cfg *Configuration) {
				cfg.SetConnection(db.ConnectionData{
					Driver:   "postgres",
					Host:     "localhost",
					Port:     "5436",
					Instance: "",
					Database: "pgdb",
					Username: "postgres",
					Password: "pwd",
					Params:   "sslmode=disable",
				})
			},
			Want: &ValidationResult{
				DatabaseConnection: &DatabaseInvalidError{
					Cause: `pq: password authentication failed for user "postgres"`,
				},
				D2DCredentials:  ErrD2DCredentialsNotConfigured,
				VisitorQuery:    ErrQueryNotConfigured,
				RadiologieQuery: ErrQueryNotConfigured,
				LabQuery:        ErrQueryNotConfigured,
				ConsultQuery:    ErrQueryNotConfigured,
				Access:          ErrAccessNotConfigured,
			},
			RequiresDatabase: true,
		},
		"incorrect password": {
			Given: func(cfg *Configuration) {
				cfg.SetConnection(db.ConnectionData{
					Driver:   "postgres",
					Host:     "localhost",
					Port:     "5436",
					Instance: "",
					Database: "pgdb",
					Username: "pguser",
					Password: "password",
					Params:   "sslmode=disable",
				})
			},
			Want: &ValidationResult{
				DatabaseConnection: &DatabaseInvalidError{
					Cause: `pq: password authentication failed for user "pguser"`,
				},
				D2DCredentials:  ErrD2DCredentialsNotConfigured,
				VisitorQuery:    ErrQueryNotConfigured,
				RadiologieQuery: ErrQueryNotConfigured,
				LabQuery:        ErrQueryNotConfigured,
				ConsultQuery:    ErrQueryNotConfigured,
				Access:          ErrAccessNotConfigured,
			},
			RequiresDatabase: true,
		},
		"incorrect database": {
			Given: func(cfg *Configuration) {
				cfg.SetConnection(db.ConnectionData{
					Driver:   "postgres",
					Host:     "localhost",
					Port:     "5436",
					Instance: "",
					Database: "database",
					Username: "pguser",
					Password: "pwd",
					Params:   "sslmode=disable",
				})
			},
			Want: &ValidationResult{
				DatabaseConnection: &DatabaseInvalidError{
					Cause: `pq: database "database" does not exist`,
				},
				D2DCredentials:  ErrD2DCredentialsNotConfigured,
				VisitorQuery:    ErrQueryNotConfigured,
				RadiologieQuery: ErrQueryNotConfigured,
				LabQuery:        ErrQueryNotConfigured,
				ConsultQuery:    ErrQueryNotConfigured,
				Access:          ErrAccessNotConfigured,
			},
			RequiresDatabase: true,
		},
		//"incorrect host": {	//  exact error string varies :-/
		//	Given: func(cfg *Configuration) {
		//		cfg.SetDSN("postgres", "postgres://pguser:pwd@localhost:9999/pgdb?sslmode=disable")
		//	},
		//	Want: &ValidationResult{
		//		DatabaseConnection: &DatabaseInvalidError{
		//			Cause: `dial tcp [::1]:9999: connect: connection refused`,
		//		},
		//		D2DCredentials:  ErrD2DCredentialsNotConfigured,
		//		VisitorQuery:    ErrQueryNotConfigured,
		//		RadiologieQuery: ErrQueryNotConfigured,
		//		LabQuery:        ErrQueryNotConfigured,
		//		ConsultQuery:    ErrQueryNotConfigured,
		//		Access:          ErrAccessNotConfigured,
		//	},
		//},
		"correct query": {
			Given: func(cfg *Configuration) {
				cfg.SetConnection(TestConnection)
				cfg.SetVisitorQuery(`SELECT * FROM correct`)
			},
			Want: &ValidationResult{
				D2DCredentials:  ErrD2DCredentialsNotConfigured,
				RadiologieQuery: ErrQueryNotConfigured,
				LabQuery:        ErrQueryNotConfigured,
				ConsultQuery:    ErrQueryNotConfigured,
				Access:          ErrAccessNotConfigured,
			},
			RequiresDatabase: true,
		},
		"correct minimal configuration orders": {
			Given: func(cfg *Configuration) {
				cfg.SetConnection(TestConnection)
				cfg.SetCredentials(TestUser, TestPassword)
				cfg.SetVisitorQuery(`SELECT * FROM correct`)
			},
			Want: &ValidationResult{
				RadiologieQuery: ErrQueryNotConfigured,
				LabQuery:        ErrQueryNotConfigured,
				ConsultQuery:    ErrQueryNotConfigured,
				Access:          ErrAccessNotConfigured,
			},
			WantValid:        true,
			RequiresDatabase: true,
		},
		"correct configuration without orders": {
			Given: func(cfg *Configuration) {
				cfg.SetConnection(TestConnection)
				cfg.SetCredentials(TestUser, TestPassword)
				cfg.SetVisitorQuery(`SELECT * FROM correct`)
				cfg.SetAccessCredentials(TestUser, TestPassword)
			},
			Want: &ValidationResult{
				RadiologieQuery: ErrQueryNotConfigured,
				LabQuery:        ErrQueryNotConfigured,
				ConsultQuery:    ErrQueryNotConfigured,
			},
			WantValid:        true,
			RequiresDatabase: true,
		},
		"correct configuration with orders": {
			Given: func(cfg *Configuration) {
				cfg.SetConnection(TestConnection)
				cfg.SetCredentials(TestUser, TestPassword)
				cfg.SetVisitorQuery(`SELECT * FROM correct`)
				cfg.SetRadiologieQuery(`SELECT * FROM correct_radiologie`)
				cfg.SetLabQuery(`SELECT * FROM correct_lab`)
				cfg.SetConsultQuery(`SELECT * FROM correct_consult`)
				cfg.SetAccessCredentials(TestUser, TestPassword)
			},
			Want:             &ValidationResult{},
			WantValid:        true,
			RequiresDatabase: true,
		},
	} {
		t.Run(name, func(t *testing.T) {
			if test.RequiresDatabase && testing.Short() {
				t.Skip("test requires functioning database")
			}

			ctx, timeout := context.WithTimeout(context.Background(), time.Second)
			defer timeout()

			Server = srv.URL
			cfg := NewConfiguration()

			test.Given(cfg)
			cfg.UpdateBaseValidation(ctx)
			cfg.UpdateRadiologieValidation(ctx)
			cfg.UpdateLabValidation(ctx)
			cfg.UpdateConsultValidation(ctx)

			got := cfg.Validate()

			// reset this stuff
			got.VisitorQueryDuration = 0
			got.VisitorQueryResults = nil
			got.RadiologieQueryDuration = 0
			got.RadiologieQueryResults = nil
			got.LabQueryDuration = 0
			got.LabQueryResults = nil
			got.ConsultQueryDuration = 0
			got.ConsultQueryResults = nil

			if !reflect.DeepEqual(got, test.Want) {
				t.Errorf("UpdateValidation() == \n\t%v, got \n\t%v", test.Want, got)
			}
			if !reflect.DeepEqual(got.IsValid(), test.WantValid) {
				t.Errorf("UpdateValidation().IsValid() == %t, got %t", test.WantValid, got.IsValid())
			}
		})
	}
}

func TestConfigurationMarshal(t *testing.T) {
	for file, want := range map[string]DataV2{
		"testdata/config.v1.json": {
			Version:  2,
			Username: "upload-user",
			Password: "upload-password",
			Connection: db.ConnectionData{
				Driver:   "sqlserver",
				Host:     "host",
				Port:     "",
				Instance: "instance",
				Database: "database",
				Username: "db-username",
				Password: "db-password",
				Params:   "p=a",
			},
			Timeout:         40 * time.Second,
			VisitorQuery:    "visitor",
			RadiologieQuery: "radio",
			LabQuery:        "lab",
			ConsultQuery:    "consult",
			Proxy:           "proxy",
			AccessUsername:  "web-user",
			AccessPassword:  "web-password",
		},
		"testdata/config.v2.json": {
			Version:  2,
			Username: "upload-user",
			Password: "upload-password",
			Connection: db.ConnectionData{
				Driver:   "sqlserver",
				Host:     "host",
				Port:     "",
				Instance: "instance",
				Database: "database",
				Username: "db-username",
				Password: "MyPassw0rd",
				Params:   "p=a",
			},
			Timeout:         40 * time.Second,
			VisitorQuery:    "visitor",
			RadiologieQuery: "radio",
			LabQuery:        "lab",
			ConsultQuery:    "consult",
			Proxy:           "proxy",
			AccessUsername:  "web-user",
			AccessPassword:  "web-password",
		},
	} {
		t.Run(file, func(t *testing.T) {
			f, err := os.Open(file)
			if err != nil {
				t.Fatal(err)
			}
			defer func() { _ = f.Close() }()

			data, err := io.ReadAll(f)
			if err != nil {
				t.Fatal(err)
			}
			var got = new(Configuration)
			if err := json.Unmarshal(data, &got); err != nil {
				t.Fatalf("%v", err)
			}

			if !reflect.DeepEqual(got.data, want) {
				t.Errorf("Got == %#v, want %#v", got.data, want)
			}
		})
	}
}
