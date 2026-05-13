package api

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
)

var (
	ErrInternal = errors.New("internal error")
)

// logs an error if it's not nil
func logError(err error) {
	if err != nil {
		log.Printf("error: %v", err)
	}
}

// writeJSON writes the data as json.
func writeJSON(w http.ResponseWriter, data interface{}, statusCode int) error {
	w.Header().Add("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)
	err := json.NewEncoder(w).Encode(data)
	return err
}

// writeError writes the message as an error
func writeError(w http.ResponseWriter, message string, statusCode int) error {
	log.Printf("response errored: %s", message)
	return writeJSON(
		w,
		map[string]string{"error": message},
		statusCode,
	)
}

// parseID parses the {id} path variable and writes a 400 response on failure.
// Returns (id, true) on success or (0, false) on failure.
func parseID(w http.ResponseWriter, raw string) (int, bool) {
	idNumber, err := strconv.ParseInt(raw, 10, 32)
	if err != nil || idNumber <= 0 {
		err := writeError(w, "invalid id; id must be a positive number", http.StatusBadRequest)
		logError(err)
		return 0, false
	}
	return int(idNumber), true
}
