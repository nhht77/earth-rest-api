package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/nhht77/earth-rest-api/mhttp"
	msql "github.com/nhht77/earth-rest-api/msql"
	"github.com/nhht77/earth-rest-api/mstring"
	"github.com/nhht77/earth-rest-api/muuid"
	pkg_v1 "github.com/nhht77/earth-rest-api/server/pkg"
)

type ContinentTypeList []pkg_v1.ContinentType

func (types ContinentTypeList) String() (str string) {
	for _, iter := range types {
		if len(str) > 0 {
			str += ","
		}
		str += strconv.FormatUint(uint64(iter), 10)
	}
	return str
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
			options.Types = append(options.Types, pkg_v1.ContinentType(v))
		}
	}
	return options, nil
}

func (db *Database) ContinentsByOptions(options ContinentQueryOptions) ([]*pkg_v1.Continent, error) {
	started := time.Now()

	if len(options.Types) == 0 {
		options.Types = append(options.Types,
			pkg_v1.ContinentType_Asia,
			pkg_v1.ContinentType_Africa,
			pkg_v1.ContinentType_Europe,
			pkg_v1.ContinentType_North_America,
			pkg_v1.ContinentType_South_America,
			pkg_v1.ContinentType_Oceania,
			pkg_v1.ContinentType_Antarctica,
		)
	}

	var (
		err     error
		results = []*pkg_v1.Continent{}
		fields  = new(pkg_v1.Continent).DatabaseFields()
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
	CheckOperation("ContinentsByOptions", err, started)
	if err != nil {
		return results, err
	}

	defer rows.Close()

	for rows.Next() {
		var (
			curr    = &pkg_v1.Continent{}
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
			Log.Warnf("DB.ContinentsByOptions Scan error - %s", err.Error())
			rows.Close()
			break
		}
	}
	if err = rows.Err(); err != nil {
		Log.Warnf("DB.ContinentsByOptions error - %s", err.Error())
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

func (db *Database) ContinentByUuid(tx *sql.Tx, uuid string) (*pkg_v1.Continent, error) {
	if _, err := muuid.UUIDFromString(uuid); err != nil {
		return nil, err
	}

	var (
		result  = &pkg_v1.Continent{}
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

	CheckOperation("ContinentByUuid", err, started)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (db *Database) IsContinentTypeExist(tx *sql.Tx, continent *pkg_v1.Continent) (exist bool, err error) {

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

func (db *Database) CreateContinent(tx *sql.Tx, continent *pkg_v1.Continent) (*pkg_v1.Continent, error) {
	if err := continent.ValidateCreate(); err != nil {
		return nil, err
	}

	type_exist, err := DB.IsContinentTypeExist(tx, continent)
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
	CheckOperation("CreateContinent", err, started)
	if err != nil {
		return nil, err
	}

	return DB.ContinentByUuid(tx, uuid.String())
}

// update continent
func (db *Database) UpdateContinent(tx *sql.Tx, continent *pkg_v1.Continent) (*pkg_v1.Continent, error) {

	if !muuid.UUIDValid(continent.Uuid) {
		return nil, errors.New("Invalid uuid")
	}

	if err := continent.ValidateUpdate(); err != nil {
		return nil, err
	}

	type_exist, err := DB.IsContinentTypeExist(tx, continent)
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
	CheckOperation("UpdateContinent", err, started)
	if err != nil {
		return nil, err
	}

	return DB.ContinentByUuid(tx, continent.Uuid.String())
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
	CheckOperation("SoftDeleteContinent", err, started)
	if err != nil {
		return err
	}

	return nil
}
