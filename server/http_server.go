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
	mhttp.WriteBodyJSON(w, "empty string")
}
