package database

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	continent "github.com/nhht77/earth-rest-api/pkg/continent"
	"github.com/nhht77/earth-rest-api/pkg/mhttp"
	"github.com/nhht77/earth-rest-api/pkg/msql"
	"github.com/nhht77/earth-rest-api/pkg/mstring"
	muuid "github.com/nhht77/earth-rest-api/pkg/muuid"
)

type ContinentTypeList []continent.ContinentType

func (types ContinentTypeList) String() (str string) {
	for _, iter := range types {
		if len(str) > 0 {
			str += ","
		}
		str += strconv.FormatUint(uint64(iter), 10)
	}
	return str
}

func (types ContinentTypeList) Contains(value continent.ContinentType) bool {
	for _, iter := range types {
		if value == iter {
			return true
		}
	}
	return false
}

type ContinentQueryOptions struct {
	WithCities    bool
	WithCountries bool

	Types ContinentTypeList

	Deleted bool
}

func ContinentOptionsFromQuery(r *http.Request) (ContinentQueryOptions, error) {
	options := ContinentQueryOptions{
		WithCities:    mhttp.QueryBoolDefault(r, "cities", false),
		WithCountries: mhttp.QueryBoolDefault(r, "countries", false),
		Deleted:       mhttp.QueryBoolDefault(r, "deleted", false),
	}

	types, err := mhttp.QueryIntList(r, "types", ",")
	if err != nil {
		return options, err
	}
	if len(types) > 0 {
		for _, v := range types {
			options.Types = append(options.Types, continent.ContinentType(v))
		}
	}
	return options, nil
}

func (db *Database) ContinentsByOptions(options ContinentQueryOptions) ([]*continent.Continent, error) {
	started := time.Now()

	if len(options.Types) == 0 {
		options.Types = append(options.Types,
			continent.ContinentType_Asia,
			continent.ContinentType_Africa,
			continent.ContinentType_Europe,
			continent.ContinentType_North_America,
			continent.ContinentType_South_America,
			continent.ContinentType_Oceania,
			continent.ContinentType_Antarctica,
		)
	}

	var (
		err     error
		results = []*continent.Continent{}
		fields  = new(continent.Continent).DatabaseFields()
	)

	query := fmt.Sprintf(
		`SELECT %s FROM continent WHERE type IN (%s) `,
		fields,
		options.Types.String(),
	)

	if !options.Deleted {
		query += fmt.Sprintf(`AND deleted_state != %d`, msql.SoftDeleted)
	} else {
		query += fmt.Sprintf(`AND deleted_state = %d`, msql.SoftDeleted)
	}

	rows, err := db.postgres.Query(query)
	db.CheckOperation("ContinentsByOptions", err, started)
	if err != nil {
		return results, err
	}

	defer rows.Close()

	for rows.Next() {
		var (
			curr    = &continent.Continent{}
			updated sql.NullTime
		)
		if err = rows.Scan(
			&curr.Index,
			&curr.Uuid,
			&curr.Name,
			&curr.Type,
			&curr.AreaByKm2,
			&curr.Creator,
			&curr.Created,
			&updated,
			&curr.DeletedState,
		); err == nil {

			if updated.Valid && !updated.Time.IsZero() {
				curr.Updated = updated.Time
			}

			results = append(results, curr)
		} else {
			db.Log.Warnf("db.ContinentsByOptions Scan error - %s", err.Error())
			rows.Close()
			break
		}
	}
	if err = rows.Err(); err != nil {
		db.Log.Warnf("db.ContinentsByOptions error - %s", err.Error())
		rows.Close()
		return results, err
	}

	if options.WithCities {
		// @todo import cities belong to continent
	}

	if options.WithCountries {
		// @todo import countries belong to continent
	}

	return results, nil
}

func (db *Database) ContinentByUuid(tx *sql.Tx, uuid string) (*continent.Continent, error) {
	if _, err := muuid.UUIDFromString(uuid); err != nil {
		return nil, err
	}

	var (
		result  = &continent.Continent{}
		started = time.Now()

		updated sql.NullTime
	)

	err := db.QueryRow(tx,
		fmt.Sprintf(
			`SELECT %s
				FROM continent
			WHERE uuid = '%s'
			AND deleted_state != 1`,
			mstring.FormatFields(result.DatabaseFields()),
			uuid,
		)).Scan(
		&result.Index,
		&result.Uuid,
		&result.Name,
		&result.Type,
		&result.AreaByKm2,
		&result.Creator,
		&result.Created,
		&updated,
		&result.DeletedState,
	)

	if updated.Valid {
		result.Updated = updated.Time
	}

	db.CheckOperation("ContinentByUuid", err, started)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (db *Database) ContinentUuidByIndex(tx *sql.Tx, index msql.DatabaseIndex) (muuid.UUID, error) {
	var (
		uuid    = muuid.UUID{}
		started = time.Now()
	)

	if index == 0 {
		return uuid, errors.New("Invalid index")
	}

	err := db.QueryRow(tx,
		fmt.Sprintf(
			`SELECT
				uuid
			FROM continent
			WHERE index IN (%d)
			AND deleted_state != 1`,
			index,
		)).Scan(&uuid)

	db.CheckOperation("ContinentUuidByIndex", err, started)
	if err != nil {
		return uuid, err
	}

	return uuid, nil
}

func (db *Database) ContinentUuidsByIndexes(indexes msql.DatabaseIndexList) (map[msql.DatabaseIndex]muuid.UUID, error) {
	var (
		indexes_map = map[msql.DatabaseIndex]muuid.UUID{}
		started     = time.Now()
	)

	if len(indexes) == 0 {
		return indexes_map, errors.New("Invalid continent index")
	}

	rows, err := db.postgres.Query(
		fmt.Sprintf(
			`SELECT
				index,
				uuid
			FROM continent
			WHERE index IN (%s)
			AND deleted_state != 1`,
			indexes.String(),
		))

	db.CheckOperation("ContinentUuidsByIndexes", err, started)
	if err != nil {
		return indexes_map, err
	}
	defer rows.Close()

	for rows.Next() {
		var (
			curr_index msql.DatabaseIndex
			curr_uuid  = muuid.UUID{}
		)
		if err = rows.Scan(
			&curr_index,
			&curr_uuid,
		); err == nil {
			indexes_map[curr_index] = curr_uuid
		} else {
			db.CheckOperation("ContinentUuidsByIndexes Scan error", err, started)
			rows.Close()
			break
		}
	}

	return indexes_map, nil
}

func (db *Database) ContinentIndexByUuid(tx *sql.Tx, uuid muuid.UUID) (msql.DatabaseIndex, error) {
	var (
		index   msql.DatabaseIndex
		started = time.Now()
	)

	if !muuid.UUIDValid(uuid) {
		return index, errors.New("Invalid continent uuid")
	}

	err := db.QueryRow(tx,
		fmt.Sprintf(
			`SELECT
				index
			FROM continent
			WHERE uuid IN ('%s')
			AND deleted_state != 1`,
			uuid.String(),
		)).Scan(&index)

	db.CheckOperation("ContinentIndexByUuid", err, started)
	if err != nil {
		return index, err
	}

	return index, nil
}

func (db *Database) ContinentUuidsByIndex(tx *sql.Tx, indexes msql.DatabaseIndexList) (*continent.Continent, error) {
	if len(indexes) == 0 {
		return nil, errors.New("Invalid index")
	}

	var (
		result  = &continent.Continent{}
		started = time.Now()
	)

	rows, err := db.postgres.Query(
		fmt.Sprintf(
			`SELECT
				uuid,
				index
				FROM continent
			WHERE index IN (%s)
			AND deleted_state != 1`,
			indexes.String(),
		))

	defer rows.Close()

	for rows.Next() {

	}

	db.CheckOperation("ContinentByUuid", err, started)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (db *Database) IsContinentTypeExist(tx *sql.Tx, continent *continent.Continent) (exist bool, err error) {

	var query = fmt.Sprintf(
		`SELECT uuid FROM continent WHERE type = %d AND deleted_state != %d `,
		continent.Type,
		msql.SoftDeleted,
	)
	if muuid.UUIDValid(continent.Uuid) {
		query += fmt.Sprintf("AND uuid != '%s'", continent.Uuid.String())
	}
	if err := db.QueryRow(tx, fmt.Sprintf("SELECT EXISTS(%s)", query)).Scan(&exist); err != nil {
		return exist, err
	}

	return exist, nil
}

func (db *Database) CreateContinent(tx *sql.Tx, continent *continent.Continent) (*continent.Continent, error) {
	if err := continent.ValidateCreate(); err != nil {
		return nil, err
	}

	type_exist, err := db.IsContinentTypeExist(tx, continent)
	if err != nil {
		return nil, err
	}

	if type_exist {
		return nil, errors.New("continent type already existed")
	}

	var (
		started         = time.Now()
		uuid            = muuid.NewUUID()
		json_creator, _ = json.Marshal(continent.Creator)

		fields = []string{
			"uuid",
			"name",
			"type",
			"area_by_km2",
			"creator",
		}
	)

	_, err = db.Exec(tx,
		fmt.Sprintf(
			`INSERT INTO continent(%s)
			VALUES(
				'%s', '%s', %d,
				%f, '%s'
			)`,
			mstring.FormatFields(fields...),
			uuid.String(),
			continent.Name,
			continent.Type,
			continent.AreaByKm2,
			json_creator,
		))
	db.CheckOperation("CreateContinent", err, started)
	if err != nil {
		return nil, err
	}

	return db.ContinentByUuid(tx, uuid.String())
}

// update continent
func (db *Database) UpdateContinent(tx *sql.Tx, continent *continent.Continent) (*continent.Continent, error) {

	if !muuid.UUIDValid(continent.Uuid) {
		return nil, errors.New("Invalid uuid")
	}

	if err := continent.ValidateUpdate(); err != nil {
		return nil, err
	}

	type_exist, err := db.IsContinentTypeExist(tx, continent)
	if err != nil {
		return nil, err
	}

	if type_exist {
		return nil, errors.New("continent type already existed")
	}

	var (
		started = time.Now()
	)

	_, err = db.Exec(tx,
		fmt.Sprintf(
			`UPDATE continent SET
			name='%s',
			type=%d,
			area_by_km2=%f
			WHERE uuid ='%s'
			AND deleted_state != 1`,
			continent.Name,
			continent.Type,
			continent.AreaByKm2,
			continent.Uuid.String(),
		))
	db.CheckOperation("UpdateContinent", err, started)
	if err != nil {
		return nil, err
	}

	return db.ContinentByUuid(tx, continent.Uuid.String())
}

// delete continent
func (db *Database) SoftDeleteContinent(tx *sql.Tx, uuid string) error {

	if _, err := muuid.UUIDFromString(uuid); err != nil {
		return err
	}

	started := time.Now()

	_, err := db.Exec(tx, fmt.Sprintf(
		`UPDATE continent SET
		deleted_state = %d
		WHERE uuid ='%s';`,
		msql.SoftDeleted,
		uuid,
	))
	db.CheckOperation("SoftDeleteContinent", err, started)
	if err != nil {
		return err
	}

	return nil
}
