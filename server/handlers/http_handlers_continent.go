package handlers

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/nhht77/earth-rest-api/pkg/mhttp"
	"github.com/nhht77/earth-rest-api/pkg/muuid"
	pkg_v1 "github.com/nhht77/earth-rest-api/pkg/v1"
	"github.com/nhht77/earth-rest-api/server/database"
)

func (h *Handlers) HandleContinents(w http.ResponseWriter, r *http.Request) {

	options, err := database.ContinentOptionsFromQuery(r)
	if err != nil {
		mhttp.WriteBadRequest(w, fmt.Sprintf("Invalid query: %s", err.Error()))
		return
	}

	results, err := h.db.ContinentsByOptions(options)
	if err != nil {
		// our fault → 500 (was wrongly 400: a DB failure is never the caller's fault)
		h.requestLog(r).Errorf("[continent] DB.ContinentsByOptions error: %s", err.Error())
		mhttp.WriteInternalServerError(w, err.Error())
		return
	}

	mhttp.WriteBodyJSON(w, results)
}

func (h *Handlers) HandleContinent(w http.ResponseWriter, r *http.Request) {

	c_uuid := mhttp.Query(r, "uuid")
	if _, err := muuid.UUIDFromString(c_uuid); err != nil {
		mhttp.WriteBadRequest(w, err.Error())
		return
	}

	result, err := h.db.ContinentByUuid(nil, c_uuid)
	if err != nil {
		h.requestLog(r).Errorf("[continent] DB.ContinentByUuid error: %s", err.Error())
		mhttp.WriteInternalServerError(w, err.Error())
		return
	}

	mhttp.WriteBodyJSON(w, result)
}

func (h *Handlers) HandleCreateContinent(w http.ResponseWriter, r *http.Request) {

	continent := &pkg_v1.Continent{}

	if err := mhttp.ReadBodyJSON(r, &continent); err != nil {
		mhttp.WriteBadRequest(w, err.Error())
		return
	}

	if err := continent.ValidateCreate(); err != nil {
		mhttp.WriteBadRequest(w, err.Error())
		return
	}

	result, err := h.db.CreateContinent(nil, continent)
	if err != nil {
		h.requestLog(r).Errorf("[continent] DB.CreateContinent error: %s", err.Error())
		mhttp.WriteInternalServerError(w, err.Error())
		return
	}

	mhttp.WriteBodyJSON(w, result)
}

func (h *Handlers) HandleUpdateContinent(w http.ResponseWriter, r *http.Request) {

	continent := &pkg_v1.Continent{}
	if err := mhttp.ReadBodyJSON(r, &continent); err != nil {
		mhttp.WriteBadRequest(w, err.Error())
		return
	}
	if !muuid.UUIDValid(continent.Uuid) {
		mhttp.WriteBadRequest(w, errors.New("Invalid uuid").Error())
		return
	}
	if err := continent.ValidateUpdate(); err != nil {
		mhttp.WriteBadRequest(w, err.Error())
		return
	}

	result, err := h.db.UpdateContinent(nil, continent)
	if err != nil {
		h.requestLog(r).Errorf("[continent] DB.UpdateContinent error: %s", err.Error())
		mhttp.WriteInternalServerError(w, err.Error())
		return
	}

	mhttp.WriteBodyJSON(w, result)
}

func (h *Handlers) HandleDeleteContinent(w http.ResponseWriter, r *http.Request) {
	var query_uuid = mhttp.Query(r, "uuid")

	if _, err := muuid.UUIDFromString(query_uuid); err != nil {
		mhttp.WriteBadRequest(w, err.Error())
		return
	}

	if err := h.db.SoftDeleteContinent(nil, query_uuid); err != nil {
		h.requestLog(r).Errorf("[continent] DB.SoftDeleteContinent error: %s", err.Error())
		mhttp.WriteInternalServerError(w, err.Error())
		return
	}

	// Note: return 200
	mhttp.WriteBodyJSON(w, "")
}
