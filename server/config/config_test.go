package config

import (
	"io"
	"testing"

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
func from(isTestBuild bool) Config {
	return Config{Framework: &FrameworkConfig{IsTestBuild: isTestBuild}}
}

func TestReadDefaultFromEnv(t *testing.T) {
	t.Setenv("APP_DATABASE", "earth")
	t.Setenv("APP_USERNAME", "earth")
	t.Setenv("APP_PASSWORD", "secret")

	c := &Config{}
	c.ReadDefault(from(false), testLogger())

	if c.Credentials.Database != "earth" || c.Credentials.Username != "earth" || c.Credentials.Password != "secret" {
		t.Errorf("ReadDefault did not load env vars: %+v", c.Credentials)
	}
}

func TestReadDefaultTestBuildUsesTestEnv(t *testing.T) {
	t.Setenv("TEST_DATABASE", "earth_test")
	t.Setenv("TEST_USERNAME", "earth")
	t.Setenv("TEST_PASSWORD", "secret")

	c := &Config{}
	c.ReadDefault(from(true), testLogger())

	if c.Credentials.Database != "earth_test" {
		t.Errorf("test build should read TEST_DATABASE, got %q", c.Credentials.Database)
	}
}

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(c *Config)
		wantErr bool
	}{
		{"complete", func(conf *Config) { conf.Credentials = Credential{"earth", "pw", "earth"} }, false},
		{"missing database", func(conf *Config) { conf.Credentials = Credential{"", "pw", "earth"} }, true},
		{"missing username", func(conf *Config) { conf.Credentials = Credential{"earth", "pw", ""} }, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Config{}
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
	cfg := &Config{}
	cfg.Credentials.Username = "earth"
	cfg.Credentials.Password = "supersecret"

	p := cfg.Printable()

	if p.Credentials.Password != "****" {
		t.Errorf("printable password = %q, want ****", p.Credentials.Password)
	}
	if cfg.Credentials.Password != "supersecret" {
		t.Errorf("BUG: Printable mutated the original password to %q", cfg.Credentials.Password)
	}
}
