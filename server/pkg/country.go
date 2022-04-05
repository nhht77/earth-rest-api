package pkg_v1

import (
	"errors"
	"time"

	"github.com/nhht77/earth-rest-api/msql"
	muuid "github.com/nhht77/earth-rest-api/muuid"
)

type Country struct {
	Index          msql.DatabaseIndex `json:"-"`
	ContinentIndex msql.DatabaseIndex `json:"-"`

	CountryUuid muuid.UUID `json:"country_uuid"`
	Uuid        muuid.UUID `json:"uuid"`

	Name      string `json:"name"`
	PhoneCode string `json:"phone_code"`
	ISOCode   string `json:"iso_code"`
	Currency  string `json:"currency"`

	Created time.Time `json:"created"`
	Updated time.Time `json:"updated"`

	Creator *UserMinimal `json:"creator"`

	DeletedState msql.DeletedState `json:"-"`
}

func (obj *Country) ValidateCreate() error {

	if len(obj.Name) == 0 {
		return errors.New("Invalid country name")
	}

	if len(obj.PhoneCode) == 0 {
		return errors.New("Invalid country phone code")
	}

	if len(obj.ISOCode) == 0 {
		return errors.New("Invalid country iso code")
	}

	if len(obj.Currency) == 0 {
		return errors.New("Invalid country currency")
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

	if len(obj.PhoneCode) == 0 {
		return errors.New("Invalid country phone code")
	}

	if len(obj.ISOCode) == 0 {
		return errors.New("Invalid country iso code")
	}

	if len(obj.Currency) == 0 {
		return errors.New("Invalid country currency")
	}
	return nil
}

func (obj *Country) DatabaseFields() string {
	return msql.FormatFields(
		"index", "continent_index",
		"uuid", "name",
		"phone_code", "iso_code",
		"currency", "creator",
		"created", "updated", "deleted_state",
	)
}
