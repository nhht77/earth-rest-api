package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/nhht77/earth-rest-api/mhttp"
	"github.com/nhht77/earth-rest-api/muuid"
	pkg_v1 "github.com/nhht77/earth-rest-api/server/pkg"
)

func HandleContinents(w http.ResponseWriter, r *http.Request) {

	options, err := ContinentOptionsFromQuery(r)
	if err != nil {
		mhttp.WriteBadRequest(w, fmt.Sprintf("Invalid query: %s", err.Error()))
		return
	}

	results, err := DB.ContinentsByOptions(options)
	if err != nil {
		mhttp.WriteBadRequest(w, err.Error())
		return
	}

	mhttp.WriteBodyJSON(w, results)
}

func HandleContinent(w http.ResponseWriter, r *http.Request) {

	c_uuid := mhttp.Query(r, "uuid")
	if _, err := muuid.UUIDFromString(c_uuid); err != nil {
		mhttp.WriteBadRequest(w, err.Error())
		return
	}

	result, err := DB.ContinentByUuid(nil, c_uuid)
	if err != nil {
		mhttp.WriteBadRequest(w, err.Error())
		return
	}

	mhttp.WriteBodyJSON(w, result)
}

func HandleCreateContinent(w http.ResponseWriter, r *http.Request) {

	continent := &pkg_v1.Continent{}

	if err := mhttp.ReadBodyJSON(r, &continent); err != nil {
		mhttp.WriteBadRequest(w, err.Error())
		return
	}

	if err := continent.ValidateCreate(); err != nil {
		mhttp.WriteBadRequest(w, err.Error())
		return
	}

	result, err := DB.CreateContinent(nil, continent)
	if err != nil {
		mhttp.WriteBadRequest(w, err.Error())
		return
	}

	mhttp.WriteBodyJSON(w, result)
}

func HandleUpdateContinent(w http.ResponseWriter, r *http.Request) {

	continent := &pkg_v1.Continent{}
	if err := mhttp.ReadBodyJSON(r, &continent); err != nil {
		mhttp.WriteBadRequest(w, err.Error())
		return
	}
	Log.Info("HandleUpdateContinent #1", continent.Uuid)

	if !muuid.UUIDValid(continent.Uuid) {
		mhttp.WriteBadRequest(w, errors.New("Invalid uuid").Error())
		return
	}
	Log.Info("HandleUpdateContinent #2")

	if err := continent.ValidateUpdate(); err != nil {
		mhttp.WriteBadRequest(w, err.Error())
		return
	}

	Log.Info("HandleUpdateContinent #3")

	result, err := DB.UpdateContinent(nil, continent)
	if err != nil {
		mhttp.WriteBadRequest(w, err.Error())
		return
	}

	mhttp.WriteBodyJSON(w, result)
}

func HandleDeleteHub(w http.ResponseWriter, r *http.Request) {
	var query_uuid = mhttp.Query(r, "uuid")

	if _, err := muuid.UUIDFromString(query_uuid); err != nil {
		mhttp.WriteBadRequest(w, err.Error())
		return
	}

	if err := DB.SoftDeleteHub(nil, query_uuid); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Note: return 200
	mhttp.WriteBodyJSON(w, "")
}
