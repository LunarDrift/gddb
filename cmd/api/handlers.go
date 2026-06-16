package main

import (
	"net/http"
	"strconv"
	"time"

	"github.com/LunarDrift/deadabase/internal"
)

func (s *server) handleGetShows(w http.ResponseWriter, r *http.Request) {
	dateStr := r.URL.Query().Get("date")
	if dateStr == "" {
		respondWithError(w, http.StatusBadRequest, "missing date parameter", nil)
		return
	}

	dateParsed, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid date format, expected YYYY-MM-DD", err)
		return
	}

	showRow, err := s.queries.GetShowFromDate(r.Context(), dateParsed)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "could not get show", err)
	}

	var myShows []internal.ShowSortInput
	for _, show := range showRow {
		myShows = append(myShows, internal.ShowSortInput{
			ShowDate: show.ShowDate,
			Venue:    show.Venue,
			SetName:  show.SetName,
			RawEntry: show.RawEntry,
		})
	}

	showResp, err := internal.SortSetPositions(myShows)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not sort set positions", err)
		return
	}

	respondWithJSON(w, http.StatusOK, showResp)
}

func (s *server) handleGetShowFromID(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Missing or invalid id", err)
		return
	}

	shows, err := s.queries.GetShowFromID(r.Context(), int32(id))
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not get shows", err)
		return
	}

	var myShows []internal.ShowSortInput
	for _, show := range shows {
		myShows = append(myShows, internal.ShowSortInput{
			ShowDate: show.ShowDate,
			Venue:    show.Venue,
			SetName:  show.SetName,
			RawEntry: show.RawEntry,
		})
	}

	showResp, err := internal.SortSetPositions(myShows)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not sort set positions", err)
		return
	}

	respondWithJSON(w, http.StatusOK, showResp)
}
