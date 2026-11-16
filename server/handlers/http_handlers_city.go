package handlers

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/nhht77/earth-rest-api/server/database"
	pkg_v1 "github.com/nhht77/earth-rest-api/server/pkg"
	"github.com/nhht77/earth-rest-api/server/pkg/mhttp"
	"github.com/nhht77/earth-rest-api/server/pkg/muuid"
)

func (h *Handlers) HandleCities(w http.ResponseWriter, r *http.Request) {

	options, err := database.CityOptionsFromQuery(r)
	if err != nil {
		mhttp.WriteBadRequest(w, fmt.Sprintf("Invalid query: %s", err.Error()))
		return
	}

	results, err := h.db.CitiesByOptions(options)
	if err != nil {
		mhttp.WriteBadRequest(w, err.Error())
		return
	}

	mhttp.WriteBodyJSON(w, results)
}

func (h *Handlers) HandleCity(w http.ResponseWriter, r *http.Request) {

	c_uuid := mhttp.Query(r, "uuid")
	if _, err := muuid.UUIDFromString(c_uuid); err != nil {
		mhttp.WriteBadRequest(w, err.Error())
		return
	}

	result, err := h.db.CityByUuid(nil, c_uuid)
	if err != nil {
		mhttp.WriteBadRequest(w, err.Error())
		return
	}

	mhttp.WriteBodyJSON(w, result)
}

func (h *Handlers) HandleCreateCity(w http.ResponseWriter, r *http.Request) {

	continent := &pkg_v1.City{}

	if err := mhttp.ReadBodyJSON(r, &continent); err != nil {
		mhttp.WriteBadRequest(w, err.Error())
		return
	}

	if err := continent.ValidateCreate(); err != nil {
		mhttp.WriteBadRequest(w, err.Error())
		return
	}

	result, err := h.db.CreateCity(nil, continent)
	if err != nil {
		mhttp.WriteBadRequest(w, err.Error())
		return
	}

	mhttp.WriteBodyJSON(w, result)
}

func (h *Handlers) HandleUpdateCity(w http.ResponseWriter, r *http.Request) {

	continent := &pkg_v1.City{}
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

	result, err := h.db.UpdateCity(nil, continent)
	if err != nil {
		mhttp.WriteBadRequest(w, err.Error())
		return
	}

	mhttp.WriteBodyJSON(w, result)
}

func (h *Handlers) HandleDeleteCity(w http.ResponseWriter, r *http.Request) {
	var query_uuid = mhttp.Query(r, "uuid")

	if _, err := muuid.UUIDFromString(query_uuid); err != nil {
		mhttp.WriteBadRequest(w, err.Error())
		return
	}

	if err := h.db.SoftDeleteCity(nil, query_uuid); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Note: return 200
	mhttp.WriteBodyJSON(w, "")
}
