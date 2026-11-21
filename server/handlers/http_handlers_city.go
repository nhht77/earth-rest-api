package handlers

import (
	"errors"
	"fmt"
	"net/http"

	city "github.com/nhht77/earth-rest-api/pkg/city"
	"github.com/nhht77/earth-rest-api/pkg/handlers"
	"github.com/nhht77/earth-rest-api/pkg/muuid"
	"github.com/nhht77/earth-rest-api/server/database"
)

func (h *Handlers) HandleCities(w http.ResponseWriter, r *http.Request) {

	options, err := database.CityOptionsFromQuery(r)
	if err != nil {
		handlers.WriteBadRequest(w, fmt.Sprintf("Invalid query: %s", err.Error()))
		return
	}

	results, err := h.db.CitiesByOptions(options)
	if err != nil {
		h.requestLog(r).Errorf("[city] DB.CitiesByOptions error: %s", err.Error())
		handlers.WriteInternalServerError(w, err.Error())
		return
	}

	handlers.WriteBodyJSON(w, results)
}

func (h *Handlers) HandleCity(w http.ResponseWriter, r *http.Request) {

	c_uuid := handlers.Query(r, "uuid")
	if _, err := muuid.UUIDFromString(c_uuid); err != nil {
		handlers.WriteBadRequest(w, err.Error())
		return
	}

	result, err := h.db.CityByUuid(nil, c_uuid)
	if err != nil {
		h.requestLog(r).Errorf("[city] DB.CityByUuid error: %s", err.Error())
		handlers.WriteInternalServerError(w, err.Error())
		return
	}

	handlers.WriteBodyJSON(w, result)
}

func (h *Handlers) HandleCreateCity(w http.ResponseWriter, r *http.Request) {

	continent := &city.City{}

	if err := handlers.ReadBodyJSON(r, &continent); err != nil {
		handlers.WriteBadRequest(w, err.Error())
		return
	}

	if err := continent.ValidateCreate(); err != nil {
		// caller's fault → 400 (validation is request correctness, not an outage)
		handlers.WriteBadRequest(w, err.Error())
		return
	}

	result, err := h.db.CreateCity(nil, continent)
	if err != nil {
		// our fault → 500, logged
		h.requestLog(r).Errorf("[city] DB.CreateCity error: %s", err.Error())
		handlers.WriteInternalServerError(w, err.Error())
		return
	}

	handlers.WriteBodyJSON(w, result)
}

func (h *Handlers) HandleUpdateCity(w http.ResponseWriter, r *http.Request) {

	continent := &city.City{}
	if err := handlers.ReadBodyJSON(r, &continent); err != nil {
		handlers.WriteBadRequest(w, err.Error())
		return
	}
	if !muuid.UUIDValid(continent.Uuid) {
		handlers.WriteBadRequest(w, errors.New("Invalid uuid").Error())
		return
	}
	if err := continent.ValidateUpdate(); err != nil {
		handlers.WriteBadRequest(w, err.Error())
		return
	}

	result, err := h.db.UpdateCity(nil, continent)
	if err != nil {
		h.requestLog(r).Errorf("[city] DB.UpdateCity error: %s", err.Error())
		handlers.WriteInternalServerError(w, err.Error())
		return
	}

	handlers.WriteBodyJSON(w, result)
}

func (h *Handlers) HandleDeleteCity(w http.ResponseWriter, r *http.Request) {
	var query_uuid = handlers.Query(r, "uuid")

	if _, err := muuid.UUIDFromString(query_uuid); err != nil {
		handlers.WriteBadRequest(w, err.Error())
		return
	}

	if err := h.db.SoftDeleteCity(nil, query_uuid); err != nil {
		h.requestLog(r).Errorf("[city] DB.SoftDeleteCity error: %s", err.Error())
		handlers.WriteInternalServerError(w, err.Error())
		return
	}

	// Note: return 200
	handlers.WriteBodyJSON(w, "")
}
