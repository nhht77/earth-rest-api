package mhttp

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
)

func ReadBodyJSON(r *http.Request, dest interface{}) error {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}
	return json.Unmarshal(body, dest)
}

func WriteJSON(w http.ResponseWriter, statusCode int, data interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	return json.NewEncoder(w).Encode(data)
}

func WriteBodyJSON(w http.ResponseWriter, data interface{}) error {
	return WriteJSON(w, http.StatusOK, data)
}

func WriteBadRequest(w http.ResponseWriter, http_err interface{}) error {
	return WriteJSON(w, http.StatusBadRequest, http_err)
}

func WriteInternalServerError(w http.ResponseWriter, http_err interface{}) error {
	return WriteJSON(w, http.StatusInternalServerError, http_err)
}

func Query(r *http.Request, query string) string {
	if r == nil {
		return ""
	}
	return r.URL.Query().Get(query)
}

func QueryList(r *http.Request, key string, separator string) []string {
	if r == nil {
		return []string{}
	}
	str := r.URL.Query().Get(key)
	if len(str) > 0 && strings.Contains(str, separator) {
		out := strings.Split(str, separator)
		for i, _ := range out {
			out[i] = strings.TrimSpace(out[i])
			return out
		}
	}
	return []string{}
}

func QueryIntList(r *http.Request, key string, separator string) ([]int, error) {
	out := []int{}
	values := QueryList(r, key, separator)
	for _, v := range values {
		num, err := strconv.Atoi(v)
		if err != nil {
			return []int{}, err
		}
		out = append(out, num)
	}
	return out, nil
}

func QueryBool(r *http.Request, key string) bool {
	str := strings.ToLower(r.URL.Query().Get(key))
	return str == "true" || str == "1"
}

func QueryBoolDefault(r *http.Request, key string, default_value bool) bool {
	str := strings.ToLower(r.URL.Query().Get(key))
	if len(str) > 0 {
		return str == "true" || str == "1"
	}
	return default_value
}
