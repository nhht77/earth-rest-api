package main

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	_ "github.com/lib/pq"

	"github.com/nhht77/earth-rest-api/server/pkg/msql"
	"github.com/nhht77/earth-rest-api/server/pkg/mstring"
)

type Database struct {
	postgres *sql.DB
}

func (db *Database) Initialize(c *Config) error {

	if err := c.ValidateConfig(); err != nil {
		return err
	}

	Log.Infof(
		"[postgre] Database Source: %s",
		c.DatabaseSourcePrintable(
			AppConfig.Framework.DatabaseHost,
			AppConfig.Framework.DatabasePort,
		),
	)

	result, err := sql.Open(
		"postgres",
		c.DatabaseSource(
			AppConfig.Framework.DatabaseHost,
			AppConfig.Framework.DatabasePort,
		))
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

func (db *Database) _ClearTable() error {

	tables := map[string]string{
		"continent": "continent_index_seq",
		"country":   "country_index_seq",
		"city":      "city_index_seq",
	}

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
		Log.Fatal("[testing] clearTable DB.Begin error", err)
		return err
	}

	if err := clear_table_by_map(tx, tables); err != nil {
		DB.Rollback(tx)
		Log.Fatal("[testing] clearTable clear_table_by_map error", err)
		return err
	}

	err = tx.Commit()
	if err != nil {
		Log.Fatal("[testing] clearTable tx.Commit error", err)
		return err
	}

	return nil
}
