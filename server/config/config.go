package config

import (
	"errors"
	"fmt"
	"os"

	"github.com/nhht77/earth-rest-api/server/pkg/logging"
)

type Config struct {
	Credentials Credential       `json:"credentials"` // was: Database DatabaseConfig
	DataSource  *DataSource      `json:"data_source"` // where to connect — nil means "not initialized yet"
	Framework   *FrameworkConfig `json:"framework"`   // server/runtime settings
}

// Credential is the identity for the database connection.
type Credential struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Database string `json:"database"`
}

// DataSource is where to connect.
type DataSource struct {
	DatabaseHost string `json:"database_host"` // default "localhost"
	DatabasePort string `json:"database_port"` // default "5432"
}

type FrameworkConfig struct {
	ServerPort  string `json:"server_port"`   // default "8080"
	IsTestBuild bool   `json:"is_test_build"` // default false
}

// Read and print Database connection

func (c *Config) ReadDefault(appConf Config, log logging.Logger) {

	// @bugfix source test db if test build flag is true
	if appConf.Framework.IsTestBuild || false {
		appConf.Framework.IsTestBuild = true
		log.Infof("[postgre] Reading from test database %t", appConf.Framework.IsTestBuild)
		c.ReadTestDefault()
		return
	}

	if len(c.Credentials.Database) == 0 {
		c.Credentials.Database = os.Getenv("APP_DATABASE")
	}
	if len(c.Credentials.Username) == 0 {
		c.Credentials.Username = os.Getenv("APP_USERNAME")
	}
	if len(c.Credentials.Password) == 0 {
		c.Credentials.Password = os.Getenv("APP_PASSWORD")
	}
}

func (c *Config) ReadTestDefault() {
	if len(c.Credentials.Database) == 0 {
		c.Credentials.Database = os.Getenv("TEST_DATABASE")
	}
	if len(c.Credentials.Username) == 0 {
		c.Credentials.Username = os.Getenv("TEST_USERNAME")
	}
	if len(c.Credentials.Password) == 0 {
		c.Credentials.Password = os.Getenv("TEST_PASSWORD")
	}
}

func (c *Config) ValidateDataSource() error {
	if c.DataSource == nil {
		return errors.New("Config.DataSource is nil")
	}
	if len(c.DataSource.DatabaseHost) == 0 {
		return errors.New("Config.DataSource.DatabaseHost")
	}
	if len(c.DataSource.DatabasePort) == 0 {
		return errors.New("Config.DataSource.DatabasePort")
	}
	return nil
}

func (c *Config) ValidateConfig() error {
	if len(c.Credentials.Database) == 0 {
		return errors.New("Config.Credentials.Database")
	}
	if len(c.Credentials.Username) == 0 {
		return errors.New("Config.Credentials.Username")
	}
	if len(c.Credentials.Password) == 0 {
		return errors.New("Config.Credentials.Password")
	}
	return nil
}

func (c *Config) DatabaseSource() string {
	if c.DataSource == nil {
		return ""
	}
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		c.DataSource.DatabaseHost,
		c.DataSource.DatabasePort,
		c.Credentials.Username,
		c.Credentials.Password,
		c.Credentials.Database,
	)
}

func (c *Config) DatabaseSourcePrintable() string {
	if c.DataSource == nil {
		return ""
	}
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		c.DataSource.DatabaseHost,
		c.DataSource.DatabasePort,
		c.Credentials.Username[0:3]+"...",
		"****",
		c.Credentials.Database[0:3]+"...",
	)
}

func (c *Config) Printable() *Config {

	temp := *c

	if l := len(c.Credentials.Database); l > 0 {
		temp.Credentials.Database = c.Credentials.Database[0:int(l/3)] + "..."
	}

	if l := len(c.Credentials.Username); l > 0 {
		temp.Credentials.Username = c.Credentials.Username[0:int(l/3)] + "..."
	}

	if len(c.Credentials.Password) > 0 {
		temp.Credentials.Password = "****"
	}

	return &temp
}
