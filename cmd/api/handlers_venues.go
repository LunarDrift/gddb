package main

import "net/http"

func (s *server) handleSearchByVenue(w http.ResponseWriter, r *http.Request) {
	venue := r.URL.Query().Get("name")

	shows, err := s.queries.SearchByVenue(r.Context(), "%"+venue+"%")
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not get venues", err)
		return
	}
	// TODO: Figure out the correct payload structure

	// var result []internal.SearchByVenueShow
	// for _, show := range shows {
	// 	result = append(result, internal.SearchByVenueShow{
	// 		Date:    show.Date.Format("2006-01-02"),
	// 		Venue:   show.Venue,
	// 		Notes:   show.Notes.String,
	// 		SetName: show.SetName,
	// 		Song:    show.Song,
	// 	})
	// }
	respondWithJSON(w, http.StatusOK, shows)
}
