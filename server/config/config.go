package config

import (
	"errors"
	"fmt"
	"os"

	"github.com/nhht77/earth-rest-api/server/pkg/logging"
)

type Config struct {
	Database  DatabaseConfig   `json:"database"`
	Framework *FrameworkConfig `json:"framework"`
}

type DatabaseConfig struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Database string `json:"database"`
}

type FrameworkConfig struct {
	ServerPort  string `json:"server_port"`   // default "8080"
	IsTestBuild bool   `json:"is_test_build"` // default false

	DatabaseHost string `json:"database_host"` // default "localhost"
	DatabasePort string `json:"database_port"` // default "5432"
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

	if len(c.Database.Database) == 0 {
		c.Database.Database = os.Getenv("APP_DATABASE")
	}
	if len(c.Database.Username) == 0 {
		c.Database.Username = os.Getenv("APP_USERNAME")
	}
	if len(c.Database.Password) == 0 {
		c.Database.Password = os.Getenv("APP_PASSWORD")
	}
}

func (c *Config) ReadTestDefault() {
	if len(c.Database.Database) == 0 {
		c.Database.Database = os.Getenv("TEST_DATABASE")
	}
	if len(c.Database.Username) == 0 {
		c.Database.Username = os.Getenv("TEST_USERNAME")
	}
	if len(c.Database.Password) == 0 {
		c.Database.Password = os.Getenv("TEST_PASSWORD")
	}
}

func (c *Config) ValidateConfig() error {
	if len(c.Database.Database) == 0 {
		return errors.New("Config.Database")
	}
	if len(c.Database.Username) == 0 {
		return errors.New("Config.Username")
	}
	if len(c.Database.Password) == 0 {
		return errors.New("Config.Database.Password")
	}
	return nil
}

func (c *Config) DatabaseSource(host string, port string) string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host,
		port,
		c.Database.Username,
		c.Database.Password,
		c.Database.Database,
	)
}

func (c *Config) DatabaseSourcePrintable(host string, port string) string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host,
		port,
		c.Database.Username[0:3]+"...",
		"****",
		c.Database.Database[0:3]+"...",
	)
}

func (c *Config) Printable() *Config {

	temp := *c

	if l := len(c.Database.Database); l > 0 {
		temp.Database.Database = c.Database.Database[0:int(l/3)] + "..."
	}

	if l := len(c.Database.Username); l > 0 {
		temp.Database.Username = c.Database.Username[0:int(l/3)] + "..."
	}

	if len(c.Database.Password) > 0 {
		c.Database.Password = "****"
	}

	return &temp
}
