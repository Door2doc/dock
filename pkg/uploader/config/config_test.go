package config

import (
	"context"
	"errors"
	"reflect"
	"testing"
)

func TestConfiguration_Validate(t *testing.T) {
	for name, test := range map[string]struct {
		Given func(cfg *Configuration)
		Want  *ValidationResult
	}{
		"unconfigured, no access": {
			Given: func(cfg *Configuration) {
				cfg.checkAccess = func() error {
					return errors.New("error")
				}
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
	} {
		t.Run(name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			cfg, err := Load(ctx)
			if err != nil {
				t.Fatal(err)
			}
			cfg.checkAccess = func() error {
				return nil
			}
			test.Given(cfg)
			got := cfg.Validate(ctx)

			if !reflect.DeepEqual(got, test.Want) {
				t.Errorf("Validate() == \n\t%#v, got \n\t%#v", test.Want, got)
			}
		})
	}
}
