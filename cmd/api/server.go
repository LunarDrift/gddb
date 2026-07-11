package main

import (
	"database/sql"
	"log"
	"net/http"

	_ "embed"

	"github.com/LunarDrift/deadabase/internal"
	_ "github.com/lib/pq"
)

type server struct {
	mux     *http.ServeMux
	db      *sql.DB
	queries internal.ShowQuerier
	logger  *log.Logger
}

func NewServer(db *sql.DB, queries internal.ShowQuerier, logger *log.Logger) *server {
	srv := &server{
		mux:     http.NewServeMux(),
		db:      db,
		queries: queries,
		logger:  logger,
	}
	srv.registerRoutes()
	return srv
}

func (s *server) registerRoutes() {
	s.mux.Handle("/", http.FileServer(http.Dir("./static")))

	s.mux.HandleFunc("GET /health", s.handleHealth)

	s.mux.HandleFunc("GET /shows", s.handleShowsFromQueryParam)
	s.mux.HandleFunc("GET /shows/{value}", s.handleShowsFromPathVal)
	s.mux.HandleFunc("GET /shows/random", s.handleGetRandomShow)

	s.mux.HandleFunc("GET /songs", s.handleSongsFromQueryParam)
	s.mux.HandleFunc("GET /songs/{song}", s.handleGetSongStats)

	s.mux.HandleFunc("GET /stats/songs-per-city", s.handleGetUniqueSongsPerCity)
}

func (s *server) handleHealth(w http.ResponseWriter, r *http.Request) {
	type healthResponse struct {
		Status string `json:"status"`
	}
	respondWithJSON(w, http.StatusOK, healthResponse{Status: `🪬 𝐎𝐍𝐋𝐈𝐍𝐄 🪬`})
}
