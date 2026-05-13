package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/chenqiang1986/rainbow_timetravel/entity"
	"github.com/chenqiang1986/rainbow_timetravel/service"
	"github.com/gorilla/mux"
)

// POST /records/{id}
// if the record exists, the record is updated, namely a new version is created.
// if the record doesn't exist, the record is created.
func (a *API) PostRecords(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idNumber, ok := parseID(w, mux.Vars(r)["id"])
	if !ok {
		return
	}

	var body map[string]*string
	err := json.NewDecoder(r.Body).Decode(&body)

	if err != nil {
		err := writeError(w, "invalid input; could not parse json", http.StatusBadRequest)
		logError(err)
		return
	}

	lock := a.idLocks.get(idNumber)
	lock.Lock()
	defer lock.Unlock()

	// first retrieve the record
	record, err := a.records.GetRecord(ctx, idNumber)

	if !errors.Is(err, service.ErrRecordDoesNotExist) { // record exists
		record, err = a.records.UpdateRecord(ctx, idNumber, body)
	} else { // record does not exist

		// exclude the delete updates
		recordMap := map[string]string{}
		for key, value := range body {
			if value != nil {
				recordMap[key] = *value
			}
		}

		record = entity.Record{
			ID:   idNumber,
			Data: recordMap,
		}
		err = a.records.CreateRecord(ctx, record)
	}

	if err != nil {
		errInWriting := writeError(w, ErrInternal.Error(), http.StatusInternalServerError)
		logError(err)
		logError(errInWriting)
		return
	}

	err = writeJSON(w, record, http.StatusOK)
	logError(err)
}
