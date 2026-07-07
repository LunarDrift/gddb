package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/LunarDrift/deadabase/internal"
	_ "github.com/lib/pq"
)

type server struct {
	mux     *http.ServeMux
	db      *sql.DB
	queries internal.ShowQuerier
}

func NewServer(db *sql.DB, queries internal.ShowQuerier) *server {
	srv := &server{
		mux:     http.NewServeMux(),
		db:      db,
		queries: queries,
	}
	srv.registerRoutes()
	return srv
}

func (s *server) registerRoutes() {
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

func respondWithJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Println("respondWithJSON - could not encode payload:", err)
	}
}

func respondWithError(w http.ResponseWriter, status int, message string, err error) {
	if err != nil {
		log.Println(err)
	}
	type errorResponse struct {
		Error string `json:"error"`
	}
	respondWithJSON(w, status, errorResponse{Error: message})
}

// fuzzyPattern wraps and inserts a '%' between every character of `input` to be used during SQL query searches
func fuzzyPattern(input string) string {
	words := strings.Fields(input)
	if len(words) == 0 {
		return ""
	}
	return "%" + strings.Join(words, "%") + "%"
}
