package handlers

import (
	"github.com/nhht77/earth-rest-api/server/database"
	"github.com/nhht77/earth-rest-api/server/pkg/logging"
)

type Handlers struct {
	db  *database.Database
	log logging.Logger
}

func NewHandlers(db *database.Database, log logging.Logger) *Handlers {
	return &Handlers{
		db:  db,
		log: log,
	}
}
