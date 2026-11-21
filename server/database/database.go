package database

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"

	"github.com/nhht77/earth-rest-api/pkg/logging"
	"github.com/nhht77/earth-rest-api/server/config"
)

type Database struct {
	postgres *sql.DB
	Log      logging.Logger
}

func NewDatabase(log logging.Logger) *Database {
	return &Database{Log: log}
}

func (db *Database) Initialize(c *config.Config) error {

	if err := c.ValidateConfig(); err != nil {
		return err
	}

	if err := c.ValidateDataSource(); err != nil {
		return err
	}

	db.Log.Infof(
		"[postgre] Database Source: %s",
		c.DatabaseSourcePrintable(),
	)

	result, err := sql.Open(
		"postgres",
		c.DatabaseSource(),
	)
	if err != nil {
		return err
	}
	db.postgres = result

	err = db.postgres.Ping()
	if err != nil {
		return err
	}

	// The application does NOT own the schema — migrations run via the
	// postgres image init (migrations/ mounted at /docker-entrypoint-initdb.d)
	// or a dedicated migration step. Startup only verifies presence and
	// fails fast with an actionable message.
	if err := db.verifySchema(); err != nil {
		return err
	}

	return nil
}

// verifySchema checks that the tables the queries expect exist.
// It deliberately does not create or repair anything: a missing table means
// the migration step was skipped, and the fix belongs there.
func (db *Database) verifySchema() error {
	const query = `
SELECT COUNT(*)
FROM information_schema.tables
WHERE table_schema = 'public'
  AND table_name IN ('city', 'country', 'continent')`

	var found int
	if err := db.postgres.QueryRow(query).Scan(&found); err != nil {
		return fmt.Errorf("[postgre] schema verification failed: %s", err.Error())
	}

	if found < 3 {
		return fmt.Errorf(
			"[postgre] schema incomplete: found %d/3 expected tables (city, country, continent) — "+
				"run the migrations first (migrations/ mounted at /docker-entrypoint-initdb.d, or 'make db-init'); "+
				"the application no longer creates the schema",
			found,
		)
	}

	db.Log.Infof("[postgre] schema verified: city, country, continent present")
	return nil
}

func (db *Database) Close() error {
	if db.postgres == nil {
		return nil
	}
	return db.postgres.Close()
}

func (db *Database) Rollback(tx *sql.Tx) {
	if tx == nil {
		return
	}
	err := tx.Rollback()
	// already committed or rolled back
	if err == sql.ErrTxDone {
		err = nil
	}
}

func (db *Database) Begin() (*sql.Tx, error) {
	return db.postgres.Begin()
}

func (db *Database) Query(tx *sql.Tx, query string) (*sql.Rows, error) {
	if tx != nil {
		return tx.Query(query)
	}
	return db.postgres.Query(query)
}

func (db *Database) QueryRow(tx *sql.Tx, query string) *sql.Row {
	if tx != nil {
		return tx.QueryRow(query)
	}
	return db.postgres.QueryRow(query)
}

func (db *Database) Exec(tx *sql.Tx, query string) (sql.Result, error) {
	if tx != nil {
		return tx.Exec(query)
	}
	return db.postgres.Exec(query)
}

func DatabaseNoResults(err error) bool {
	return err == sql.ErrNoRows
}

func (db *Database) CheckOperation(op string, err error, started time.Time) bool {
	spent := ""
	if !started.IsZero() {
		spent = time.Since(started).String()
	}

	hasError := err != nil && err != sql.ErrNoRows
	if hasError {
		db.Log.Errorf("[postgre] db.%s error: %s (%s)", op, err.Error(), spent)
		return true
	}

	db.Log.Infof("[postgre] db.%s %s", op, spent)
	return hasError == false
}

func (db *Database) ClearTable() error {

	tables := map[string]string{
		"continent": "continent_index_seq",
		"country":   "country_index_seq",
		"city":      "city_index_seq",
	}

	clear_table_by_map := func(tx *sql.Tx, tables map[string]string) error {
		for table, sequence := range tables {
			_, err := db.Exec(tx, fmt.Sprintf("TRUNCATE %s CASCADE", table))
			if err != nil {
				return err
			}
			_, err = db.Exec(tx, fmt.Sprintf("ALTER SEQUENCE %s RESTART WITH 1", sequence))
			if err != nil {
				return err
			}
		}
		return nil
	}

	tx, err := db.Begin()
	if err != nil {
		db.Log.Fatal("[testing] clearTable db.Begin error", err)
		return err
	}

	if err := clear_table_by_map(tx, tables); err != nil {
		db.Rollback(tx)
		db.Log.Fatal("[testing] clearTable clear_table_by_map error", err)
		return err
	}

	err = tx.Commit()
	if err != nil {
		db.Log.Fatal("[testing] clearTable tx.Commit error", err)
		return err
	}

	return nil
}
