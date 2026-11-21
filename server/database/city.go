package database

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	city "github.com/nhht77/earth-rest-api/pkg/city"
	continent "github.com/nhht77/earth-rest-api/pkg/continent"
	country "github.com/nhht77/earth-rest-api/pkg/country"
	"github.com/nhht77/earth-rest-api/pkg/handlers"
	"github.com/nhht77/earth-rest-api/pkg/msql"
	"github.com/nhht77/earth-rest-api/pkg/mstring"
	muuid "github.com/nhht77/earth-rest-api/pkg/muuid"
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
		WithCountry:   handlers.QueryBoolDefault(r, "with_country", false),
		WithContinent: handlers.QueryBoolDefault(r, "with_continent", false),
		Deleted:       handlers.QueryBoolDefault(r, "deleted", false),

		CountryUuids: handlers.QueryList(r, "countries", ","),
		CityUuids:    handlers.QueryList(r, "cities", ","),
	}

	continent_types, err := handlers.QueryIntList(r, "continent_types", ",")
	if err != nil {
		return options, err
	}
	if len(continent_types) > 0 {
		for _, v := range continent_types {
			options.ContinentTypes = append(options.ContinentTypes, continent.ContinentType(v))
		}
	}

	return options, nil
}

func (db *Database) CitiesByOptions(options CityQueryOptions) ([]*city.City, error) {

	var (
		err     error
		started = time.Now()
		results = []*city.City{}
	)

	if len(options.ContinentTypes) == 0 {
		options.ContinentTypes = append(options.ContinentTypes,
			continent.ContinentType_Asia,
			continent.ContinentType_Africa,
			continent.ContinentType_Europe,
			continent.ContinentType_North_America,
			continent.ContinentType_South_America,
			continent.ContinentType_Oceania,
			continent.ContinentType_Antarctica,
		)
	}

	query := fmt.Sprintf(`SELECT %s FROM city `, new(city.City).DatabaseFields())

	if !options.Deleted {
		query += fmt.Sprintf(`WHERE deleted_state != %d`, msql.SoftDeleted)
	} else {
		query += fmt.Sprintf(`WHERE deleted_state = %d`, msql.SoftDeleted)
	}

	if len(options.CityUuids) > 0 {
		query += fmt.Sprintf(`AND uuid IN (%s) `, mstring.FormatStringValues(options.CityUuids...))
	}

	rows, err := db.postgres.Query(query)
	db.CheckOperation("CitysByOptions", err, started)
	if err != nil {
		return results, err
	}

	defer rows.Close()

	for rows.Next() {
		var (
			curr    = &city.City{}
			updated sql.NullTime
		)
		if err = rows.Scan(
			&curr.Index,
			&curr.ContinentIndex,
			&curr.CountryIndex,
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
			db.Log.Warnf("db.CitiesByOptions Scan error - %s", err.Error())
			rows.Close()
			break
		}
	}
	if err = rows.Err(); err != nil {
		db.Log.Warnf("db.CitiesByOptions error - %s", err.Error())
		rows.Close()
		return results, err
	}

	filter_results, err := db.ToCities(started, results, options)
	if err != nil {
		return results, err
	}

	return filter_results, nil
}

func (db *Database) ToCities(started time.Time, cities city.CityList, options CityQueryOptions) (city.CityList, error) {
	if started.IsZero() {
		started = time.Now()
	}

	continents, err := db.ContinentsByOptions(ContinentQueryOptions{
		Types: options.ContinentTypes,
	})
	db.CheckOperation("ToCities ", err, started)
	if err != nil {
		return city.CityList{}, err
	}

	countries, err := db.CountriesByOptions(CountryQueryOptions{
		CountryUuids: options.CountryUuids,
	})
	db.CheckOperation("ToCities ", err, started)
	if err != nil {
		return city.CityList{}, err
	}

	var (
		continent_map = map[msql.DatabaseIndex]*continent.Continent{}
		country_map   = map[msql.DatabaseIndex]*country.Country{}
		filter_cities = city.CityList{}
	)

	// create continents map
	for _, iter := range continents {
		continent_map[iter.Index] = iter
	}

	// create country map
	for _, iter := range countries {
		country_map[iter.Index] = iter
	}

	for _, iter := range cities {

		iter_country := country_map[iter.CountryIndex]
		iter_continent := continent_map[iter.ContinentIndex]

		if iter_continent == nil || iter_country == nil {
			continue
		}

		if !options.ContinentTypes.Contains(iter_continent.Type) {
			continue
		}

		is_country_not_match := len(options.CountryUuids) > 0 && !mstring.SliceContains(options.CountryUuids, iter_country.Uuid.String())
		if is_country_not_match {
			continue
		}

		if options.WithContinent {
			iter.Details.Continent = iter_continent
		}

		if options.WithCountry {
			iter.Details.Country = iter_country
		}

		iter.ContinentUuid = iter_continent.Uuid
		iter.CountryUuid = iter_country.Uuid

		filter_cities = append(filter_cities, iter)
	}
	return filter_cities, nil
}

func (db *Database) CityByUuid(tx *sql.Tx, uuid string) (*city.City, error) {
	if _, err := muuid.UUIDFromString(uuid); err != nil {
		return nil, err
	}

	var (
		result  = &city.City{}
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

	db.CheckOperation("CityByUuid", err, started)
	if err != nil {
		return nil, err
	}

	continent_uuid, err := db.ContinentUuidByIndex(tx, result.ContinentIndex)
	if err != nil {
		return nil, err
	}

	country_uuid, err := db.CountryUuidByIndex(tx, result.CountryIndex)
	if err != nil {
		return nil, err
	}

	result.ContinentUuid = continent_uuid
	result.CountryUuid = country_uuid

	return result, nil
}

// Note: Country can have only one capital
func (db *Database) IsCapitalExist(tx *sql.Tx, city *city.City, country *country.Country) (exist bool, err error) {

	if city.Details != nil && !city.Details.IsCapital {
		return false, nil
	}

	var query = fmt.Sprintf(
		`SELECT uuid FROM city
		WHERE (details->>'is_capital')::boolean = true
		AND country_index = %d
		AND deleted_state != %d `,
		country.Index,
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

func (db *Database) CreateCity(tx *sql.Tx, city *city.City) (*city.City, error) {
	if err := city.ValidateCreate(); err != nil {
		return nil, err
	}

	country, err := db.CountryByUuid(tx, city.CountryUuid.String())
	if err != nil {
		return nil, err
	}

	is_exist, err := db.IsCapitalExist(tx, city, country)
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

	continent_index, err := db.ContinentIndexByUuid(tx, city.ContinentUuid)
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
			country.Index,
			uuid.String(),
			city.Name,
			json_details,
			json_creator,
		))
	db.CheckOperation("CreateCity", err, started)
	if err != nil {
		return nil, err
	}

	return db.CityByUuid(tx, uuid.String())
}

func (db *Database) UpdateCity(tx *sql.Tx, city *city.City) (*city.City, error) {

	if !muuid.UUIDValid(city.Uuid) {
		return nil, errors.New("Invalid uuid")
	}

	if err := city.ValidateUpdate(); err != nil {
		return nil, err
	}

	country, err := db.CountryByUuid(tx, city.CountryUuid.String())
	if err != nil {
		return nil, err
	}

	is_exist, err := db.IsCapitalExist(tx, city, country)
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
	db.CheckOperation("UpdateCity", err, started)
	if err != nil {
		return nil, err
	}

	return db.CityByUuid(tx, city.Uuid.String())
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
	db.CheckOperation("SoftDeleteCity", err, started)
	if err != nil {
		return err
	}

	return nil
}
