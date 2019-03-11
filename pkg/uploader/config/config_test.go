package config

import (
	"context"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

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
		Given     func(cfg *Configuration)
		Want      *ValidationResult
		WantValid bool
	}{
		"unconfigured, no access": {
			Given: func(cfg *Configuration) {
				Server = failing.URL
			},
			Want: &ValidationResult{
				DatabaseConnection: ErrDatabaseNotConfigured,
				VisitorQuery:       ErrVisitorQueryNotConfigured,
				D2DConnection:      ErrD2DConnectionFailed,
				D2DCredentials:     ErrD2DCredentialsNotConfigured,
			},
		},
		"unconfigured with endpoint access": {
			Given: func(cfg *Configuration) {
			},
			Want: &ValidationResult{
				DatabaseConnection: ErrDatabaseNotConfigured,
				VisitorQuery:       ErrVisitorQueryNotConfigured,
				D2DCredentials:     ErrD2DCredentialsNotConfigured,
			},
		},
		"correct credentials": {
			Given: func(cfg *Configuration) {
				cfg.Username = TestUser
				cfg.Password = TestPassword
			},
			Want: &ValidationResult{
				DatabaseConnection: ErrDatabaseNotConfigured,
				VisitorQuery:       ErrVisitorQueryNotConfigured,
			},
		},
		"incorrect credentials": {
			Given: func(cfg *Configuration) {
				cfg.Username = "konijn"
				cfg.Password = TestPassword
			},
			Want: &ValidationResult{
				DatabaseConnection: ErrDatabaseNotConfigured,
				VisitorQuery:       ErrVisitorQueryNotConfigured,
				D2DCredentials:     ErrD2DCredentialsInvalid,
			},
		},
		"correct database": {
			Given: func(cfg *Configuration) {
				cfg.DSN = TestDSN
			},
			Want: &ValidationResult{
				D2DCredentials: ErrD2DCredentialsNotConfigured,
				VisitorQuery:   ErrVisitorQueryNotConfigured,
			},
		},
		"incorrect user": {
			Given: func(cfg *Configuration) {
				cfg.DSN = "postgres://postgres:pwd@localhost:5436/pgdb?sslmode=disable"
			},
			Want: &ValidationResult{
				DatabaseConnection: &ErrDatabaseInvalid{
					Cause: `pq: password authentication failed for user "postgres"`,
				},
				D2DCredentials: ErrD2DCredentialsNotConfigured,
				VisitorQuery:   ErrVisitorQueryNotConfigured,
			},
		},
		"incorrect password": {
			Given: func(cfg *Configuration) {
				cfg.DSN = "postgres://pguser:password@localhost:5436/pgdb?sslmode=disable"
			},
			Want: &ValidationResult{
				DatabaseConnection: &ErrDatabaseInvalid{
					Cause: `pq: password authentication failed for user "pguser"`,
				},
				D2DCredentials: ErrD2DCredentialsNotConfigured,
				VisitorQuery:   ErrVisitorQueryNotConfigured,
			},
		},
		"incorrect database": {
			Given: func(cfg *Configuration) {
				cfg.DSN = "postgres://pguser:pwd@localhost:5436/database?sslmode=disable"
			},
			Want: &ValidationResult{
				DatabaseConnection: &ErrDatabaseInvalid{
					Cause: `pq: database "database" does not exist`,
				},
				D2DCredentials: ErrD2DCredentialsNotConfigured,
				VisitorQuery:   ErrVisitorQueryNotConfigured,
			},
		},
		"incorrect host": {
			Given: func(cfg *Configuration) {
				cfg.DSN = "postgres://pguser:pwd@localhost:9999/pgdb?sslmode=disable"
			},
			Want: &ValidationResult{
				DatabaseConnection: &ErrDatabaseInvalid{
					Cause: `dial tcp [::1]:9999: connect: connection refused`,
				},
				D2DCredentials: ErrD2DCredentialsNotConfigured,
				VisitorQuery:   ErrVisitorQueryNotConfigured,
			},
		},
		"correct query": {
			Given: func(cfg *Configuration) {
				cfg.DSN = TestDSN
				cfg.Query = `select * from correct`
			},
			Want: &ValidationResult{
				D2DCredentials: ErrD2DCredentialsNotConfigured,
			},
		},
		"correct configuration": {
			Given: func(cfg *Configuration) {
				cfg.DSN = TestDSN
				cfg.Username = TestUser
				cfg.Password = TestPassword
				cfg.Query = `select * from correct`
			},
			Want:      &ValidationResult{},
			WantValid: true,
		},
	} {
		t.Run(name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			Server = srv.URL
			cfg := &Configuration{
				Driver: "postgres",
			}

			test.Given(cfg)
			got := cfg.Validate(ctx)

			if !reflect.DeepEqual(got, test.Want) {
				t.Errorf("Validate() == \n\t%v, got \n\t%v", test.Want, got)
			}
			if !reflect.DeepEqual(got.IsValid(), test.WantValid) {
				t.Errorf("Validate().IsValid() == %t, got %t", test.WantValid, got.IsValid())
			}
		})
	}
}
