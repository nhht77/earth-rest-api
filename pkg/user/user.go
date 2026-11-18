// Package user holds the shared minimal user identity used across
// the API wire models (creator/auditor fields).
package user

import (
	"database/sql/driver"
	"errors"

	"github.com/nhht77/earth-rest-api/pkg/msql"
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
