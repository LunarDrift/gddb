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
// sorting the sets+songs and attaching any footnotes before writing the JSON response
func (s *server) respondWithShow(w http.ResponseWriter, r *http.Request, parsedShowRows []internal.ShowSortInput) {
	if len(parsedShowRows) > 0 && parsedShowRows[0].RawEntry == "" {
		row := parsedShowRows[0]
		respondWithJSON(w, http.StatusOK, internal.ShowWithNoSetlist{
			ShowMeta: internal.ShowMeta{
				ShowID: row.ShowID,
				Date:   row.Date.Format(time.DateOnly),
				Venue:  row.Venue,
				City:   row.City,
				State:  row.State,
				Notes:  row.Notes,
			},
			Message: "No setlist available for this show",
		})
		return
	}

	showResp, err := internal.SortSetPositions(parsedShowRows)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not sort set positions", err)
		return
	}

	footnoteRows, err := s.queries.GetFootnotesFromShowID(r.Context(), parsedShowRows[0].ShowID)
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
func (s *server) handleShowsFromQueryParam(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	switch {
	case query.Has("year") && query.Has("state"):
		s.handleGetShowsFromYearAndState(w, r)

	case query.Has("year"):
		s.handleGetShowsFromYear(w, r)

	case query.Has("state"):
		s.handleGetShowsFromState(w, r)

	case query.Has("song"):
		s.handleGetShowsFromSongName(w, r)

	case query.Has("set_name"):
		s.handleGetShowsFromSetName(w, r)

	case query.Has("venue"):
		s.handleGetShowsFromVenueName(w, r)

	case query.Has("has_notes"):
		s.handleGetShowsFromNotes(w, r)

	case query.Has("start_date") || query.Has("end_date"):
		s.handleGetShowsBetweenDates(w, r)

	default:
		respondWithError(w, http.StatusBadRequest, "Must provide a valid query parameter: song, set_name, venue, has_notes, start_date&end_date, year, year&state", nil)
		return
	}
}

// handleGetShow parses the `value` path variable and chooses the appropriate endpoint
// to send it to
func (s *server) handleShowsFromPathVal(w http.ResponseWriter, r *http.Request) {
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
	startDateStr := r.URL.Query().Get("start_date")
	endDateStr := r.URL.Query().Get("end_date")
	if startDateStr == "" || endDateStr == "" {
		respondWithError(w, http.StatusBadRequest, "Missing date parameter. Must provide both start_date and end_date", nil)
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
	setName := r.URL.Query().Get("set_name")

	validSetNames := []string{"set_1", "set_2", "set_3", "encore", "acoustic_1", "acoustic_2", "acoustic", "electric"}
	if !slices.Contains(validSetNames, setName) {
		respondWithError(w, http.StatusBadRequest, fmt.Sprintf("Invalid set_name %q. Valid options: %s", setName, strings.Join(validSetNames, ", ")), nil)
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

func (s *server) handleGetShowsFromYearAndState(w http.ResponseWriter, r *http.Request) {
	yearStr := r.URL.Query().Get("year")
	if yearStr == "" {
		respondWithError(w, http.StatusBadRequest, "Missing year parameter", nil)
		return
	}
	state := r.URL.Query().Get("state")
	if state == "" {
		respondWithError(w, http.StatusBadRequest, "Missing state parameter", nil)
		return
	}

	year, err := strconv.Atoi(yearStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid year parameter", err)
		return
	}

	showRows, err := s.queries.GetShowsFromYearAndState(r.Context(), database.GetShowsFromYearAndStateParams{
		Year:           int32(year),
		StateOrCountry: state,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not get shows", err)
		return
	}

	var results []internal.ShowMeta
	for _, row := range showRows {
		results = append(results, database.RowToShowMeta(row))
	}

	respondWithJSON(w, http.StatusOK, results)
}

func (s *server) handleGetShowsFromYear(w http.ResponseWriter, r *http.Request) {
	yearStr := r.URL.Query().Get("year")
	if yearStr == "" {
		respondWithError(w, http.StatusBadRequest, "Missing year parameter", nil)
		return
	}

	year, err := strconv.Atoi(yearStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid year value", err)
		return
	}

	showRows, err := s.queries.GetShowsFromYear(r.Context(), int32(year))
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not get shows", err)
		return
	}

	var results []internal.ShowMeta
	for _, row := range showRows {
		results = append(results, database.RowToShowMeta(row))
	}

	respondWithJSON(w, http.StatusOK, results)
}

func (s *server) handleGetShowsFromState(w http.ResponseWriter, r *http.Request) {
	state := r.URL.Query().Get("state")
	if state == "" {
		respondWithError(w, http.StatusBadRequest, "Missing state parameter", nil)
		return
	}

	showRows, err := s.queries.GetShowsFromState(r.Context(), state)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not get shows", err)
		return
	}

	var results []internal.ShowMeta
	for _, row := range showRows {
		results = append(results, database.RowToShowMeta(row))
	}

	respondWithJSON(w, http.StatusOK, results)
}
