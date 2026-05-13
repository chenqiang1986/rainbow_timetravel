package api

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/chenqiang1986/rainbow_timetravel/entity"
	"github.com/chenqiang1986/rainbow_timetravel/service"
	"github.com/gorilla/mux"
)

// GET /records/{id}/at?version={n}
// GET /records/{id}/at?timestamp={RFC3339}
// Returns the record as of the given version OR right before timestamp.
// Exactly one of the two query parameters must be supplied.
func (a *API) GetRecordAt(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	idNumber, ok := parseID(w, mux.Vars(r)["id"])
	if !ok {
		return
	}

	versionStr := r.URL.Query().Get("version")
	timestampStr := r.URL.Query().Get("timestamp")

	if (versionStr == "") == (timestampStr == "") {
		err := writeError(w, "exactly one of 'version' or 'timestamp' query parameters is required", http.StatusBadRequest)
		logError(err)
		return
	}

	lock := a.idLocks.get(idNumber)
	lock.RLock()
	defer lock.RUnlock()

	var (
		snapshot entity.RecordSnapshot
		err      error
	)

	if versionStr != "" {
		version, parseErr := strconv.Atoi(versionStr)
		if parseErr != nil || version <= 0 {
			err := writeError(w, "invalid version; version must be a positive integer", http.StatusBadRequest)
			logError(err)
			return
		}
		snapshot, err = a.records.GetRecordAtVersion(ctx, idNumber, version)
	} else {
		ts, parseErr := time.Parse(time.RFC3339, timestampStr)
		if parseErr != nil {
			err := writeError(w, "invalid timestamp; must be RFC3339 (e.g. 2026-05-13T12:00:00Z)", http.StatusBadRequest)
			logError(err)
			return
		}
		snapshot, err = a.records.GetRecordAtTimestamp(ctx, idNumber, ts)
	}

	if errors.Is(err, service.ErrRecordDoesNotExist) {
		err := writeError(w, fmt.Sprintf("no record of id %v exists at the requested point", idNumber), http.StatusNotFound)
		logError(err)
		return
	}
	if err != nil {
		errInWriting := writeError(w, ErrInternal.Error(), http.StatusInternalServerError)
		logError(err)
		logError(errInWriting)
		return
	}

	err = writeJSON(w, snapshot, http.StatusOK)
	logError(err)
}
