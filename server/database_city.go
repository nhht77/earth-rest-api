package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/nhht77/earth-rest-api/mhttp"
	msql "github.com/nhht77/earth-rest-api/msql"
	"github.com/nhht77/earth-rest-api/mstring"
	"github.com/nhht77/earth-rest-api/muuid"
	pkg_v1 "github.com/nhht77/earth-rest-api/server/pkg"
)

type CityQueryOptions struct {
	WithCountry   bool
	WithContinent bool

	CountryUuids   []string
	CityUuids      []string
	ContinentTypes ContinentTypeList

	Deleted bool
}

func CityOptionsFromQuery(r *http.Request) (CityQueryOptions, error) {
	options := CityQueryOptions{
		WithCountry:   mhttp.QueryBoolDefault(r, "country", false),
		WithContinent: mhttp.QueryBoolDefault(r, "continent", false),
		Deleted:       mhttp.QueryBoolDefault(r, "deleted", false),

		CountryUuids: mhttp.QueryList(r, "countries", ","),
		CityUuids:    mhttp.QueryList(r, "cities", ","),
	}

	continent_types, err := mhttp.QueryIntList(r, "continent_types", ",")
	if err != nil {
		return options, err
	}
	if len(continent_types) > 0 {
		for _, v := range continent_types {
			options.ContinentTypes = append(options.ContinentTypes, pkg_v1.ContinentType(v))
		}
	}

	return options, nil
}

func (db *Database) CitiesByOptions(options CityQueryOptions) ([]*pkg_v1.City, error) {
	started := time.Now()

	var (
		err     error
		results = []*pkg_v1.City{}
		fields  = new(pkg_v1.City).DatabaseFields()
	)

	if len(options.ContinentTypes) == 0 {
		options.ContinentTypes = append(options.ContinentTypes,
			pkg_v1.ContinentType_Asia,
			pkg_v1.ContinentType_Africa,
			pkg_v1.ContinentType_Europe,
			pkg_v1.ContinentType_North_America,
			pkg_v1.ContinentType_South_America,
			pkg_v1.ContinentType_Oceania,
			pkg_v1.ContinentType_Antarctica,
		)
	}

	query := fmt.Sprintf(
		// @todo join continent for city continent type
		// @todo join country for city country
		`SELECT %s FROM city`,
		fields,
	)

	if !options.Deleted {
		query += fmt.Sprintf(`AND deleted_state != %d`, msql.SoftDeleted)
	} else {
		query += fmt.Sprintf(`AND deleted_state = %d`, msql.SoftDeleted)
	}

	rows, err := db.postgres.Query(query)
	CheckOperation("CitysByOptions", err, started)
	if err != nil {
		return results, err
	}

	defer rows.Close()

	for rows.Next() {
		var (
			curr    = &pkg_v1.City{}
			updated sql.NullTime
		)
		if err = rows.Scan(
			&curr.Index,
			&curr.ContinentIndex,
			&curr.Uuid,
			&curr.Name,
			&curr.Details,
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
			Log.Warnf("DB.CitiesByOptions Scan error - %s", err.Error())
			rows.Close()
			break
		}
	}
	if err = rows.Err(); err != nil {
		Log.Warnf("DB.CitiesByOptions error - %s", err.Error())
		rows.Close()
		return results, err
	}

	if options.WithCountry {
		// @todo import country belong to country
	}

	if options.WithContinent {
		// @todo import continent belong to country
	}

	return results, nil
}

func (db *Database) CityByUuid(tx *sql.Tx, uuid string) (*pkg_v1.City, error) {
	if _, err := muuid.UUIDFromString(uuid); err != nil {
		return nil, err
	}

	var (
		result  = &pkg_v1.City{}
		started = time.Now()

		updated sql.NullTime
	)

	err := db.QueryRow(tx,
		fmt.Sprintf(
			`SELECT %s
				FROM city
			WHERE uuid = '%s'
			AND deleted_state != 1`,
			mstring.FormatFields(result.DatabaseFields()),
			uuid,
		)).Scan(
		&result.Index,
		&result.ContinentIndex,
		&result.CountryIndex,
		&result.Uuid,
		&result.Name,
		&result.Details,
		&result.Creator,
		&result.Created,
		&updated,
		&result.DeletedState,
	)

	if updated.Valid {
		result.Updated = updated.Time
	}

	CheckOperation("CityByUuid", err, started)
	if err != nil {
		return nil, err
	}

	continent_uuid, err := DB.ContinentUuidByIndex(tx, result.ContinentIndex)
	if err != nil {
		return nil, err
	}

	country_uuid, err := DB.CountryUuidByIndex(tx, result.CountryIndex)
	if err != nil {
		return nil, err
	}

	result.ContinentUuid = continent_uuid
	result.CountryUuid = country_uuid

	return result, nil
}

func (db *Database) IsCapitalExist(tx *sql.Tx, city *pkg_v1.City) (exist bool, err error) {

	var query = fmt.Sprintf(
		`SELECT uuid FROM city
		WHERE (details->>'is_capital')::boolean = %t
		AND country_index = %d
		AND deleted_state != %d `,
		city.Details.IsCapital,
		city.CountryIndex,
		msql.SoftDeleted,
	)
	if muuid.UUIDValid(city.Uuid) {
		query += fmt.Sprintf("AND uuid != '%s'", city.Uuid.String())
	}
	if err := db.QueryRow(tx, fmt.Sprintf("SELECT EXISTS(%s)", query)).Scan(&exist); err != nil {
		return exist, err
	}

	return exist, nil
}

func (db *Database) CreateCity(tx *sql.Tx, city *pkg_v1.City) (*pkg_v1.City, error) {
	if err := city.ValidateCreate(); err != nil {
		return nil, err
	}

	is_exist, err := DB.IsCapitalExist(tx, city)
	if err != nil {
		return nil, err
	}

	if is_exist {
		return nil, errors.New("country already has capital")
	}

	var (
		started         = time.Now()
		uuid            = muuid.NewUUID()
		json_creator, _ = json.Marshal(city.Creator)
		json_details, _ = json.Marshal(city.Details)

		fields = []string{
			"continent_index",
			"country_index",
			"uuid",
			"name",
			"details",
			"creator",
		}
	)

	continent_index, err := DB.ContinentIndexByUuid(tx, city.ContinentUuid)
	if err != nil {
		return nil, err
	}
	country_index, err := DB.CountryIndexByUuid(tx, city.CountryUuid)
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(tx,
		fmt.Sprintf(
			`INSERT INTO city(%s)
			VALUES(
				%d, %d, '%s',
				'%s', '%s', '%s'
			)`,
			mstring.FormatFields(fields...),
			continent_index,
			country_index,
			uuid.String(),
			city.Name,
			json_details,
			json_creator,
		))
	CheckOperation("CreateCity", err, started)
	if err != nil {
		return nil, err
	}

	return DB.CityByUuid(tx, uuid.String())
}

func (db *Database) UpdateCity(tx *sql.Tx, city *pkg_v1.City) (*pkg_v1.City, error) {

	if !muuid.UUIDValid(city.Uuid) {
		return nil, errors.New("Invalid uuid")
	}

	if err := city.ValidateUpdate(); err != nil {
		return nil, err
	}

	is_exist, err := DB.IsCapitalExist(tx, city)
	if err != nil {
		return nil, err
	}

	if is_exist {
		return nil, errors.New("country already has capital")
	}

	var (
		started         = time.Now()
		json_details, _ = json.Marshal(city.Details)
	)

	_, err = db.Exec(tx,
		fmt.Sprintf(
			`UPDATE city SET
			name='%s',
			details='%s'
			WHERE uuid ='%s'
			AND deleted_state != 1`,
			city.Name,
			json_details,
			city.Uuid.String(),
		))
	CheckOperation("UpdateCity", err, started)
	if err != nil {
		return nil, err
	}

	return DB.CityByUuid(tx, city.Uuid.String())
}

func (db *Database) SoftDeleteCity(tx *sql.Tx, uuid string) error {

	if _, err := muuid.UUIDFromString(uuid); err != nil {
		return err
	}

	started := time.Now()

	_, err := db.Exec(tx, fmt.Sprintf(
		`UPDATE city SET
		deleted_state = %d
		WHERE uuid ='%s';`,
		msql.SoftDeleted,
		uuid,
	))
	CheckOperation("SoftDeleteCity", err, started)
	if err != nil {
		return err
	}

	return nil
}
