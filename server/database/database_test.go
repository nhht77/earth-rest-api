package database

import (
	"io"
	"testing"

	_ "github.com/lib/pq"
	"github.com/nhht77/earth-rest-api/server/config"
	"github.com/nhht77/earth-rest-api/server/pkg/logging"
	"github.com/sirupsen/logrus"
)

// type Database struct {
// 	postgres *sql.DB
// 	Log      logging.Logger
// }

func testLogger() logging.Logger {
	l := logrus.New()
	l.SetOutput(io.Discard)
	return logging.NewFromLogrus(l)
}

// Regression for the nil-interface lesson: NewDatabase must wire the logger.
func TestNewDatabaseWiresLogger(t *testing.T) {
	db := NewDatabase(testLogger())
	if db.Log == nil {
		t.Fatal("NewDatabase must set Log — a nil Log field panics on first use")
	}
}

func TestInitializeWiresLogger(t *testing.T) {
	db := NewDatabase(testLogger())
	err := db.Initialize(&config.Config{})
	if err == nil {
		t.Fatal("Expect validation error")
	}
}

func TestInitializeValidConfigNilDBSource(t *testing.T) {
	db := NewDatabase(testLogger())
	err := db.Initialize(&config.Config{
		Credentials: config.Credential{
			Database: "earth",
			Username: "earth",
			Password: "secret",
		}})
	if err == nil {
		t.Fatal("Expect validation error")
	}
}

func TestInitializeValidConfigUnreachableHost(t *testing.T) {
	db := NewDatabase(testLogger())
	err := db.Initialize(&config.Config{
		Credentials: config.Credential{
			Database: "earth",
			Username: "earth",
			Password: "secret",
		},
		DataSource: &config.DataSource{
			DatabaseHost: "unreachable",
			DatabasePort: "5432",
		},
	})
	if err == nil {
		t.Fatal("Expect unreachable host")
	}
}
