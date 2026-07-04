package main

import (
	"database/sql"
	"net/http"
	"strconv"
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

	type resp struct {
		City            string `json:"city"`
		StateOrCountry  string `json:"state_or_country"`
		UniqueSongCount int    `json:"unique_song_count"`
	}

	var results []resp
	for _, row := range songRows {
		results = append(results, resp{
			City:            row.City,
			StateOrCountry:  row.StateOrCountry,
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

	type resp struct {
		SongName string `json:"song"`
		Venue    string `json:"venue"`
		City     string `json:"city"`
		State    string `json:"state"`
	}

	var results []resp
	for _, row := range songRows {
		results = append(results, resp{
			SongName: row.SongName.String,
			Venue:    row.Venue,
			City:     row.City,
			State:    row.State,
		})
	}

	respondWithJSON(w, http.StatusOK, results)
}

func (s *server) handleGetSongStats(w http.ResponseWriter, r *http.Request) {
	song := r.PathValue("song")
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

	type resp struct {
		TimesPlayed int    `json:"times_played"`
		FirstPlayed string `json:"first_played"`
		LastPlayed  string `json:"last_played"`
	}

	respondWithJSON(w, http.StatusOK, resp{
		TimesPlayed: int(songStatRow.TimesPlayed),
		FirstPlayed: songStatRow.FirstPlayed.Format(time.DateOnly),
		LastPlayed:  songStatRow.LastPlayed.Format(time.DateOnly),
	})
}
