package handlers

import (
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/nhht77/earth-rest-api/pkg/handlers"
	"github.com/nhht77/earth-rest-api/pkg/logging"
	"github.com/nhht77/earth-rest-api/pkg/muuid"
	"github.com/nhht77/earth-rest-api/server/config"
)

func RunHTTP(appConf *config.Config, h *Handlers) error {
	router := mux.NewRouter().StrictSlash(true)
	router.Use(RequestMiddleware(h.log))
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

// statusWriter captures the response status code, which net/http does not expose.
type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

// MonitorMiddleware logs one line per request: method, path, status, duration.
// The request-scoped logger from the context carries the request_id; it falls
// back to the injected log when RequestMiddleware did not run (or tests).
func MonitorMiddleware(log logging.Logger) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestLog := logging.FromContext(r.Context())
			if requestLog == nil {
				requestLog = log
			}

			started := time.Now()
			sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}

			next.ServeHTTP(sw, r)

			requestLog.Infof("[http] %s %s -> %d in %s",
				r.Method, r.URL.Path, sw.status, time.Since(started))
		})
	}
}

func RequestMiddleware(log logging.Logger) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestID := muuid.NewUUID().String()
			requestLog := log.With("request_id", requestID)

			// expose the ID to clients for support/correlation
			w.Header().Set("X-Request-ID", requestID)

			next.ServeHTTP(w, r.WithContext(
				logging.NewContext(r.Context(), requestLog),
			))
		})
	}
}

func Ping(w http.ResponseWriter, r *http.Request) {
	handlers.WriteBodyJSON(w, "")
}
