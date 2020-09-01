package accounts

import (
	"database/sql"
	"encoding/json"
	"net/http"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

func jsonResponse(w http.ResponseWriter, value interface{}) {
	jsonResponseStatus(w, http.StatusOK, value)
}

func jsonResponseStatus(w http.ResponseWriter, status int, value interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(status)
	e := json.NewEncoder(w)
	e.SetIndent("", "  ")
	e.Encode(value)
}

func errorResponse(w http.ResponseWriter, err error) {
	switch err.(type) {
	case *json.InvalidUnmarshalError:
		w.WriteHeader(400)
	case validation.Errors:
		jsonResponseStatus(w, http.StatusBadRequest, err)
	default:
		switch err {
		case sql.ErrNoRows:
			w.WriteHeader(404)
		default:
			w.WriteHeader(500)
		}
	}
}
