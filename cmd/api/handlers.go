package main

import (
	"net/http"
	"sort"
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

	setsMap := map[string][]string{}
	var venue string
	var date time.Time

	for _, row := range showRow {
		venue = row.Venue
		date = row.ShowDate
		setsMap[row.SetName] = append(setsMap[row.SetName], row.RawEntry)
	}

	setNames := make([]string, 0, len(setsMap))
	for k := range setsMap {
		setNames = append(setNames, k)
	}
	sort.Slice(setNames, func(i, j int) bool {
		return internal.SetPosition(setNames[i]) < internal.SetPosition(setNames[j])
	})

	sets := []internal.SetResponse{}
	for _, key := range setNames {
		sets = append(sets, internal.SetResponse{
			SetName: key,
			Songs:   setsMap[key],
		})
	}

	respondWithJSON(w, http.StatusOK, internal.ShowResponse{
		Date:  date.Format("2006-01-02"),
		Venue: venue,
		Sets:  sets,
	})
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

	setsMap := map[string][]string{}
	var venue string
	var date time.Time

	for _, row := range shows {
		venue = row.Venue
		date = row.ShowDate
		setsMap[row.SetName] = append(setsMap[row.SetName], row.RawEntry)
	}

	setNames := make([]string, 0, len(setsMap))
	for k := range setsMap {
		setNames = append(setNames, k)
	}
	sort.Slice(setNames, func(i, j int) bool {
		return internal.SetPosition(setNames[i]) < internal.SetPosition(setNames[j])
	})

	sets := []internal.SetResponse{}
	for _, key := range setNames {
		sets = append(sets, internal.SetResponse{
			SetName: key,
			Songs:   setsMap[key],
		})
	}

	respondWithJSON(w, http.StatusOK, internal.ShowResponse{
		Date:  date.Format("2006-01-02"),
		Venue: venue,
		Sets:  sets,
	})
}
