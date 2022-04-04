package continent

import (
	"database/sql/driver"

	"github.com/nhht77/earth-rest-api/msql"
	muuid "github.com/nhht77/earth-rest-api/muuid"
	"google.golang.org/protobuf/types/known/timestamppb"
)

////////////////////////////
/////// Basic schema struct

type DatabaseIndex uint

type DB_OBJECT struct {
	DB_INDEX DatabaseIndex `json:"-" yaml:"-"`
}

type DeletedState int

const (
	NotDeleted  DeletedState = 0
	SoftDeleted DeletedState = 1
)

type UserMinimal struct {
	Email string `json:"email"`
	Name  string `json:"name"`
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
	ContinentType_Asia          ContinentType = 1
	ContinentType_Africa        ContinentType = 2
	ContinentType_Europe        ContinentType = 3
	ContinentType_North_America ContinentType = 4
	ContinentType_South_America ContinentType = 5
	ContinentType_Oceania       ContinentType = 6
	ContinentType_Antarctica    ContinentType = 7
)

type Continent struct {
	Index DatabaseIndex `json:"-" yaml:"-"`
	Uuid  muuid.UUID    `json:"uuid"`

	Name      string        `json:"string"`
	Type      ContinentType `json:"type"`
	AreaByKm2 int           `area_by_km2`

	Created *timestamppb.Timestamp `json:"created"`
	Updated *timestamppb.Timestamp `json:"updated"`

	Creator *UserMinimal `json:"creator"`

	DeletedState DeletedState `json:"-"`
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
