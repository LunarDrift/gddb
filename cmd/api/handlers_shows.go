package main

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/LunarDrift/deadabase/internal"
	"github.com/LunarDrift/deadabase/internal/database"
)

// respondWithShow takes a slice of ShowSortInput rows for a single show (already
// fetched by date, ID, or random selection) and handles the rest of the response:
// returning a "no setlist available" message if the show has no songs, otherwise
// sorting the sets/songs and attaching any footnotes before writing the JSON response
func (s *server) respondWithShow(w http.ResponseWriter, r *http.Request, parsedShow []internal.ShowSortInput) {
	if len(parsedShow) > 0 && parsedShow[0].RawEntry == "" {
		row := parsedShow[0]
		respondWithJSON(w, http.StatusOK, internal.ShowWithNoSetlist{
			ShowMeta: internal.ShowMeta{
				ShowID: row.ShowID,
				Date:   row.ShowDate.Format(time.DateOnly),
				Venue:  row.Venue,
				City:   row.City,
				State:  row.State,
				Notes:  row.Notes,
			},
			Message: "No setlist available for this show",
		})
		return
	}

	showResp, err := internal.SortSetPositions(parsedShow)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not sort set positions", err)
		return
	}

	footnoteRows, err := s.queries.GetFootnotesFromShowID(r.Context(), parsedShow[0].ShowID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not get footnotes", err)
		return
	}
	showResp.Footnotes = make(map[string]string)
	for _, f := range footnoteRows {
		showResp.Footnotes[f.Marker] = f.NoteText
	}

	respondWithJSON(w, http.StatusOK, showResp)
}

// handlerShows parses the query parameter and chooses the appropriate endpoint
func (s *server) handlerShowsWithQueryParam(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	switch {
	case query.Has("song"):
		s.handleGetShowsFromSongName(w, r)

	case query.Has("set_type"):
		s.handleGetShowsFromSetName(w, r)

	case query.Has("venue"):
		s.handleGetShowsFromVenueName(w, r)

	case query.Has("has_notes"):
		s.handleGetShowsFromNotes(w, r)

	default:
		respondWithError(w, http.StatusBadRequest, "Must provide a song query parameter", nil)
		return
	}
}

// handleGetShow parses the `value` path variable and chooses the appropriate endpoint
// to send it to
func (s *server) handleGetShow(w http.ResponseWriter, r *http.Request) {
	value := r.PathValue("value")

	if id, err := strconv.Atoi(value); err == nil {
		s.getShowFromID(w, r, int32(id))
		return
	}

	if date, err := time.Parse(time.DateOnly, value); err == nil {
		s.getShowFromDate(w, r, date)
		return
	}

	respondWithError(w, http.StatusBadRequest, "Invalid show identifier, expected an ID or YYYY-MM-DD date", nil)
}

func (s *server) getShowFromDate(w http.ResponseWriter, r *http.Request, date time.Time) {
	showRows, err := s.queries.GetShowFromDate(r.Context(), date)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not get show", err)
	}

	var parsedShows []internal.ShowSortInput
	for _, row := range showRows {
		parsedShows = append(parsedShows, database.RowToShowSortInput(row))
	}

	s.respondWithShow(w, r, parsedShows)
}

func (s *server) getShowFromID(w http.ResponseWriter, r *http.Request, id int32) {
	showRows, err := s.queries.GetShowFromID(r.Context(), int32(id))
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not get shows", err)
		return
	}

	var parsedShow []internal.ShowSortInput
	for _, row := range showRows {
		parsedShow = append(parsedShow, database.RowToShowSortInput(row))
	}

	s.respondWithShow(w, r, parsedShow)
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

	var parsedShow []internal.ShowSortInput
	for _, row := range showRows {
		parsedShow = append(parsedShow, database.RowToShowSortInput(row))
	}

	s.respondWithShow(w, r, parsedShow)
}

func (s *server) handleGetShowsBetweenDates(w http.ResponseWriter, r *http.Request) {
	startDateStr := r.URL.Query().Get("startdate")
	endDateStr := r.URL.Query().Get("enddate")
	if startDateStr == "" || endDateStr == "" {
		respondWithError(w, http.StatusBadRequest, "Missing date parameter", nil)
		return
	}

	startDateParsed, err := time.Parse(time.DateOnly, startDateStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid date format, expected YYYY-MM-DD", err)
		return
	}
	endDateParsed, err := time.Parse(time.DateOnly, endDateStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid date format, expected YYYY-MM-DD", err)
		return
	}

	showRows, err := s.queries.GetShowsBetweenDates(r.Context(), database.GetShowsBetweenDatesParams{
		ShowDate:   startDateParsed,
		ShowDate_2: endDateParsed,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not get shows between dates", err)
		return
	}
	var showResults []internal.ShowMeta
	for _, row := range showRows {
		showResults = append(showResults, database.RowToShowMeta(row))
	}
	respondWithJSON(w, http.StatusOK, showResults)
}

func (s *server) handleGetShowsFromSongName(w http.ResponseWriter, r *http.Request) {
	songName := r.URL.Query().Get("song")
	if songName == "" {
		respondWithError(w, http.StatusBadRequest, "Missing 'song' query parameter", nil)
		return
	}

	searchPattern := fuzzyPattern(songName)
	showRows, err := s.queries.GetShowsFromSongName(r.Context(), searchPattern)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not get shows", err)
		return
	}

	var showResults []internal.ShowMeta
	for _, row := range showRows {
		showResults = append(showResults, database.RowToShowMeta(row))
	}
	respondWithJSON(w, http.StatusOK, showResults)
}

func (s *server) handleGetShowsFromSetName(w http.ResponseWriter, r *http.Request) {
	setName := r.URL.Query().Get("set_type")

	validSetNames := []string{"set_1", "set_2", "set_3", "encore", "acoustic", "electric"}
	if !slices.Contains(validSetNames, setName) {
		respondWithError(w, http.StatusBadRequest, fmt.Sprintf("Invalid set_type %q. Valid options: %s", setName, strings.Join(validSetNames, ", ")), nil)
		return
	}

	showRows, err := s.queries.GetShowsFromSetName(r.Context(), setName)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not get shows", err)
		return
	}

	var showResults []internal.ShowMeta
	for _, row := range showRows {
		showResults = append(showResults, database.RowToShowMeta(row))
	}

	respondWithJSON(w, http.StatusOK, showResults)
}

func (s *server) handleGetShowsFromVenueName(w http.ResponseWriter, r *http.Request) {
	venue := r.URL.Query().Get("venue")
	if venue == "" {
		respondWithError(w, http.StatusBadRequest, "Missing 'name' query parameter", nil)
	}

	searchPattern := fuzzyPattern(venue)
	venueRows, err := s.queries.SearchByVenue(r.Context(), searchPattern)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not get venues", err)
		return
	}

	var venueResults []internal.ShowMeta
	for _, row := range venueRows {
		venueResults = append(venueResults, database.RowToShowMeta(row))
	}
	respondWithJSON(w, http.StatusOK, venueResults)
}

func (s *server) handleGetShowsFromNotes(w http.ResponseWriter, r *http.Request) {
	val := r.URL.Query().Get("has_notes")

	b, err := strconv.ParseBool(val)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "has_notes must be true or false", nil)
		return
	}

	if b {
		results, err := s.showsWithNotes(r.Context())
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Could not get shows", err)
			return
		}
		respondWithJSON(w, http.StatusOK, results)
	} else {
		results, err := s.showsNoNotes(r.Context())
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Could not get shows", err)
			return
		}
		respondWithJSON(w, http.StatusOK, results)
	}
}

func (s *server) showsWithNotes(ctx context.Context) ([]internal.ShowMeta, error) {
	showRows, err := s.queries.ShowsWithShowNotes(ctx)
	if err != nil {
		return nil, err
	}

	var results []internal.ShowMeta
	for _, row := range showRows {
		results = append(results, database.RowToShowMeta(row))
	}
	return results, nil
}

func (s *server) showsNoNotes(ctx context.Context) ([]internal.ShowMeta, error) {
	showRows, err := s.queries.ShowsWithoutNotes(ctx)
	if err != nil {
		return nil, err
	}

	var results []internal.ShowMeta
	for _, row := range showRows {
		results = append(results, database.RowToShowMeta(row))
	}
	return results, nil
}
