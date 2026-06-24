package main

import (
	"net/http"
	"strconv"

	"github.com/LunarDrift/deadabase/internal"
)

func (s *server) handleMostPlayedSongs(w http.ResponseWriter, r *http.Request) {
	songs, err := s.queries.MostPlayedSongs(r.Context())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not get most played songs", err)
		return
	}
	respondWithJSON(w, http.StatusOK, songs)
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
		results = append(results, internal.SongsTimesPlayed{
			Song:        row.Song.String,
			TimesPlayed: int(row.TimesPlayed),
		})
	}

	respondWithJSON(w, http.StatusOK, results)
}
