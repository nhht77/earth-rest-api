package pkg_v1

import (
	"database/sql/driver"
	"errors"
	"time"

	"github.com/nhht77/earth-rest-api/server/pkg/msql"
	muuid "github.com/nhht77/earth-rest-api/server/pkg/muuid"
)

type UserMinimal struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}

func (user *UserMinimal) IsValid() error {
	if user == nil {
		return errors.New("Invalid user")
	}

	if len(user.Email) == 0 {
		return errors.New("Invalid email")
	}
	if len(user.Name) == 0 {
		return errors.New("Invalid name")
	}
	return nil
}

func (v *UserMinimal) Value() (driver.Value, error) {
	return msql.JSONValue(v)
}

func (v *UserMinimal) Scan(src interface{}) error {
	if src != nil && v == nil {
		v = &UserMinimal{}
	}
	return msql.JSONScan(src, v)
}

////////////////////////
/////// Continent struct

type ContinentType int

const (
	ContinentType_Invalid       ContinentType = 0
	ContinentType_Asia          ContinentType = 1
	ContinentType_Africa        ContinentType = 2
	ContinentType_Europe        ContinentType = 3
	ContinentType_North_America ContinentType = 4
	ContinentType_South_America ContinentType = 5
	ContinentType_Oceania       ContinentType = 6
	ContinentType_Antarctica    ContinentType = 7
)

type Continent struct {
	Index msql.DatabaseIndex `json:"-"`
	Uuid  muuid.UUID         `json:"uuid"`

	Name      string        `json:"name"`
	Type      ContinentType `json:"type"`
	AreaByKm2 float64       `json:"area_by_km2"`

	Created time.Time `json:"created"`
	Updated time.Time `json:"updated"`

	Creator *UserMinimal `json:"creator"`

	DeletedState msql.DeletedState `json:"-"`
}

func (obj *Continent) ValidateCreate() error {

	// check type
	if err := obj.IsValidContinentType(obj.Type); err != nil {
		return err
	}

	if obj.Creator == nil {
		return errors.New("Empty continent creator")
	}

	// check name
	if len(obj.Name) == 0 {
		return errors.New("Invalid continent name")
	}

	// check areaByKm2
	if obj.AreaByKm2 == 0 {
		return errors.New("Invalid continent area by km2")
	}

	// Check for creator
	if err := obj.Creator.IsValid(); err != nil {
		return err
	}

	return nil
}

func (obj *Continent) ValidateUpdate() error {

	if !muuid.UUIDValid(obj.Uuid) {
		return errors.New("Invalid continent uuid")
	}

	// check type
	if err := obj.IsValidContinentType(obj.Type); err != nil {
		return err
	}

	// check name
	if len(obj.Name) == 0 {
		return errors.New("Invalid continent name")
	}

	// check areaByKm2
	if obj.AreaByKm2 == 0 {
		return errors.New("Invalid continent area by km2")
	}

	return nil
}

func (obj *Continent) IsValidContinentType(con_type ContinentType) error {

	if con_type == ContinentType_Invalid {
		return errors.New("Invalid continent type")
	}

	types := []ContinentType{
		ContinentType_Asia,
		ContinentType_Africa,
		ContinentType_Europe,
		ContinentType_North_America,
		ContinentType_South_America,
		ContinentType_Oceania,
		ContinentType_Antarctica,
	}

	for _, iter := range types {
		if con_type == iter {
			return nil
		}
	}

	return errors.New("Unsupported continent type")
}

func (obj *Continent) DatabaseFields() string {
	return msql.FormatFields(
		"index", "uuid",
		"name", "type",
		"area_by_km2",
		"creator",
		"created", "updated", "deleted_state",
	)
}
