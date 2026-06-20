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

	var venueResults []internal.VenueSearchResult
	for _, result := range searchResults {
		venueResults = append(venueResults, internal.VenueSearchResult{
			ShowID: int(result.ShowID),
			Date:   result.Date.Format("2006-01-02"),
			Venue:  result.Venue,
			City:   result.City,
			State:  result.State,
		})
	}
	respondWithJSON(w, http.StatusOK, venueResults)
}
