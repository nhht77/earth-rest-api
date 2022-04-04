package main_test

import (
	"database/sql"
	"fmt"
	"os"
	"testing"

	main "github.com/nhht77/earth-rest-api/server"
	"github.com/sirupsen/logrus"
)

var (
	Log = func() *logrus.Logger {
		log := logrus.New()
		log.WithField("[test]", "")
		return log
	}()

	DB        = &main.Database{}
	AppConfig = &main.Config{}
)

func TestMain(m *testing.M) {

	AppConfig.ReadTestDefault()

	err := DB.Initialize(AppConfig)
	if err != nil {
		Log.Fatalf("[test] DB.Initialize error %s", err)
	}
	var (
		tables = map[string]string{
			"continent": "continent_index_seq",
		}
	)

	ensureTableExists(tables)

	code := m.Run()

	clearTable(tables)

	os.Exit(code)
}

func ensureTableExists(tables map[string]string) {

	for table, _ := range tables {
		_, table_exist := DB.Query(nil, fmt.Sprintf("SELECT uuid FROM %s", table))
		if table_exist != nil {
			Log.Fatal(table_exist)
		}
	}
}

func clearTable(tables map[string]string) {

	clear_table_by_map := func(tx *sql.Tx, tables map[string]string) error {
		for table, sequence := range tables {
			_, err := DB.Exec(tx, fmt.Sprintf("TRUNCATE %s CASCADE", table))
			if err != nil {
				return err
			}
			_, err = DB.Exec(tx, fmt.Sprintf("ALTER SEQUENCE %s RESTART WITH 1", sequence))
			if err != nil {
				return err
			}
		}
		return nil
	}

	tx, err := DB.Begin()
	if err != nil {
		Log.Fatal("clearTable DB.Begin error", err)
	}

	if err := clear_table_by_map(tx, tables); err != nil {
		DB.Rollback(tx)
		Log.Fatal("clearTable clear_table_by_map error", err)
		return
	}

	err = tx.Commit()
	if err != nil {
		Log.Fatal("clearTable tx.Commit error", err)
	}
}
