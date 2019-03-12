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
				cfg.SetCredentials(TestUser, TestPassword)
			},
			Want: &ValidationResult{
				DatabaseConnection: ErrDatabaseNotConfigured,
				VisitorQuery:       ErrVisitorQueryNotConfigured,
			},
		},
		"incorrect credentials": {
			Given: func(cfg *Configuration) {
				cfg.SetCredentials("konijn", TestPassword)
			},
			Want: &ValidationResult{
				DatabaseConnection: ErrDatabaseNotConfigured,
				VisitorQuery:       ErrVisitorQueryNotConfigured,
				D2DCredentials:     ErrD2DCredentialsInvalid,
			},
		},
		"correct database": {
			Given: func(cfg *Configuration) {
				cfg.SetDSN("postgres", TestDSN)
			},
			Want: &ValidationResult{
				D2DCredentials: ErrD2DCredentialsNotConfigured,
				VisitorQuery:   ErrVisitorQueryNotConfigured,
			},
		},
		"incorrect user": {
			Given: func(cfg *Configuration) {
				cfg.SetDSN("postgres", "postgres://postgres:pwd@localhost:5436/pgdb?sslmode=disable")
			},
			Want: &ValidationResult{
				DatabaseConnection: &DatabaseInvalidError{
					Cause: `pq: password authentication failed for user "postgres"`,
				},
				D2DCredentials: ErrD2DCredentialsNotConfigured,
				VisitorQuery:   ErrVisitorQueryNotConfigured,
			},
		},
		"incorrect password": {
			Given: func(cfg *Configuration) {
				cfg.SetDSN("postgres", "postgres://pguser:password@localhost:5436/pgdb?sslmode=disable")
			},
			Want: &ValidationResult{
				DatabaseConnection: &DatabaseInvalidError{
					Cause: `pq: password authentication failed for user "pguser"`,
				},
				D2DCredentials: ErrD2DCredentialsNotConfigured,
				VisitorQuery:   ErrVisitorQueryNotConfigured,
			},
		},
		"incorrect database": {
			Given: func(cfg *Configuration) {
				cfg.SetDSN("postgres", "postgres://pguser:pwd@localhost:5436/database?sslmode=disable")
			},
			Want: &ValidationResult{
				DatabaseConnection: &DatabaseInvalidError{
					Cause: `pq: database "database" does not exist`,
				},
				D2DCredentials: ErrD2DCredentialsNotConfigured,
				VisitorQuery:   ErrVisitorQueryNotConfigured,
			},
		},
		"incorrect host": {
			Given: func(cfg *Configuration) {
				cfg.SetDSN("postgres", "postgres://pguser:pwd@localhost:9999/pgdb?sslmode=disable")
			},
			Want: &ValidationResult{
				DatabaseConnection: &DatabaseInvalidError{
					Cause: `dial tcp [::1]:9999: connect: connection refused`,
				},
				D2DCredentials: ErrD2DCredentialsNotConfigured,
				VisitorQuery:   ErrVisitorQueryNotConfigured,
			},
		},
		"correct query": {
			Given: func(cfg *Configuration) {
				cfg.SetDSN("postgres", TestDSN)
				cfg.SetQuery(`select * from correct`)
			},
			Want: &ValidationResult{
				D2DCredentials: ErrD2DCredentialsNotConfigured,
			},
		},
		"correct configuration": {
			Given: func(cfg *Configuration) {
				cfg.SetDSN("postgres", TestDSN)
				cfg.SetCredentials(TestUser, TestPassword)
				cfg.SetQuery(`select * from correct`)
			},
			Want:      &ValidationResult{},
			WantValid: true,
		},
	} {
		t.Run(name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			Server = srv.URL
			cfg := NewConfiguration()

			test.Given(cfg)
			cfg.UpdateValidation(ctx)
			got := cfg.Validate()

			if !reflect.DeepEqual(got, test.Want) {
				t.Errorf("UpdateValidation() == \n\t%v, got \n\t%v", test.Want, got)
			}
			if !reflect.DeepEqual(got.IsValid(), test.WantValid) {
				t.Errorf("UpdateValidation().IsValid() == %t, got %t", test.WantValid, got.IsValid())
			}
		})
	}
}
