package main

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/nhht77/earth-rest-api/server/config"
	"github.com/nhht77/earth-rest-api/server/pkg/logging"
	"github.com/nhht77/earth-rest-api/server/pkg/mhttp"
)

func RunHTTP(appConf *config.Config, h *Handlers) error {
	router := mux.NewRouter().StrictSlash(true)
	router.Use(MonitorMiddleware(h.log))

	router.HandleFunc("/api/v1/ping", Ping).Methods("GET")

	router.HandleFunc("/api/v1/continents", h.HandleContinents).Methods("GET")
	router.HandleFunc("/api/v1/continent", h.HandleContinent).Methods("GET")
	router.HandleFunc("/api/v1/continent/create", h.HandleCreateContinent).Methods("POST")
	router.HandleFunc("/api/v1/continent/update", h.HandleUpdateContinent).Methods("PUT")
	router.HandleFunc("/api/v1/continent/delete", h.HandleDeleteContinent).Methods("DELETE")

	router.HandleFunc("/api/v1/countries", h.HandleCountries).Methods("GET")
	router.HandleFunc("/api/v1/country", h.HandleCountry).Methods("GET")
	router.HandleFunc("/api/v1/country/create", h.HandleCreateCountry).Methods("POST")
	router.HandleFunc("/api/v1/country/update", h.HandleUpdateCountry).Methods("PUT")
	router.HandleFunc("/api/v1/country/delete", h.HandleDeleteCountry).Methods("DELETE")

	router.HandleFunc("/api/v1/cities", h.HandleCities).Methods("GET")
	router.HandleFunc("/api/v1/city", h.HandleCity).Methods("GET")
	router.HandleFunc("/api/v1/city/create", h.HandleCreateCity).Methods("POST")
	router.HandleFunc("/api/v1/city/update", h.HandleUpdateCity).Methods("PUT")
	router.HandleFunc("/api/v1/city/delete", h.HandleDeleteCity).Methods("DELETE")

	return ListenAndServe(":"+appConf.Framework.ServerPort, h, router)
}

func ListenAndServe(addr string, h *Handlers, handler http.Handler) error {
	h.log.Infof("[http] Listen and serve at port %s", addr[1:])
	return http.ListenAndServe(addr, handler)
}

func MonitorMiddleware(log logging.Logger) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log.Infof("[http] %s %s", r.Method, r.URL)
			next.ServeHTTP(w, r)
		})
	}
}

func Ping(w http.ResponseWriter, r *http.Request) {
	mhttp.WriteBodyJSON(w, "")
}
