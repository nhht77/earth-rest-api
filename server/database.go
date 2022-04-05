package main

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	_ "github.com/lib/pq"

	msql "github.com/nhht77/earth-rest-api/msql"
	mstring "github.com/nhht77/earth-rest-api/mstring"
)

var (
	host = "127.0.0.1"
	port = "5432"
)

type Database struct {
	postgres *sql.DB
}

func (db *Database) Initialize(c *Config) error {

	if err := c.ValidateConfig(); err != nil {
		return err
	}

	Log.Infof("[postgre] Database Source: %s", c.DatabaseSourcePrintable(host, port))

	result, err := sql.Open("postgres", c.DatabaseSource(host, port))
	if err != nil {
		return err
	}
	db.postgres = result

	err = db.postgres.Ping()
	if err != nil {
		return err
	}

	for _, migration_file := range []string{
		"01-create-table.sql",
		"02-trigger-function.sql",
		"03-create-trigger.sql",
	} {
		statements, err := msql.ReadStatementsFromFile(migration_file)
		if err != nil {
			return err
		}

		for stmt_idx, stmt := range statements {
			var (
				started      = time.Now()
				tableCreated = false
			)

			if has_condition := strings.Contains(stmt, "-- condition:"); has_condition {
				var (
					condition_value = true
					condition_query = mstring.Between(stmt, "-- condition:", "condition --")
				)
				if err := db.postgres.QueryRow(condition_query).Scan(&condition_value); err != nil {
					return fmt.Errorf("[postgre] %s #%d: %s", migration_file, stmt_idx+1, err.Error())
				}
				if condition_value {
					Log.Infof("[postgre] %s #%d skipped due to condition result is true", migration_file, stmt_idx+1)
					continue
				}
			}

			// check table creation
			if msql.IsCreateTable(stmt) && !msql.CreateTableExists(stmt, db.postgres) {
				tableCreated = true
			}

			results, err := db.postgres.Exec(stmt)
			if err != nil {
				return err
			}

			changed := db.resultChange(results)
			if changed || !tableCreated {
				duration := time.Since(started)
				Log.Infof("[postgre] %s #%d ran in %s", migration_file, stmt_idx+1, duration.String())
			}
		}
	}

	return nil
}

func (db *Database) Close() error {
	if db.postgres == nil {
		return nil
	}
	return db.postgres.Close()
}

func (db *Database) resultChange(result sql.Result) bool {
	num1, _ := result.RowsAffected()
	num2, _ := result.LastInsertId()
	return num1 > 0 || num2 > 0
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

func CheckOperation(op string, err error, started time.Time) bool {
	spent := ""
	if !started.IsZero() {
		spent = time.Since(started).String()
	}

	hasError := err != nil && err != sql.ErrNoRows
	if hasError {
		Log.Errorf("[postgre] DB.%s error: %s (%s)", op, err.Error(), spent)
		return true
	}

	Log.Infof("[postgre] DB.%s %s", op, spent)
	return hasError == false
}
