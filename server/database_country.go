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

type CountryQueryOptions struct {
	WithCities    bool
	WithContinent bool

	CountryUuids   []string
	ContinentTypes ContinentTypeList

	Deleted bool
}

func CountryOptionsFromQuery(r *http.Request) (CountryQueryOptions, error) {
	options := CountryQueryOptions{
		WithCities:    mhttp.QueryBoolDefault(r, "cities", false),
		WithContinent: mhttp.QueryBoolDefault(r, "continent", false),
		Deleted:       mhttp.QueryBoolDefault(r, "deleted", false),

		CountryUuids: mhttp.QueryList(r, "countries", ","),
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

func (db *Database) CountriesByOptions(options CountryQueryOptions) ([]*pkg_v1.Country, error) {
	started := time.Now()

	var (
		err     error
		results = []*pkg_v1.Country{}
		fields  = new(pkg_v1.Country).DatabaseFields()
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
		`SELECT %s FROM country`,
		fields,
	)

	if !options.Deleted {
		query += fmt.Sprintf(`AND deleted_state != %d`, msql.SoftDeleted)
	} else {
		query += fmt.Sprintf(`AND deleted_state = %d`, msql.SoftDeleted)
	}

	rows, err := db.postgres.Query(query)
	CheckOperation("CountrysByOptions", err, started)
	if err != nil {
		return results, err
	}

	defer rows.Close()

	for rows.Next() {
		var (
			curr    = &pkg_v1.Country{}
			updated sql.NullTime
		)
		if err = rows.Scan(
			&curr.Index,
			&curr.ContinentIndex,
			&curr.Uuid,
			&curr.Name,
			&curr.PhoneCode,
			&curr.ISOCode,
			&curr.Currency,
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
			Log.Warnf("DB.CountriesByOptions Scan error - %s", err.Error())
			rows.Close()
			break
		}
	}
	if err = rows.Err(); err != nil {
		Log.Warnf("DB.CountriesByOptions error - %s", err.Error())
		rows.Close()
		return results, err
	}

	if options.WithCities {
		// @todo import cities belong to country
	}

	if options.WithContinent {
		// @todo import continent belong to country
	} else {
		// @todo import continent uuid of country
	}

	return results, nil
}

func (db *Database) CountryByUuid(tx *sql.Tx, uuid string) (*pkg_v1.Country, error) {
	if _, err := muuid.UUIDFromString(uuid); err != nil {
		return nil, err
	}

	var (
		result  = &pkg_v1.Country{}
		started = time.Now()

		updated sql.NullTime
	)

	err := db.QueryRow(tx,
		fmt.Sprintf(
			`SELECT %s
				FROM country
			WHERE uuid = '%s'
			AND deleted_state != 1`,
			mstring.FormatFields(result.DatabaseFields()),
			uuid,
		)).Scan(
		&result.Index,
		&result.ContinentIndex,
		&result.Uuid,
		&result.Name,
		&result.PhoneCode,
		&result.ISOCode,
		&result.Currency,
		&result.Creator,
		&result.Created,
		&updated,
		&result.DeletedState,
	)

	if updated.Valid {
		result.Updated = updated.Time
	}

	CheckOperation("CountryByUuid", err, started)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (db *Database) IsCountryExist(tx *sql.Tx, country *pkg_v1.Country) (exist bool, err error) {

	var query = fmt.Sprintf(
		`SELECT uuid FROM country
		WHERE phone_code = '%s'
		OR iso_code = '%s'
		OR currency = '%s'
		AND deleted_state != %d `,
		country.PhoneCode,
		country.ISOCode,
		country.Currency,
		msql.SoftDeleted,
	)
	if muuid.UUIDValid(country.Uuid) {
		query += fmt.Sprintf("AND uuid != '%s'", country.Uuid.String())
	}
	if err := db.QueryRow(tx, fmt.Sprintf("SELECT EXISTS(%s)", query)).Scan(&exist); err != nil {
		return exist, err
	}

	return exist, nil
}

func (db *Database) CreateCountry(tx *sql.Tx, country *pkg_v1.Country) (*pkg_v1.Country, error) {
	if err := country.ValidateCreate(); err != nil {
		return nil, err
	}

	is_exist, err := DB.IsCountryExist(tx, country)
	if err != nil {
		return nil, err
	}

	if is_exist {
		return nil, errors.New("country already existed")
	}

	var (
		started         = time.Now()
		uuid            = muuid.NewUUID()
		json_creator, _ = json.Marshal(country.Creator)

		fields = []string{
			"continent_index",
			"uuid",
			"name",
			"phone_code",
			"iso_code",
			"currency",
			"creator",
		}
	)

	// @todo get continent index by uuid
	continent_index := 0

	_, err = db.Exec(tx,
		fmt.Sprintf(
			`INSERT INTO country(%s)
			VALUES(
				%d, '%s', '%s',
				'%s', '%s', '%s',
				'%s'
			)`,
			mstring.FormatFields(fields...),
			continent_index,
			uuid.String(),
			country.Name,
			country.PhoneCode,
			country.ISOCode,
			country.Currency,
			json_creator,
		))
	CheckOperation("CreateCountry", err, started)
	if err != nil {
		return nil, err
	}

	return DB.CountryByUuid(tx, uuid.String())
}

func (db *Database) UpdateCountry(tx *sql.Tx, country *pkg_v1.Country) (*pkg_v1.Country, error) {

	if !muuid.UUIDValid(country.Uuid) {
		return nil, errors.New("Invalid uuid")
	}

	if err := country.ValidateUpdate(); err != nil {
		return nil, err
	}

	type_exist, err := DB.IsCountryExist(tx, country)
	if err != nil {
		return nil, err
	}

	if type_exist {
		return nil, errors.New("country type already existed")
	}

	var (
		started = time.Now()
	)

	_, err = db.Exec(tx,
		fmt.Sprintf(
			`UPDATE country SET
			name='%s',
			phone_code='%s',
			iso_code='%s',
			currency='%s'
			WHERE uuid ='%s'
			AND deleted_state != 1`,
			country.Name,
			country.PhoneCode,
			country.ISOCode,
			country.Currency,
			country.Uuid.String(),
		))
	CheckOperation("UpdateCountry", err, started)
	if err != nil {
		return nil, err
	}

	return DB.CountryByUuid(tx, country.Uuid.String())
}

func (db *Database) SoftDeleteCountry(tx *sql.Tx, uuid string) error {

	if _, err := muuid.UUIDFromString(uuid); err != nil {
		return err
	}

	started := time.Now()

	_, err := db.Exec(tx, fmt.Sprintf(
		`UPDATE country SET
		deleted_state = %d
		WHERE uuid ='%s';`,
		msql.SoftDeleted,
		uuid,
	))
	CheckOperation("SoftDeleteCountry", err, started)
	if err != nil {
		return err
	}

	return nil
}
