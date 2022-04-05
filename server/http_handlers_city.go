package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/nhht77/earth-rest-api/mhttp"
	"github.com/nhht77/earth-rest-api/muuid"
	pkg_v1 "github.com/nhht77/earth-rest-api/server/pkg"
)

func HandleCities(w http.ResponseWriter, r *http.Request) {

	options, err := CityOptionsFromQuery(r)
	if err != nil {
		mhttp.WriteBadRequest(w, fmt.Sprintf("Invalid query: %s", err.Error()))
		return
	}

	results, err := DB.CitiesByOptions(options)
	if err != nil {
		mhttp.WriteBadRequest(w, err.Error())
		return
	}

	mhttp.WriteBodyJSON(w, results)
}

func HandleCity(w http.ResponseWriter, r *http.Request) {

	c_uuid := mhttp.Query(r, "uuid")
	if _, err := muuid.UUIDFromString(c_uuid); err != nil {
		mhttp.WriteBadRequest(w, err.Error())
		return
	}

	result, err := DB.CityByUuid(nil, c_uuid)
	if err != nil {
		mhttp.WriteBadRequest(w, err.Error())
		return
	}

	mhttp.WriteBodyJSON(w, result)
}

func HandleCreateCity(w http.ResponseWriter, r *http.Request) {

	continent := &pkg_v1.City{}

	if err := mhttp.ReadBodyJSON(r, &continent); err != nil {
		mhttp.WriteBadRequest(w, err.Error())
		return
	}

	if err := continent.ValidateCreate(); err != nil {
		mhttp.WriteBadRequest(w, err.Error())
		return
	}

	result, err := DB.CreateCity(nil, continent)
	if err != nil {
		mhttp.WriteBadRequest(w, err.Error())
		return
	}

	mhttp.WriteBodyJSON(w, result)
}

func HandleUpdateCity(w http.ResponseWriter, r *http.Request) {

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

	result, err := DB.UpdateCity(nil, continent)
	if err != nil {
		mhttp.WriteBadRequest(w, err.Error())
		return
	}

	mhttp.WriteBodyJSON(w, result)
}

func HandleDeleteCity(w http.ResponseWriter, r *http.Request) {
	var query_uuid = mhttp.Query(r, "uuid")

	if _, err := muuid.UUIDFromString(query_uuid); err != nil {
		mhttp.WriteBadRequest(w, err.Error())
		return
	}

	if err := DB.SoftDeleteCity(nil, query_uuid); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Note: return 200
	mhttp.WriteBodyJSON(w, "")
}
