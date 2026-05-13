package main

import (
	"flag"
	"log"

	"github.com/chenqiang1986/rainbow_timetravel/database"
)

func main() {
	dbPath := flag.String("db", "database/rainbow.db", "path to the SQLite database file")
	schemaPath := flag.String("schema", "database/schema.sql", "path to the schema SQL file")
	flag.Parse()

	db, err := database.Setup(*dbPath, *schemaPath)
	if err != nil {
		log.Fatalf("setup failed: %v", err)
	}
	defer db.Close()

	log.Printf("database ready at %s (schema applied from %s)", *dbPath, *schemaPath)
}
