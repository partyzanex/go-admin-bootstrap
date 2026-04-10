package goadmin

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestConfig_Validate(t *testing.T) {
	validConfig := func() *Config {
		return &Config{
			Port:      8080,
			JWTSecret: []byte("this-secret-is-exactly-32-bytes!"),
			UserCase:  &stubUseCase{},
		}
	}

	tests := []struct {
		name    string
		modify  func(*Config)
		wantErr error
	}{
		{"valid", func(_ *Config) {}, nil},
		{"nil config", nil, ErrRequiredConfig},
		{"no port", func(c *Config) { c.Port = 0 }, ErrInvalidPort},
		{"no jwt secret", func(c *Config) { c.JWTSecret = nil }, ErrRequiredJWTSecret},
		{"jwt secret too short", func(c *Config) { c.JWTSecret = []byte("short") }, ErrJWTSecretTooShort},
		{"jwt secret too short allowed in dev mode", func(c *Config) {
			c.JWTSecret = []byte("short")
			c.DevMode = true
		}, nil},
		{"no usercase", func(c *Config) { c.UserCase = nil }, ErrRequiredUserCase},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.modify == nil {
				err := (*Config)(nil).Validate()
				assert.ErrorIs(t, err, tt.wantErr)

				return
			}

			cfg := validConfig()
			tt.modify(cfg)

			err := cfg.Validate()

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestConfig_Clone(t *testing.T) {
	cfg := &Config{
		Host: "localhost",
		Port: 9900,
	}

	clone := cfg.Clone()

	assert.Equal(t, cfg.Host, clone.Host)
	assert.Equal(t, cfg.Port, clone.Port)

	clone.Host = "changed"
	assert.NotEqual(t, cfg.Host, clone.Host)
}

// stubUseCase is a minimal implementation for config validation tests.
type stubUseCase struct{}

func (s *stubUseCase) Validate(_ *User, _ bool) error                           { return nil }
func (s *stubUseCase) SearchByLogin(_ context.Context, _ string) (*User, error) { return nil, nil }
func (s *stubUseCase) SearchByID(_ context.Context, _ int64) (*User, error)     { return nil, nil }
func (s *stubUseCase) SetLastLogged(_ context.Context, _ *User) error           { return nil }
func (s *stubUseCase) Register(_ context.Context, _ *User) error                { return nil }
func (s *stubUseCase) UpdateUser(_ context.Context, _ *User) (*User, error)     { return nil, nil }
func (s *stubUseCase) DeleteUser(_ context.Context, _ int64) error              { return nil }
func (s *stubUseCase) ListUsers(_ context.Context, _ *UserFilter) ([]*User, int64, error) {
	return nil, 0, nil
}
func (s *stubUseCase) ComparePassword(_ *User, _ string) (bool, error) { return false, nil }
func (s *stubUseCase) EncodePassword(_ *User) error                    { return nil }
func (s *stubUseCase) CreateAuthToken(_ context.Context, _ *User, _ string, _ time.Duration) (*Token, error) {
	return nil, nil
}
func (s *stubUseCase) SearchToken(_ context.Context, _ string) (*Token, error) { return nil, nil }
func (s *stubUseCase) RevokeUserTokens(_ context.Context, _ int64) error       { return nil }
func (s *stubUseCase) CleanupExpiredTokens(_ context.Context) (int64, error)   { return 0, nil }
