package main

import (
	"net/http"
	"strconv"

	"github.com/LunarDrift/deadabase/internal"
	"github.com/LunarDrift/deadabase/internal/database"
)

func (s *server) handleMostPlayedSongs(w http.ResponseWriter, r *http.Request) {
	songRows, err := s.queries.MostPlayedSongs(r.Context())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not get most played songs", err)
		return
	}
	var results []internal.SongsTimesPlayed
	for _, row := range songRows {
		results = append(results, database.RowToSongsTimesPlayed(row))
	}
	respondWithJSON(w, http.StatusOK, results)
}

func (s *server) handleSongsPlayedLessThanNTimes(w http.ResponseWriter, r *http.Request) {
	pval := r.URL.Query().Get("played_lt")
	num, err := strconv.Atoi(pval)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid value. Expecting number", err)
	}

	songRows, err := s.queries.SongsPlayedLessThan(r.Context(), num)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not get songs", err)
	}
	var results []internal.SongsTimesPlayed
	for _, row := range songRows {
		results = append(results, database.RowToSongsTimesPlayed(row))
	}

	respondWithJSON(w, http.StatusOK, results)
}

func (s *server) handleMostCommonEncoreSongs(w http.ResponseWriter, r *http.Request) {
	songRows, err := s.queries.MostCommonEncore(r.Context())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not get songs", err)
	}
	var results []internal.SongsTimesPlayed
	for _, row := range songRows {
		results = append(results, database.RowToSongsTimesPlayed(row))
	}
	respondWithJSON(w, http.StatusOK, results)
}

func (s *server) handleUniqueSongsPerCity(w http.ResponseWriter, r *http.Request) {
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
