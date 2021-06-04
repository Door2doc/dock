package config

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
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
				cfg.SetVisitorQuery(`select * from correct`)
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
				cfg.SetVisitorQuery(`select * from correct`)
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
				cfg.SetVisitorQuery(`select * from correct`)
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
				cfg.SetVisitorQuery(`select * from correct`)
				cfg.SetRadiologieQuery(`select * from correct_radiologie`)
				cfg.SetLabQuery(`select * from correct_lab`)
				cfg.SetConsultQuery(`select * from correct_consult`)
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
			cfg.UpdateValidation(ctx)
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

	for name, test := range map[string]*Configuration{
		"empty":    {connection: defaultConnection},
		"username": {username: "user", connection: defaultConnection},
		"password": {password: "pass", connection: defaultConnection},
		"dsn": {connection: db.ConnectionData{
			Driver:   "postgres",
			Host:     "localhost",
			Port:     "5436",
			Database: "pgdb",
			Username: "pguser",
			Password: "pass",
			Params:   "sslmode=disable",
		}},
		"query":         {visitorQuery: "query", connection: defaultConnection},
		"order queries": {radiologieQuery: "a", labQuery: "b", consultQuery: "c", connection: defaultConnection},
		"access":        {accessUsername: "username", accessPassword: "password", connection: db.ConnectionData{Driver: "sqlserver"}},
	} {
		t.Run(name, func(t *testing.T) {
			bs, err := json.Marshal(test)
			if err != nil {
				t.Fatal(err)
			}

			got := &Configuration{}
			if err := json.Unmarshal(bs, &got); err != nil {
				t.Fatal(err)
			}

			if !reflect.DeepEqual(got, test) {
				t.Errorf("Marshal/Unmarshal == %v, got %v", test, got)
			}
		})
	}
}
