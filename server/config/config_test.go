package config_test

import (
	"io"
	"testing"

	c "github.com/nhht77/earth-rest-api/server/config"
	"github.com/nhht77/earth-rest-api/server/pkg/logging"
	"github.com/sirupsen/logrus"
)

// testLogger discards output — config methods log, tests shouldn't be noisy.
func testLogger() logging.Logger {
	l := logrus.New()
	l.SetOutput(io.Discard)
	return logging.NewFromLogrus(l)
}

// from returns a Config usable as ReadDefault's parameter.
// (Note how we NEED Framework set — nil would panic. Keep that smell in mind.)
func from(isTestBuild bool) c.Config {
	return c.Config{Framework: &c.FrameworkConfig{IsTestBuild: isTestBuild}}
}

func TestReadDefaultFromEnv(t *testing.T) {
	t.Setenv("APP_DATABASE", "earth")
	t.Setenv("APP_USERNAME", "earth")
	t.Setenv("APP_PASSWORD", "secret")

	c := &c.Config{}
	c.ReadDefault(from(false), testLogger())

	if c.Database.Database != "earth" || c.Database.Username != "earth" || c.Database.Password != "secret" {
		t.Errorf("ReadDefault did not load env vars: %+v", c.Database)
	}
}

func TestReadDefaultTestBuildUsesTestEnv(t *testing.T) {
	t.Setenv("TEST_DATABASE", "earth_test")
	t.Setenv("TEST_USERNAME", "earth")
	t.Setenv("TEST_PASSWORD", "secret")

	c := &c.Config{}
	c.ReadDefault(from(true), testLogger())

	if c.Database.Database != "earth_test" {
		t.Errorf("test build should read TEST_DATABASE, got %q", c.Database.Database)
	}
}

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(c *c.Config)
		wantErr bool
	}{
		{"complete", func(conf *c.Config) { conf.Database = c.DatabaseConfig{"earth", "pw", "earth"} }, false},
		{"missing database", func(conf *c.Config) { conf.Database = c.DatabaseConfig{"", "pw", "earth"} }, true},
		{"missing username", func(conf *c.Config) { conf.Database = c.DatabaseConfig{"earth", "pw", ""} }, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &c.Config{}
			tt.mutate(c)
			err := c.ValidateConfig()
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// EXPOSES Bug A: Printable must redact its copy, not the original.
func TestPrintableRedactsButDoesNotMutate(t *testing.T) {
	cfg := &c.Config{}
	cfg.Database.Username = "earth"
	cfg.Database.Password = "supersecret"

	p := cfg.Printable()

	if p.Database.Password != "****" {
		t.Errorf("printable password = %q, want ****", p.Database.Password)
	}
	if cfg.Database.Password != "supersecret" {
		t.Errorf("BUG: Printable mutated the original password to %q", cfg.Database.Password)
	}
}
