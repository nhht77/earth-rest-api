package app

import (
	"bytes"
	"testing"

	"github.com/nhht77/earth-rest-api/pkg/logging"
	"github.com/sirupsen/logrus"
)

func newTestLogger() logging.Logger {
	l := logrus.New()
	buffer := bytes.NewBuffer([]byte{})
	l.SetOutput(buffer)
	return logging.NewFromLogrus(l)
}

func TestSetupParsesFlagsIntoConfig(t *testing.T) {

	args := []string{"-database-host", "localhost", "-database-port", "5432", "-test-build", "true", "-port", "8080"}
	app := New(newTestLogger())
	err := app.Initialize(args)
	if err != nil {
		t.Fatal("Expected no error, got", err)
	}

	if app.Conf.DataSource.DatabaseHost != "localhost" {
		t.Fatal("Expected database host to be localhost, got", app.Conf.DataSource.DatabaseHost)
	}
	if app.Conf.DataSource.DatabasePort != "5432" {
		t.Fatal("Expected database port to be 5432, got", app.Conf.DataSource.DatabasePort)
	}
	if app.Conf.Framework.IsTestBuild != true {
		t.Fatal("Expected test build to be true, got", app.Conf.Framework.IsTestBuild)
	}
	if app.Conf.Framework.ServerPort != "8080" {
		t.Fatal("Expected server port to be 8080, got", app.Conf.Framework.ServerPort)
	}
}

func TestInitFrameworkRejectsUnknownFlag(t *testing.T) {
	a := New(newTestLogger())

	err := a.Initialize([]string{"--definitely-not-a-flag", "1"})

	if err == nil {
		t.Fatal("expected error for unknown flag, got nil")
	}
}

func TestNewWiresDependencies(t *testing.T) {
	a := New(newTestLogger())

	if a.Conf == nil || a.DB == nil || a.DB.Log == nil {
		t.Fatal("New must wire Conf, DB and DB.Log — nothing may stay nil after construction")
	}
}
