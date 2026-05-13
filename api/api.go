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

// CreateRoutes registers the v1 routes.
func (a *API) CreateRoutes(routes *mux.Router) {
	routes.Path("/records/{id}").HandlerFunc(a.GetRecords).Methods("GET")
	routes.Path("/records/{id}").HandlerFunc(a.PostRecords).Methods("POST")
}

// CreateRoutesV2 registers the v2 routes: everything v1 exposes, plus
// version listing and point-in-time lookup.
func (a *API) CreateRoutesV2(routes *mux.Router) {
	a.CreateRoutes(routes)
	routes.Path("/records/{id}/versions").HandlerFunc(a.ListVersions).Methods("GET")
	routes.Path("/records/{id}/at").HandlerFunc(a.GetRecordAt).Methods("GET")
}
