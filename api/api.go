package api

import (
	"github.com/chenqiang1986/rainbow_timetravel/service"
	"github.com/gorilla/mux"
)

type API struct {
	records  service.RecordService
	idLocks  *idLocks
}

func NewAPI(records service.RecordService) *API {
	return &API{records: records, idLocks: newIDLocks()}
}

// generates all api routes
func (a *API) CreateRoutes(routes *mux.Router) {
	routes.Path("/records/{id}").HandlerFunc(a.GetRecords).Methods("GET")
	routes.Path("/records/{id}").HandlerFunc(a.PostRecords).Methods("POST")
}
