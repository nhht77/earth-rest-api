package main

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/nhht77/earth-rest-api/server/pkg/mhttp"
)

func RunHTTP() error {
	router := mux.NewRouter().StrictSlash(true)
	router.Use(MonitorHandle)

	router.HandleFunc("/api/v1/ping", Ping).Methods("GET")

	router.HandleFunc("/api/v1/continents", HandleContinents).Methods("GET")
	router.HandleFunc("/api/v1/continent", HandleContinent).Methods("GET")
	router.HandleFunc("/api/v1/continent/create", HandleCreateContinent).Methods("POST")
	router.HandleFunc("/api/v1/continent/update", HandleUpdateContinent).Methods("PUT")
	router.HandleFunc("/api/v1/continent/delete", HandleDeleteContinent).Methods("DELETE")

	router.HandleFunc("/api/v1/countries", HandleCountries).Methods("GET")
	router.HandleFunc("/api/v1/country", HandleCountry).Methods("GET")
	router.HandleFunc("/api/v1/country/create", HandleCreateCountry).Methods("POST")
	router.HandleFunc("/api/v1/country/update", HandleUpdateCountry).Methods("PUT")
	router.HandleFunc("/api/v1/country/delete", HandleDeleteCountry).Methods("DELETE")

	router.HandleFunc("/api/v1/cities", HandleCities).Methods("GET")
	router.HandleFunc("/api/v1/city", HandleCity).Methods("GET")
	router.HandleFunc("/api/v1/city/create", HandleCreateCity).Methods("POST")
	router.HandleFunc("/api/v1/city/update", HandleUpdateCity).Methods("PUT")
	router.HandleFunc("/api/v1/city/delete", HandleDeleteCity).Methods("DELETE")

	if AppConfig.Framework.IsTestBuild {
		router.HandleFunc("/api/v1/__clear-test-table", HandleClearTestTable).Methods("DELETE")
	}

	return ListenAndServe(":8080", router)
}

func ListenAndServe(addr string, handler http.Handler) error {
	Log.Infof("[http] Listen and serve at port %s", addr[1:])
	return http.ListenAndServe(addr, handler)
}

func MonitorHandle(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		Log.Infof("[http] %s %s", r.Method, r.URL)
		h.ServeHTTP(w, r)
	})
}

func Ping(w http.ResponseWriter, r *http.Request) {
	mhttp.WriteBodyJSON(w, "")
}

func HandleClearTestTable(w http.ResponseWriter, r *http.Request) {

	if err := DB._ClearTable(); err != nil {
		mhttp.WriteInternalServerError(w, err.Error())
	}

	mhttp.WriteBodyJSON(w, "")
}
