package handlers

import (
	"errors"
	"fmt"
	"net/http"

	country "github.com/nhht77/earth-rest-api/pkg/country"
	"github.com/nhht77/earth-rest-api/pkg/handlers"
	"github.com/nhht77/earth-rest-api/pkg/muuid"
	"github.com/nhht77/earth-rest-api/server/database"
)

func (h *Handlers) HandleCountries(w http.ResponseWriter, r *http.Request) {

	options, err := database.CountryOptionsFromQuery(r)
	if err != nil {
		handlers.WriteBadRequest(w, fmt.Sprintf("Invalid query: %s", err.Error()))
		return
	}

	results, err := h.db.CountriesByOptions(options)
	if err != nil {
		h.requestLog(r).Errorf("[country] DB.CountriesByOptions error: %s", err.Error())
		handlers.WriteInternalServerError(w, err.Error())
		return
	}

	handlers.WriteBodyJSON(w, results)
}

func (h *Handlers) HandleCountry(w http.ResponseWriter, r *http.Request) {

	c_uuid := handlers.Query(r, "uuid")
	if _, err := muuid.UUIDFromString(c_uuid); err != nil {
		handlers.WriteBadRequest(w, err.Error())
		return
	}

	result, err := h.db.CountryByUuid(nil, c_uuid)
	if err != nil {
		h.requestLog(r).Errorf("[country] DB.CountryByUuid error: %s", err.Error())
		handlers.WriteInternalServerError(w, err.Error())
		return
	}

	handlers.WriteBodyJSON(w, result)
}

func (h *Handlers) HandleCreateCountry(w http.ResponseWriter, r *http.Request) {

	country := &country.Country{}

	if err := handlers.ReadBodyJSON(r, &country); err != nil {
		handlers.WriteBadRequest(w, err.Error())
		return
	}

	if err := country.ValidateCreate(); err != nil {
		handlers.WriteBadRequest(w, err.Error())
		return
	}

	result, err := h.db.CreateCountry(nil, country)
	if err != nil {
		h.requestLog(r).Errorf("[country] DB.CreateCountry error: %s", err.Error())
		handlers.WriteInternalServerError(w, err.Error())
		return
	}

	handlers.WriteBodyJSON(w, result)
}

func (h *Handlers) HandleUpdateCountry(w http.ResponseWriter, r *http.Request) {

	continent := &country.Country{}
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

	result, err := h.db.UpdateCountry(nil, continent)
	if err != nil {
		h.requestLog(r).Errorf("[country] DB.UpdateCountry error: %s", err.Error())
		handlers.WriteInternalServerError(w, err.Error())
		return
	}

	handlers.WriteBodyJSON(w, result)
}

func (h *Handlers) HandleDeleteCountry(w http.ResponseWriter, r *http.Request) {
	var query_uuid = handlers.Query(r, "uuid")

	if _, err := muuid.UUIDFromString(query_uuid); err != nil {
		handlers.WriteBadRequest(w, err.Error())
		return
	}

	if err := h.db.SoftDeleteCountry(nil, query_uuid); err != nil {
		h.requestLog(r).Errorf("[country] DB.SoftDeleteCountry error: %s", err.Error())
		handlers.WriteInternalServerError(w, err.Error())
		return
	}

	// Note: return 200
	handlers.WriteBodyJSON(w, "")
}
