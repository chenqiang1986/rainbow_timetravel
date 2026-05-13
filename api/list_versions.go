package api

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/chenqiang1986/rainbow_timetravel/entity"
	"github.com/chenqiang1986/rainbow_timetravel/service"
	"github.com/gorilla/mux"
)

// GET /records/{id}/versions
// Lists every version (version number + timestamp) of the record.
func (a *API) ListVersions(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	idNumber, ok := parseID(w, mux.Vars(r)["id"])
	if !ok {
		return
	}

	lock := a.idLocks.get(idNumber)
	lock.RLock()
	defer lock.RUnlock()

	versions, err := a.records.ListVersions(ctx, idNumber)
	if errors.Is(err, service.ErrRecordDoesNotExist) {
		err := writeError(w, fmt.Sprintf("record of id %v does not exist", idNumber), http.StatusNotFound)
		logError(err)
		return
	}
	if err != nil {
		errInWriting := writeError(w, ErrInternal.Error(), http.StatusInternalServerError)
		logError(err)
		logError(errInWriting)
		return
	}

	resp := struct {
		ID       int                    `json:"id"`
		Versions []entity.RecordVersion `json:"versions"`
	}{ID: idNumber, Versions: versions}
	err = writeJSON(w, resp, http.StatusOK)
	logError(err)
}
