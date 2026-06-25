package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/LunarDrift/deadabase/internal/database"

	_ "github.com/lib/pq"
)

type server struct {
	mux     *http.ServeMux
	db      *sql.DB
	queries *database.Queries
}

func NewServer(db *sql.DB, queries *database.Queries) *server {
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

	s.mux.HandleFunc("GET /shows", s.handlerShows)
	s.mux.HandleFunc("GET /shows/{value}", s.handleGetShow)
	s.mux.HandleFunc("GET /shows/random", s.handleGetRandomShow)
	s.mux.HandleFunc("GET /shows/between", s.handleGetShowsBetweenDates)

	s.mux.HandleFunc("GET /venues", s.handleSearchByVenue)

	s.mux.HandleFunc("GET /songs/mostplayed", s.handleMostPlayedSongs)
	s.mux.HandleFunc("GET /songs", s.handleSongsPlayedLessThanNTimes)

	s.mux.HandleFunc("GET /stats/top-encores", s.handleMostCommonEncoreSongs)
}

func (s *server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = fmt.Fprintln(w, `🪬𝐎𝐍𝐋𝐈𝐍𝐄🪬`)
}

func respondWithJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
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

func (s *server) handlerShows(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	switch {
	case query.Has("song"):
		s.handleGetShowsFromSongName(w, r)
		return
	default:
		respondWithError(w, http.StatusBadRequest, "Must provide a song query parameter", nil)
		return
	}
}
