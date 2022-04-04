package msql

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	_ "github.com/lib/pq"

	mstring "github.com/nhht77/earth-rest-api/mstring"
)

func ReadStatementsFromFile(file string) ([]string, error) {
	file_path, err := filepath.Abs(fmt.Sprintf("./sql/%s", file))
	if err != nil {
		return []string{}, err
	}

	file_b, err := ioutil.ReadFile(file_path)
	if err != nil {
		return []string{}, err
	}

	statements := strings.Split(string(file_b), ";")

	if file == "02-trigger-function.sql" {
		statements = []string{string(file_b)}
	}

	return statements, nil
}

func IsCreateTable(statement string) bool {
	return strings.HasPrefix(strings.ToLower(statement), strings.ToLower("CREATE TABLE "))
}

func CreateTableExists(statement string, db *sql.DB) bool {
	if IsCreateTable(statement) {
		table := mstring.Between(statement, "CREATE TABLE IF NOT EXISTS ", " (")
		ignored := 0
		err := db.QueryRow(fmt.Sprintf("SELECT 1 FROM %s LIMIT 1", table)).Scan(&ignored)
		return err == nil
	}
	return false
}
