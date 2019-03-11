package config

import (
	"context"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

const (
	TestUser     = "test-user"
	TestPassword = "test-password"
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
		Given func(cfg *Configuration)
		Want  *ValidationResult
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
				Server = srv.URL
			},
			Want: &ValidationResult{
				DatabaseConnection: ErrDatabaseNotConfigured,
				VisitorQuery:       ErrVisitorQueryNotConfigured,
				D2DCredentials:     ErrD2DCredentialsNotConfigured,
			},
		},
		"correct credentials": {
			Given: func(cfg *Configuration) {
				Server = srv.URL
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
				Server = srv.URL
				cfg.Username = "konijn"
				cfg.Password = TestPassword
			},
			Want: &ValidationResult{
				DatabaseConnection: ErrDatabaseNotConfigured,
				VisitorQuery:       ErrVisitorQueryNotConfigured,
				D2DCredentials:     ErrD2DCredentialsInvalid,
			},
		},
	} {
		t.Run(name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			cfg, err := Load(ctx)
			if err != nil {
				t.Fatal(err)
			}
			test.Given(cfg)
			got := cfg.Validate(ctx)

			if !reflect.DeepEqual(got, test.Want) {
				t.Errorf("Validate() == \n\t%v, got \n\t%v", test.Want, got)
			}
		})
	}
}
