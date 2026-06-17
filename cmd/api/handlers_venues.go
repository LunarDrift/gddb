package main

import (
	"net/http"

	"github.com/LunarDrift/deadabase/internal"
)

func (s *server) handleSearchByVenue(w http.ResponseWriter, r *http.Request) {
	venue := r.URL.Query().Get("name")

	searchResults, err := s.queries.SearchByVenue(r.Context(), "%"+venue+"%")
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not get venues", err)
		return
	}

	result := internal.GroupByVenue(searchResults)

	respondWithJSON(w, http.StatusOK, result)
}
