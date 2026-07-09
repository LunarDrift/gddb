package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/LunarDrift/deadabase/internal"
)

func (s *server) handleSongsFromQueryParam(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	switch {
	case query.Get("sort") == "most_played" && query.Has("set_name"):
		s.handleGetMostPlayedSongsBySetName(w, r)

	case query.Get("sort") == "most_played":
		s.handleGetMostPlayedSongs(w, r)

	case query.Has("played_lt"):
		s.handleGetSongsPlayedLessThanNTimes(w, r)

	case query.Has("venue"):
		s.handleGetSongsPlayedAtVenue(w, r)

	default:
		respondWithError(w, http.StatusBadRequest, "Must provide a valid query parameter: played_lt, venue, sort=most_played, sort=most_played&set_name=", nil)
		return
	}
}

func (s *server) handleGetMostPlayedSongs(w http.ResponseWriter, r *http.Request) {
	songRows, err := s.queries.MostPlayedSongs(r.Context())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not get most played songs", err)
		return
	}
	var results []internal.SongsTimesPlayed
	for _, row := range songRows {
		results = append(results, internal.RowToSongsTimesPlayed(row))
	}
	respondWithJSON(w, http.StatusOK, results)
}

func (s *server) handleGetSongsPlayedLessThanNTimes(w http.ResponseWriter, r *http.Request) {
	val := r.URL.Query().Get("played_lt")
	num, err := strconv.Atoi(val)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid value. Expecting number", err)
	}

	songRows, err := s.queries.SongsPlayedLessThan(r.Context(), int32(num))
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not get songs", err)
	}

	var results []internal.SongsTimesPlayed
	for _, row := range songRows {
		results = append(results, internal.RowToSongsTimesPlayed(row))
	}

	respondWithJSON(w, http.StatusOK, results)
}

func (s *server) handleGetMostPlayedSongsBySetName(w http.ResponseWriter, r *http.Request) {
	setName := r.URL.Query().Get("set_name")
	if setName == "" {
		respondWithError(w, http.StatusBadRequest, "Missing set_name query parameter", nil)
		return
	}

	songRows, err := s.queries.MostCommonSongsBySetName(r.Context(), setName)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not get songs", err)
	}

	validSetNames := []string{"set_1", "set_2", "set_3", "encore", "acoustic_1", "acoustic_2", "acoustic", "electric"}
	if !slices.Contains(validSetNames, setName) {
		respondWithError(w, http.StatusBadRequest, fmt.Sprintf("Invalid set_name %q. Valid options: %s", setName, strings.Join(validSetNames, ", ")), nil)
		return
	}

	var results []internal.SongsTimesPlayed
	for _, row := range songRows {
		results = append(results, internal.RowToSongsTimesPlayed(row))
	}

	respondWithJSON(w, http.StatusOK, results)
}

func (s *server) handleGetUniqueSongsPerCity(w http.ResponseWriter, r *http.Request) {
	songRows, err := s.queries.UniqueSongsPerCity(r.Context())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not get data", err)
		return
	}

	var results []internal.UniqueSongsPerCity
	for _, row := range songRows {
		results = append(results, internal.UniqueSongsPerCity{
			City:            row.City,
			Location:        row.StateOrCountry,
			UniqueSongCount: int(row.UniqueSongCount),
		})
	}

	respondWithJSON(w, http.StatusOK, results)
}

func (s *server) handleGetSongsPlayedAtVenue(w http.ResponseWriter, r *http.Request) {
	venue := r.URL.Query().Get("venue")
	if venue == "" {
		respondWithError(w, http.StatusBadRequest, "Missing venue parameter", nil)
		return
	}

	searchPattern := fuzzyPattern(venue)

	songRows, err := s.queries.AllSongsPlayedAtVenue(r.Context(), searchPattern)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not get songs", err)
		return
	}

	if len(songRows) == 0 {
		respondWithError(w, http.StatusNotFound, "Venue not found", nil)
		return
	}

	results := []internal.SongsFromVenue{}
	for _, row := range songRows {
		results = append(results, internal.SongsFromVenue{
			SongName: row.SongName.String,
			Venue:    row.Venue,
			City:     row.City,
			Location: row.State,
		})
	}

	respondWithJSON(w, http.StatusOK, results)
}

func (s *server) handleGetSongStats(w http.ResponseWriter, r *http.Request) {
	song := r.PathValue("song")

	// This can never actually happen client-side; `/songs/` without a parameter will always 404.
	// But I'll keep it just in case I make changes in the future to /songs
	// And I'm asserting the 400 status code in the test. Otherwise I'd have to connect to the real server mux for a single test case
	if song == "" {
		respondWithError(w, http.StatusBadRequest, "Missing song name parameter", nil)
		return
	}

	// Split the song and add % for sql search pattern
	// So that searching for e.g. "Help on the way > Slipknot! > Franklin's Tower" becomes "Help%On%The%Way%>..."
	searchPattern := fuzzyPattern(song)
	songStatRow, err := s.queries.SongStats(r.Context(), sql.NullString{String: searchPattern, Valid: true})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not get song stats", err)
		return
	}

	respondWithJSON(w, http.StatusOK, internal.SongStats{
		TimesPlayed: int(songStatRow.TimesPlayed),
		FirstPlayed: songStatRow.FirstPlayed.Format(time.DateOnly),
		LastPlayed:  songStatRow.LastPlayed.Format(time.DateOnly),
	})
}
