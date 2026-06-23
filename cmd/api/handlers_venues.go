package main

import (
	"net/http"

	"github.com/LunarDrift/deadabase/internal"
)

func (s *server) handleSearchByVenue(w http.ResponseWriter, r *http.Request) {
	venue := r.URL.Query().Get("name")
	if venue == "" {
		respondWithError(w, http.StatusBadRequest, "Missing 'name' query parameter", nil)
	}

	searchResults, err := s.queries.SearchByVenue(r.Context(), "%"+venue+"%")
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not get venues", err)
		return
	}

	var venueResults []internal.ListOfShowsResult
	for _, result := range searchResults {
		venueResults = append(venueResults, internal.ListOfShowsResult{
			ShowID: result.ShowID,
			Date:   result.ShowDate.Format("2006-01-02"),
			Venue:  result.Venue,
			City:   result.City,
			State:  result.State,
		})
	}
	respondWithJSON(w, http.StatusOK, venueResults)
}
