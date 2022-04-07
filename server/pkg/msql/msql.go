package msql

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strconv"
	"strings"

	_ "github.com/lib/pq"

	mstring "github.com/nhht77/earth-rest-api/server/pkg/mstring"
)

////////////////////////////
/////// Basic schema struct

type DatabaseIndex uint

type DatabaseIndexList []DatabaseIndex

func (indexes DatabaseIndexList) String() (str string) {
	for _, index := range indexes {
		if len(str) > 0 {
			str += ","
		}
		str += strconv.FormatUint(uint64(index), 10)
	}
	return str
}

type DeletedState int

const (
	NotDeleted  DeletedState = 0
	SoftDeleted DeletedState = 1
)

//////////////////////////
/////// Basic DB function

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

func FormatFields(field ...string) string {
	return strings.Join(field, ", ")
}

func JSONScan(src, dest interface{}) error {
	if src == nil {
		return nil
	}
	switch src.(type) {
	case []byte:
		return json.Unmarshal(src.([]byte), dest)
	case sql.RawBytes:
		return json.Unmarshal(src.(sql.RawBytes), dest)
	}
	return fmt.Errorf("JSONScan: cannot read %T from %T", dest, src)
}

func JSONValue(v interface{}) (driver.Value, error) {
	return json.Marshal(v)
}
