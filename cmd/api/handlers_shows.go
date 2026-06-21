package main

import (
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/LunarDrift/deadabase/internal"
)

func (s *server) handleGetShows(w http.ResponseWriter, r *http.Request) {
	dateStr := r.URL.Query().Get("date")
	if dateStr == "" {
		respondWithError(w, http.StatusBadRequest, "Missing date parameter", nil)
		return
	}

	dateParsed, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid date format, expected YYYY-MM-DD", err)
		return
	}

	showRows, err := s.queries.GetShowFromDate(r.Context(), dateParsed)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not get show", err)
	}

	footnoteRows, err := s.queries.GetFootnotesFromShowID(r.Context(), showRows[0].ShowID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not get footnotes", err)
		return
	}

	if len(showRows) > 0 && !showRows[0].RawEntry.Valid {
		s := showRows[0]
		respondWithJSON(w, http.StatusOK, internal.ShowWithNoSetlist{
			ShowMeta: internal.ShowMeta{
				Date:  s.ShowDate.Format("2006-01-02"),
				Venue: s.Venue,
				City:  s.City,
				State: s.State,
				Notes: s.Notes.String,
			},
			Message: "No setlist available for this show",
		})
		return
	}

	var parsedShows []internal.ShowSortInput
	for _, show := range showRows {
		parsedShows = append(parsedShows, internal.ShowSortInput{
			ShowDate: show.ShowDate,
			Venue:    show.Venue,
			City:     show.City,
			State:    show.State,
			Notes:    show.Notes.String,
			SetName:  show.SetName.String,
			RawEntry: show.RawEntry.String,
		})
	}

	showResp, err := internal.SortSetPositions(parsedShows)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not sort set positions", err)
		return
	}
	showResp.Footnotes = make(map[string]string)
	for _, f := range footnoteRows {
		showResp.Footnotes[f.Marker] = f.NoteText
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

	showRows, err := s.queries.GetShowFromID(r.Context(), int32(id))
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not get shows", err)
		return
	}

	footnoteRows, err := s.queries.GetFootnotesFromShowID(r.Context(), int32(id))
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not get footnotes", err)
		return
	}

	if len(showRows) > 0 && !showRows[0].RawEntry.Valid {
		s := showRows[0]
		respondWithJSON(w, http.StatusOK, internal.ShowWithNoSetlist{
			ShowMeta: internal.ShowMeta{
				Date:  s.ShowDate.Format("2006-01-02"),
				Venue: s.Venue,
				City:  s.City,
				State: s.State,
				Notes: s.Notes.String,
			},
			Message: "No setlist available for this show",
		})
		return
	}

	var parsedShow []internal.ShowSortInput
	for _, row := range showRows {
		parsedShow = append(parsedShow, internal.ShowSortInput{
			ShowDate: row.ShowDate,
			Venue:    row.Venue,
			City:     row.City,
			State:    row.State,
			Notes:    row.Notes.String,
			SetName:  row.SetName.String,
			RawEntry: row.RawEntry.String,
		})
	}

	showResp, err := internal.SortSetPositions(parsedShow)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not sort set positions", err)
		return
	}

	showResp.Footnotes = make(map[string]string)
	for _, f := range footnoteRows {
		showResp.Footnotes[f.Marker] = f.NoteText
	}

	respondWithJSON(w, http.StatusOK, showResp)
}

func (s *server) handleGetRandomShow(w http.ResponseWriter, r *http.Request) {
	allIDs, err := s.queries.GetAllShowIDs(r.Context())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not get list of IDs", err)
		return
	}
	id := allIDs[rand.Intn(len(allIDs))]

	showRows, err := s.queries.GetShowFromID(r.Context(), int32(id))
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not get show", err)
		return
	}

	footnoteRows, err := s.queries.GetFootnotesFromShowID(r.Context(), id)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not get footnotes", err)
		return
	}

	if len(showRows) > 0 && !showRows[0].RawEntry.Valid {
		s := showRows[0]
		respondWithJSON(w, http.StatusOK, internal.ShowWithNoSetlist{
			ShowMeta: internal.ShowMeta{
				Date:  s.ShowDate.Format("2006-01-02"),
				Venue: s.Venue,
				City:  s.City,
				State: s.State,
				Notes: s.Notes.String,
			},
			Message: "No setlist available for this show",
		})
		return
	}

	var parsedShow []internal.ShowSortInput
	for _, row := range showRows {
		parsedShow = append(parsedShow, internal.ShowSortInput{
			ShowDate: row.ShowDate,
			Venue:    row.Venue,
			City:     row.City,
			State:    row.State,
			Notes:    row.Notes.String,
			SetName:  row.SetName.String,
			RawEntry: row.RawEntry.String,
		})
	}

	showResp, err := internal.SortSetPositions(parsedShow)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not sort set positions", err)
		return
	}
	showResp.Footnotes = make(map[string]string)
	for _, f := range footnoteRows {
		showResp.Footnotes[f.Marker] = f.NoteText
	}

	respondWithJSON(w, http.StatusOK, showResp)
}
