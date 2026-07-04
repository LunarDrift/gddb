package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/LunarDrift/deadabase/internal/database"

	_ "github.com/lib/pq"
)

type ShowQuerier interface {
	GetAllShowIDs(ctx context.Context) ([]int32, error)
	GetShowFromDate(ctx context.Context, showDate time.Time) ([]database.GetShowFromDateRow, error)
	GetShowFromID(ctx context.Context, showID int32) ([]database.GetShowFromIDRow, error)
	GetShowsBetweenDates(ctx context.Context, arg database.GetShowsBetweenDatesParams) ([]database.GetShowsBetweenDatesRow, error)
	GetShowsFromSetName(ctx context.Context, setName string) ([]database.GetShowsFromSetNameRow, error)
	GetShowsFromSongName(ctx context.Context, rawEntry string) ([]database.GetShowsFromSongNameRow, error)
	GetShowsFromState(ctx context.Context, stateOrCountry string) ([]database.GetShowsFromStateRow, error)
	GetShowsFromYear(ctx context.Context, year int32) ([]database.GetShowsFromYearRow, error)
	GetShowsFromYearAndState(ctx context.Context, arg database.GetShowsFromYearAndStateParams) ([]database.GetShowsFromYearAndStateRow, error)
	SearchByVenue(ctx context.Context, venue string) ([]database.SearchByVenueRow, error)
	ShowsWithShowNotes(ctx context.Context) ([]database.ShowsWithShowNotesRow, error)
	ShowsWithoutNotes(ctx context.Context) ([]database.ShowsWithoutNotesRow, error)
	SongStats(ctx context.Context, songName sql.NullString) (database.SongStatsRow, error)
	AllSongsPlayedAtVenue(ctx context.Context, venue string) ([]database.AllSongsPlayedAtVenueRow, error)
	MostCommonSongsBySetName(ctx context.Context, setName string) ([]database.MostCommonSongsBySetNameRow, error)
	MostPlayedSongs(ctx context.Context) ([]database.MostPlayedSongsRow, error)
	SongsPlayedLessThan(ctx context.Context, dollar_1 interface{}) ([]database.SongsPlayedLessThanRow, error)
	UniqueSongsPerCity(ctx context.Context) ([]database.UniqueSongsPerCityRow, error)
}

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

	s.mux.HandleFunc("GET /shows", s.handleShowsFromQueryParam)
	s.mux.HandleFunc("GET /shows/{value}", s.handleShowsFromPathVal)
	s.mux.HandleFunc("GET /shows/random", s.handleGetRandomShow)

	s.mux.HandleFunc("GET /songs", s.handleSongsFromQueryParam)
	s.mux.HandleFunc("GET /songs/{song}", s.handleGetSongStats)

	s.mux.HandleFunc("GET /stats/songs-per-city", s.handleGetUniqueSongsPerCity)
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

func fuzzyPattern(input string) string {
	words := strings.Fields(input)
	return "%" + strings.Join(words, "%") + "%"
}
