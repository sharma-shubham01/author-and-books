// Package rest holds the request and response helpers shared by the HTTP handlers.
package rest

import (
	"authorAndBooks/apperrors"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

const maxBodyBytes = 1 << 20

type errorResponse struct {
	Error string `json:"error"`
}

// PathID parses the {id} path variable, which must be a positive integer.
func PathID(r *http.Request) (int, error) {
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil || id <= 0 {
		return 0, apperrors.Validation{Reason: "id must be a positive integer"}
	}

	return id, nil
}

// DecodeJSON decodes the request body into v.
func DecodeJSON(w http.ResponseWriter, r *http.Request, v any) error {
	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)

	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		return apperrors.Validation{Reason: "invalid request body"}
	}

	return nil
}

// WriteJSON writes v as a JSON response with the given status code.
func WriteJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("write response: %v", err)
	}
}

// WriteError maps err to an HTTP status code and writes it as {"error": "..."}.
// Unexpected errors are logged and reported without their details.
func WriteError(w http.ResponseWriter, err error) {
	var (
		validation apperrors.Validation
		notFound   apperrors.NotFound
		conflict   apperrors.Conflict
	)

	switch {
	case errors.As(err, &validation):
		WriteJSON(w, http.StatusBadRequest, errorResponse{Error: validation.Error()})
	case errors.As(err, &notFound):
		WriteJSON(w, http.StatusNotFound, errorResponse{Error: notFound.Error()})
	case errors.As(err, &conflict):
		WriteJSON(w, http.StatusConflict, errorResponse{Error: conflict.Error()})
	default:
		log.Printf("internal error: %v", err)
		WriteJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal server error"})
	}
}
