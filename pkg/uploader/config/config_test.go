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
	TestDSN      = "postgres://pguser:pwd@localhost:5436/pgdb?sslmode=disable"
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
				cfg.SetDSN("postgres", TestDSN)
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
				cfg.SetDSN("postgres", "postgres://postgres:pwd@localhost:5436/pgdb?sslmode=disable")
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
				cfg.SetDSN("postgres", "postgres://pguser:password@localhost:5436/pgdb?sslmode=disable")
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
				cfg.SetDSN("postgres", "postgres://pguser:pwd@localhost:5436/database?sslmode=disable")
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
				cfg.SetDSN("postgres", TestDSN)
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
				cfg.SetDSN("postgres", TestDSN)
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
				cfg.SetDSN("postgres", TestDSN)
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
				cfg.SetDSN("postgres", TestDSN)
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

func TestConfigurationJSON(t *testing.T) {
	defaultConnection := db.ConnectionData{Driver: "sqlserver"}
	defaultTimeout := NewConfiguration().timeout

	for name, test := range map[string]*Configuration{
		"empty":    {},
		"username": {username: "user"},
		"password": {password: "pass"},
		"dsn": {connection: db.ConnectionData{
			Driver:   "postgres",
			Host:     "localhost",
			Port:     "5436",
			Database: "pgdb",
			Username: "pguser",
			Password: "pass",
			Params:   "sslmode=disable",
		}},
		"query":         {visitorQuery: "query"},
		"order queries": {radiologieQuery: "a", labQuery: "b", consultQuery: "c"},
		"access":        {accessUsername: "username", accessPassword: "password", connection: db.ConnectionData{Driver: "sqlserver"}},
		"timeout":       {timeout: 100 * time.Second},
	} {
		t.Run(name, func(t *testing.T) {
			if test.connection == (db.ConnectionData{}) {
				test.connection = defaultConnection
			}
			if test.timeout == 0 {
				test.timeout = defaultTimeout
			}

			bs, err := json.Marshal(test)
			if err != nil {
				t.Fatal(err)
			}

			got := &Configuration{}
			if err := json.Unmarshal(bs, &got); err != nil {
				t.Fatal(err)
			}

			if !reflect.DeepEqual(got, test) {
				t.Errorf("Marshal/Unmarshal == \n\t%v, want \n\t%v", got, test)
			}
		})
	}
}

func TestConfigurationMarshal(t *testing.T) {
	for file, want := range map[string]*Configuration{
		"testdata/config.v1.json": {
			username: "upload-user",
			password: "upload-password",
			connection: db.ConnectionData{
				Driver:   "sqlserver",
				Host:     "host",
				Port:     "",
				Instance: "instance",
				Database: "database",
				Username: "db-username",
				Password: "db-password",
				Params:   "p=a",
			},
			timeout:         40 * time.Second,
			visitorQuery:    "visitor",
			radiologieQuery: "radio",
			labQuery:        "lab",
			consultQuery:    "consult",
			proxy:           "proxy",
			accessUsername:  "web-user",
			accessPassword:  "web-password",
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
				t.Fatal(err)
			}

			if got.username != want.username {
				t.Errorf("Username == %q, want %q", got.username, want.username)
			}
			if got.password != want.password {
				t.Errorf("Password == %v, want %v", got.password, want.password)
			}
			if !reflect.DeepEqual(got.connection, want.connection) {
				t.Errorf("Connection == %v, want %v", got.connection, want.connection)
			}
			if got.timeout != want.timeout {
				t.Errorf("Timeout == %s, want %s", got.timeout, want.timeout)
			}
			if got.visitorQuery != want.visitorQuery {
				t.Errorf("Visitor query == %q, want %q", got.visitorQuery, want.visitorQuery)
			}
			if got.radiologieQuery != want.radiologieQuery {
				t.Errorf("Radiologie query == %q, want %q", got.radiologieQuery, want.radiologieQuery)
			}
			if got.labQuery != want.labQuery {
				t.Errorf("Lab query == %q, want %q", got.labQuery, want.labQuery)
			}
			if got.consultQuery != want.consultQuery {
				t.Errorf("Consult query == %q, want %q", got.consultQuery, want.consultQuery)
			}
			if got.active != want.active {
				t.Errorf("Active == %v, want %v", got.active, want.active)
			}
			if got.interval != want.interval {
				t.Errorf("Interval == %s, want %s", got.interval, want.interval)
			}
			if got.proxy != want.proxy {
				t.Errorf("Proxy == %q, want %q", got.proxy, want.proxy)
			}
			if got.accessUsername != want.accessUsername {
				t.Errorf("Access username == %q, want %q", got.accessUsername, want.accessUsername)
			}
			if got.accessPassword != want.accessPassword {
				t.Errorf("Access == %v, want %v", got.accessPassword, want.accessPassword)
			}

		})
	}
}
