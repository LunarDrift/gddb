package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/LunarDrift/deadabase/internal/importer"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	fmt.Println("step 1: loading env")
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Couldn't find environment file")
	}
	fmt.Println("step 2: env loaded")

	connStr := os.Getenv("DB_URL")
	fmt.Println("step 3: connStr = ", connStr)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("Could not connect to database:", err)
	}
	fmt.Println("step 4: db opened")
	defer db.Close()

	// Verify connection works before trying to import anything
	if err = db.Ping(); err != nil {
		log.Fatal("Could not reach database: ", err)
	}
	fmt.Println("step 5: db pinged ok")

	const file string = "data/data.json"
	fmt.Println("step 6: about to run import on ", file)
	err = importer.Run(db, file)
	if err != nil {
		log.Fatalf("error running import script on '%s': %v", file, err)
	}
	fmt.Println("step 7: import finished")
}
