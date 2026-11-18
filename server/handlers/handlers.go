package handlers

import (
	"net/http"

	"github.com/nhht77/earth-rest-api/pkg/logging"
	"github.com/nhht77/earth-rest-api/server/database"
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

// requestLog returns the request-scoped logger (carrying request_id),
// falling back to the app-level logger when the context carries none.
func (h *Handlers) requestLog(r *http.Request) logging.Logger {
	if l := logging.FromContext(r.Context()); l != nil {
		return l
	}
	return h.log
}
