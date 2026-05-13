package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/chenqiang1986/rainbow_timetravel/api"
	"github.com/chenqiang1986/rainbow_timetravel/database"
	"github.com/chenqiang1986/rainbow_timetravel/service"
	"github.com/gorilla/mux"
)

// logError logs all non-nil errors
func logError(err error) {
	if err != nil {
		log.Printf("error: %v", err)
	}
}

func main() {
	router := mux.NewRouter()

	db, err := database.Open("database/rainbow.db")
	if err != nil {
		log.Fatalf("open database failed: %v", err)
	}
	defer db.Close()

	store := service.NewSQLiteRecordStore(db)
	apiHandler := api.NewAPI(store)

	healthHandler := func(w http.ResponseWriter, r *http.Request) {
		err := json.NewEncoder(w).Encode(map[string]bool{"ok": true})
		logError(err)
	}

	apiV1Route := router.PathPrefix("/api/v1").Subrouter()
	apiV1Route.Path("/health").HandlerFunc(healthHandler)
	apiHandler.CreateRoutes(apiV1Route)

	apiV2Route := router.PathPrefix("/api/v2").Subrouter()
	apiV2Route.Path("/health").HandlerFunc(healthHandler)
	apiHandler.CreateRoutesV2(apiV2Route)

	address := "127.0.0.1:8000"
	srv := &http.Server{
		Handler:      router,
		Addr:         address,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Printf("listening on %s", address)
	log.Fatal(srv.ListenAndServe())
}
