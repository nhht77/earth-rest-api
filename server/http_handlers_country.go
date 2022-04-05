package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/nhht77/earth-rest-api/mhttp"
	"github.com/nhht77/earth-rest-api/muuid"
	pkg_v1 "github.com/nhht77/earth-rest-api/server/pkg"
)

func HandleCountries(w http.ResponseWriter, r *http.Request) {

	options, err := CountryOptionsFromQuery(r)
	if err != nil {
		mhttp.WriteBadRequest(w, fmt.Sprintf("Invalid query: %s", err.Error()))
		return
	}

	results, err := DB.CountriesByOptions(options)
	if err != nil {
		mhttp.WriteBadRequest(w, err.Error())
		return
	}

	mhttp.WriteBodyJSON(w, results)
}

func HandleCountry(w http.ResponseWriter, r *http.Request) {

	c_uuid := mhttp.Query(r, "uuid")
	if _, err := muuid.UUIDFromString(c_uuid); err != nil {
		mhttp.WriteBadRequest(w, err.Error())
		return
	}

	result, err := DB.CountryByUuid(nil, c_uuid)
	if err != nil {
		mhttp.WriteBadRequest(w, err.Error())
		return
	}

	mhttp.WriteBodyJSON(w, result)
}

func HandleCreateCountry(w http.ResponseWriter, r *http.Request) {

	continent := &pkg_v1.Country{}

	if err := mhttp.ReadBodyJSON(r, &continent); err != nil {
		mhttp.WriteBadRequest(w, err.Error())
		return
	}

	if err := continent.ValidateCreate(); err != nil {
		mhttp.WriteBadRequest(w, err.Error())
		return
	}

	result, err := DB.CreateCountry(nil, continent)
	if err != nil {
		mhttp.WriteBadRequest(w, err.Error())
		return
	}

	mhttp.WriteBodyJSON(w, result)
}

func HandleUpdateCountry(w http.ResponseWriter, r *http.Request) {

	continent := &pkg_v1.Country{}
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

	result, err := DB.UpdateCountry(nil, continent)
	if err != nil {
		mhttp.WriteBadRequest(w, err.Error())
		return
	}

	mhttp.WriteBodyJSON(w, result)
}

func HandleDeleteCountry(w http.ResponseWriter, r *http.Request) {
	var query_uuid = mhttp.Query(r, "uuid")

	if _, err := muuid.UUIDFromString(query_uuid); err != nil {
		mhttp.WriteBadRequest(w, err.Error())
		return
	}

	if err := DB.SoftDeleteCountry(nil, query_uuid); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Note: return 200
	mhttp.WriteBodyJSON(w, "")
}
