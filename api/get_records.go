package api

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

// GET /records/{id}
// GetRecord retrieves the record of latest version.
func (a *API) GetRecords(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idNumber, ok := parseID(w, mux.Vars(r)["id"])
	if !ok {
		return
	}

	lock := a.idLocks.get(idNumber)
	lock.RLock()
	defer lock.RUnlock()

	record, err := a.records.GetRecord(ctx, idNumber)
	if err != nil {
		err := writeError(w, fmt.Sprintf("record of id %v does not exist", idNumber), http.StatusBadRequest)
		logError(err)
		return
	}

	err = writeJSON(w, record, http.StatusOK)
	logError(err)
}
