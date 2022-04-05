package main

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/nhht77/earth-rest-api/mhttp"
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
