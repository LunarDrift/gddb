package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"github.com/LunarDrift/deadabase/cmd/api/middleware"
	"github.com/LunarDrift/deadabase/internal/database"
	"github.com/joho/godotenv"
)

var logger = log.New(os.Stdout, "DEBUG: ", log.LstdFlags)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal(err)
	}
	connectionString := os.Getenv("DB_URL")
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		log.Fatal(err)
	}

	queries := database.New(db)
	srv := NewServer(db, queries)

	requestLogger := middleware.RequestLogger(logger)
	limiter := middleware.NewIPRateLimiter(2, 10) // 2 req/sec sustained, burst of 10
	handler := requestLogger(limiter.Middleware(srv.mux))

	logger.Printf("Listening on %s...\n", port)
	log.Fatal(http.ListenAndServe(":"+port, handler))
}
