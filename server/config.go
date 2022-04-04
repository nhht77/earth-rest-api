package main

import (
	"errors"
	"fmt"
	"os"
)

type Config struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Database string `json:"database"`
}

// Read and print Database connection

func (c *Config) ReadDefault() {
	if len(c.Database) == 0 {
		c.Database = os.Getenv("APP_DATABASE")
	}
	if len(c.Username) == 0 {
		c.Username = os.Getenv("APP_USERNAME")
	}
	if len(c.Password) == 0 {
		c.Password = os.Getenv("APP_PASSWORD")
	}
}

func (c *Config) ReadTestDefault() {
	if len(c.Database) == 0 {
		c.Database = os.Getenv("TEST_DATABASE")
	}
	if len(c.Username) == 0 {
		c.Username = os.Getenv("TEST_USERNAME")
	}
	if len(c.Password) == 0 {
		c.Password = os.Getenv("TEST_PASSWORD")
	}
}

func (c *Config) ValidateConfig() error {
	if len(c.Database) == 0 {
		return errors.New("Config.Database")
	}
	if len(c.Username) == 0 {
		return errors.New("Config.Username")
	}
	if len(c.Password) == 0 {
		return errors.New("Config.Password")
	}
	return nil
}

func (c *Config) DatabaseSource(host string, port string) string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host,
		port,
		c.Username,
		c.Password,
		c.Database,
	)
}

func (c *Config) DatabaseSourcePrintable(host string, port string) string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host,
		port,
		c.Username[0:3],
		"****",
		c.Database[0:3],
	)
}

func (c *Config) Printable() *Config {

	temp := *c

	if l := len(c.Database); l > 0 {
		temp.Database = c.Database[0:int(l/3)] + "..."
	}

	if l := len(c.Username); l > 0 {
		temp.Username = c.Username[0:int(l/3)] + "..."
	}

	if len(c.Password) > 0 {
		c.Password = "****"
	}

	return &temp
}
