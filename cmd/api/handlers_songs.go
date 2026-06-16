package main

import (
	"net/http"
)

func (s *server) handleMostPlayedSongs(w http.ResponseWriter, r *http.Request) {
	songs, err := s.queries.MostPlayedSongs(r.Context())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not get most played songs", err)
		return
	}
	respondWithJSON(w, http.StatusOK, songs)
}
