package main

import (
	"database/sql"
	"log"
	"os"

	"github.com/LunarDrift/deadabase/internal/importer"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	log.Println("step 1: loading env")
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Couldn't find environment file")
	}
	log.Println("step 2: env loaded")

	log.Println("step 3: grabbing db url from env")
	connStr := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("Could not connect to database:", err)
	}
	log.Println("step 4: db opened")
	defer db.Close()

	// Verify connection works before trying to import anything
	if err = db.Ping(); err != nil {
		log.Fatal("Could not reach database: ", err)
	}
	log.Println("step 5: db pinged ok")

	const file string = "data/data.json"
	log.Printf("step 6: about to run import on %q\n", file)
	err = importer.Run(db, file)
	if err != nil {
		log.Fatalf("error running import script on '%s': %v", file, err)
	}
	log.Println("step 7: import finished")
}
