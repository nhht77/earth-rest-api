package pkg_v1

import (
	"database/sql/driver"
	"errors"
	"time"

	"github.com/nhht77/earth-rest-api/msql"
	muuid "github.com/nhht77/earth-rest-api/muuid"
)

type City struct {
	Index          msql.DatabaseIndex `json:"-"`
	ContinentIndex msql.DatabaseIndex `json:"-"`
	CountryIndex   msql.DatabaseIndex `json:"-"`

	ContinentUuid muuid.UUID `json:"continent_uuid"`
	CountryUuid   muuid.UUID `json:"country_uuid"`
	Uuid          muuid.UUID `json:"uuid"`

	Name    string       `json:"name"`
	Details *CityDetails `json:"details"`

	Created time.Time `json:"created"`
	Updated time.Time `json:"updated"`

	Creator *UserMinimal `json:"creator"`

	DeletedState msql.DeletedState `json:"-"`
}

type CityDetails struct {
	IsCapital bool `json:"is_capital"`
}

func (v *CityDetails) Value() (driver.Value, error) {
	return msql.JSONValue(v)
}

func (v *CityDetails) Scan(src interface{}) error {
	if src != nil && v == nil {
		v = &CityDetails{}
	}
	return msql.JSONScan(src, v)
}

func (obj *City) ValidateCreate() error {

	if !muuid.UUIDValid(obj.ContinentUuid) {
		return errors.New("Invalid continent uuid")
	}

	if !muuid.UUIDValid(obj.CountryUuid) {
		return errors.New("Invalid country uuid")
	}

	if len(obj.Name) == 0 {
		return errors.New("Invalid city name")
	}

	if err := obj.Creator.IsValid(); err != nil {
		return err
	}

	return nil
}

func (obj *City) ValidateUpdate() error {

	if !muuid.UUIDValid(obj.Uuid) {
		return errors.New("Invalid city uuid")
	}

	if obj.Details == nil {
		return errors.New("Invalid city details")
	}

	if len(obj.Name) == 0 {
		return errors.New("Invalid country name")
	}

	return nil
}

func (obj *City) DatabaseFields() string {
	return msql.FormatFields(
		"index", "continent_index",
		"country_index", "uuid",
		"name", "details", "creator",
		"created", "updated", "deleted_state",
	)
}
