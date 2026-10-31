package main

import (
	"database/sql"
	"log/slog"
	"net/http"

	"github.com/LunarDrift/deadabase/internal"
	_ "github.com/lib/pq"
)

type server struct {
	mux     *http.ServeMux
	db      *sql.DB
	queries internal.ShowQuerier
	logger  *slog.Logger
}

func NewServer(db *sql.DB, queries internal.ShowQuerier, logger *slog.Logger) *server {
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
	s.mux.HandleFunc("GET /stats/all-songs", s.handleGetAllSongStats)
	s.mux.HandleFunc("GET /stats/venues", s.handleGetVenueStats)

	s.mux.HandleFunc("GET /all-venues", s.handleGetAllVenues)
	s.mux.HandleFunc("GET /all-locations", s.handleGetAllLocations)
	s.mux.HandleFunc("GET /all-cities", s.handleGetAllCities)
}

func (s *server) handleHealth(w http.ResponseWriter, r *http.Request) {
	type healthResponse struct {
		Status string `json:"status"`
	}
	respondWithJSON(w, http.StatusOK, healthResponse{Status: `🪬 𝐎𝐍𝐋𝐈𝐍𝐄 🪬`})
}

func (s *server) handleGetAllVenues(w http.ResponseWriter, r *http.Request) {
	venues, err := s.queries.GetAllVenues(r.Context())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "could not get venues", err)
	}
	respondWithJSON(w, http.StatusOK, internal.AllVenues{Venues: venues})
}

func (s *server) handleGetAllLocations(w http.ResponseWriter, r *http.Request) {
	locations, err := s.queries.GetAllLocations(r.Context())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "could not get locations", err)
	}
	respondWithJSON(w, http.StatusOK, internal.AllLocations{Locations: locations})
}

func (s *server) handleGetAllCities(w http.ResponseWriter, r *http.Request) {
	cities, err := s.queries.GetAllCities(r.Context())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "could not get cities", err)
	}
	respondWithJSON(w, http.StatusOK, internal.AllCities{Cities: cities})
}
