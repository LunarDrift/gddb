package main

import (
	"database/sql"
	"log"
	"os"

	"github.com/LunarDrift/deadabase/internal/importer"
)

func main() {
	connStr := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("Could not connect to database:", err)
	}
	defer db.Close()

	// Verify connection works before trying to import anything
	if err = db.Ping(); err != nil {
		log.Fatal("Could not reach database: ", err)
	}

	err = importer.Run(db, "data/data.json")
	if err != nil {
		log.Fatal(err)
	}
}
