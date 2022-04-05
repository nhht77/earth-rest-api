package muuid

import (
	"fmt"

	_uuid "gopkg.in/satori/go.uuid.v1"
)

// UUID
type UUID = _uuid.UUID
type NullUUID = _uuid.NullUUID

// Create new random v4 UUID.
func NewUUID() UUID {
	return _uuid.NewV4()
}

func UUIDValid(uuid UUID) bool {
	return UUIDNil(uuid) == false
}

func UUIDNil(uuid UUID) bool {
	return uuid == _uuid.Nil
}

func UUIDFromString(uuid string) (UUID, error) {
	if len(uuid) == 0 {
		return _uuid.Nil, fmt.Errorf("empty UUID")
	}
	result, err := _uuid.FromString(uuid)
	if err != nil {
		return _uuid.Nil, err
	}
	return result, nil
}
