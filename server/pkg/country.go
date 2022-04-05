package pkg_v1

import (
	"database/sql/driver"
	"errors"
	"time"

	"github.com/nhht77/earth-rest-api/msql"
	muuid "github.com/nhht77/earth-rest-api/muuid"
)

type Country struct {
	Index          msql.DatabaseIndex `json:"-"`
	ContinentIndex msql.DatabaseIndex `json:"-"`

	ContinentUuid muuid.UUID `json:"continent_uuid"`
	Uuid          muuid.UUID `json:"uuid"`

	Name    string          `json:"name"`
	Details *CountryDetails `json:"details"`

	Created time.Time `json:"created"`
	Updated time.Time `json:"updated"`

	Creator *UserMinimal `json:"creator"`

	DeletedState msql.DeletedState `json:"-"`
}

type CountryDetails struct {
	PhoneCode string `json:"phone_code"`
	ISOCode   string `json:"iso_code"`
	Currency  string `json:"currency"`
}

func (v *CountryDetails) Value() (driver.Value, error) {
	return msql.JSONValue(v)
}

func (v *CountryDetails) Scan(src interface{}) error {
	if src != nil && v == nil {
		v = &CountryDetails{}
	}
	return msql.JSONScan(src, v)
}

func (details *CountryDetails) Validate() error {
	if len(details.PhoneCode) == 0 {
		return errors.New("Invalid country phone code")
	}

	if len(details.ISOCode) == 0 {
		return errors.New("Invalid country iso code")
	}

	if len(details.Currency) == 0 {
		return errors.New("Invalid country currency")
	}

	return nil
}

func (obj *Country) ValidateCreate() error {

	if len(obj.Name) == 0 {
		return errors.New("Invalid country name")
	}

	if obj.Details == nil || obj.Creator == nil {
		return errors.New("Empty country details || creator")
	}

	if err := obj.Details.Validate(); err != nil {
		return err
	}

	if err := obj.Creator.IsValid(); err != nil {
		return err
	}

	return nil
}

func (obj *Country) ValidateUpdate() error {

	if !muuid.UUIDValid(obj.Uuid) {
		return errors.New("Invalid country uuid")
	}

	if len(obj.Name) == 0 {
		return errors.New("Invalid country name")
	}

	if obj.Details == nil {
		return errors.New("Empty country details")
	}

	if err := obj.Details.Validate(); err != nil {
		return err
	}

	return nil
}

func (obj *Country) DatabaseFields() string {
	return msql.FormatFields(
		"index", "continent_index",
		"uuid", "name",
		"details", "creator",
		"created", "updated", "deleted_state",
	)
}
