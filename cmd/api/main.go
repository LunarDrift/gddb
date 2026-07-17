package main

import (
	"database/sql"
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/LunarDrift/deadabase/cmd/api/middleware"
	"github.com/LunarDrift/deadabase/internal/database"
	"github.com/joho/godotenv"
	"github.com/lmittmann/tint"
	"github.com/mattn/go-isatty"
)

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
	logger := slog.New(tint.NewTextHandler(os.Stderr, &tint.Options{
		Level:      slog.LevelDebug,
		TimeFormat: time.DateTime,
		NoColor:    !isatty.IsTerminal(os.Stderr.Fd()) && !isatty.IsCygwinTerminal(os.Stderr.Fd()),
	}))
	slog.SetDefault(logger)

	srv := NewServer(db, queries, logger)

	requestLogger := middleware.LoggerMiddleware(logger)
	limiter := middleware.NewIPRateLimiter(2, 10) // 2 req/sec sustained, burst of 10
	handler := requestLogger(limiter.Middleware(srv.mux))

	srv.logger.Info("Listening", "port", port)
	log.Fatal(http.ListenAndServe(":"+port, handler))
}
