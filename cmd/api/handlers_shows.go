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

// handlerShows parses the query parameter and chooses the appropriate endpoint
func (s *server) handleShowsFromQueryParam(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	switch {
	case query.Has("year") && query.Has("location"):
		s.handleGetShowsFromYearAndLocation(w, r)

	case query.Has("year"):
		s.handleGetShowsFromYear(w, r)

	case query.Has("location"):
		s.handleGetShowsFromLocation(w, r)

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

// handleShowsFromPathVal parses the `value` path variable and chooses the appropriate endpoint
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
		return
	}

	if len(showRows) == 0 {
		respondWithError(w, http.StatusNotFound, "No show on that date", nil)
		return
	}

	// some dates have multiple shows attached - early show + late show
	// need to sort those separately so they don't get combined into a single show object
	var groups [][]internal.ShowSortInput
	for _, row := range showRows {
		parsed := internal.RowToShowSortInput(row)
		if n := len(groups); n > 0 && groups[n-1][0].ShowID == parsed.ShowID {
			groups[n-1] = append(groups[n-1], parsed)
		} else {
			groups = append(groups, []internal.ShowSortInput{parsed})
		}
	}

	results := []any{}
	for _, group := range groups {
		resp, err := s.buildShowResponse(r, group)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Could not build show response", err)
			return
		}
		results = append(results, resp)
	}

	respondWithJSON(w, http.StatusOK, results)
}

func (s *server) getShowFromID(w http.ResponseWriter, r *http.Request, id int32) {
	showRows, err := s.queries.GetShowFromID(r.Context(), int32(id))
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not get shows", err)
		return
	}

	if len(showRows) == 0 {
		respondWithError(w, http.StatusNotFound, "No show with that ID", nil)
		return
	}

	var parsedShow []internal.ShowSortInput
	for _, row := range showRows {
		parsedShow = append(parsedShow, internal.RowToShowSortInput(row))
	}

	resp, err := s.buildShowResponse(r, parsedShow)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not build show response", err)
		return
	}

	respondWithJSON(w, http.StatusOK, resp)
}

func (s *server) handleGetRandomShow(w http.ResponseWriter, r *http.Request) {
	allIDs, err := s.queries.GetAllShowIDs(r.Context())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not get list of IDs", err)
		return
	}
	if len(allIDs) == 0 {
		respondWithError(w, http.StatusInternalServerError, "No shows available", nil)
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
		parsedShow = append(parsedShow, internal.RowToShowSortInput(row))
	}

	resp, err := s.buildShowResponse(r, parsedShow)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not build show response", err)
		return
	}

	respondWithJSON(w, http.StatusOK, resp)
}

func (s *server) handleGetShowsBetweenDates(w http.ResponseWriter, r *http.Request) {
	startDateStr := r.URL.Query().Get("start_date")
	endDateStr := r.URL.Query().Get("end_date")
	if startDateStr == "" || endDateStr == "" {
		respondWithError(w, http.StatusBadRequest, "Missing date parameter. Must provide both start_date and end_date", nil)
		return
	}

	startDate, err := time.Parse(time.DateOnly, startDateStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid date format, expected YYYY-MM-DD", err)
		return
	}
	endDate, err := time.Parse(time.DateOnly, endDateStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid date format, expected YYYY-MM-DD", err)
		return
	}

	if endDate.Before(startDate) {
		respondWithError(w, http.StatusBadRequest, "end_date must be later than start_date", nil)
		return
	}

	showRows, err := s.queries.GetShowsBetweenDates(r.Context(), database.GetShowsBetweenDatesParams{
		ShowDate:   startDate,
		ShowDate_2: endDate,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not get shows between dates", err)
		return
	}
	showResults := []internal.ShowMeta{}
	for _, row := range showRows {
		showResults = append(showResults, internal.RowToShowMeta(row))
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

	if len(showRows) == 0 {
		respondWithError(w, http.StatusNotFound, "Song not found", nil)
		return
	}

	showResults := []internal.ShowMeta{}
	for _, row := range showRows {
		showResults = append(showResults, internal.RowToShowMeta(row))
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
		showResults = append(showResults, internal.RowToShowMeta(row))
	}

	respondWithJSON(w, http.StatusOK, showResults)
}

func (s *server) handleGetShowsFromVenueName(w http.ResponseWriter, r *http.Request) {
	venue := r.URL.Query().Get("venue")
	if venue == "" {
		respondWithError(w, http.StatusBadRequest, "Missing 'venue' query parameter", nil)
		return
	}

	searchPattern := fuzzyPattern(venue)
	venueRows, err := s.queries.SearchByVenue(r.Context(), searchPattern)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not get venues", err)
		return
	}

	if len(venueRows) == 0 {
		msg := fmt.Sprintf("Venue '%s' not found", venue)
		respondWithError(w, http.StatusNotFound, msg, nil)
		return
	}

	var venueResults []internal.ShowMeta
	for _, row := range venueRows {
		venueResults = append(venueResults, internal.RowToShowMeta(row))
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
		return nil, fmt.Errorf("showsWithNotes: %w", err)
	}

	var results []internal.ShowMeta
	for _, row := range showRows {
		results = append(results, internal.RowToShowMeta(row))
	}
	return results, nil
}

func (s *server) showsNoNotes(ctx context.Context) ([]internal.ShowMeta, error) {
	showRows, err := s.queries.ShowsWithoutNotes(ctx)
	if err != nil {
		return nil, fmt.Errorf("showsNoNotes: %w", err)
	}

	var results []internal.ShowMeta
	for _, row := range showRows {
		results = append(results, internal.RowToShowMeta(row))
	}
	return results, nil
}

func (s *server) handleGetShowsFromYearAndLocation(w http.ResponseWriter, r *http.Request) {
	yearStr := r.URL.Query().Get("year")
	if yearStr == "" {
		respondWithError(w, http.StatusBadRequest, "Missing 'year' parameter", nil)
		return
	}
	location := r.URL.Query().Get("location")
	if location == "" {
		respondWithError(w, http.StatusBadRequest, "Missing 'location' parameter", nil)
		return
	}
	location = strings.ToLower(location)

	validLocations, err := s.queries.GetValidLocations(r.Context())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not get valid locations", err)
		return
	}
	if len(validLocations) == 0 {
		respondWithError(w, http.StatusInternalServerError, "Valid locations list empty", nil)
		return
	}

	if !slices.Contains(validLocations, location) {
		respondWithError(w, http.StatusBadRequest, "Invalid location", nil)
		return
	}

	year, err := strconv.Atoi(yearStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid year parameter", err)
		return
	}

	showRows, err := s.queries.GetShowsFromYearAndLocation(r.Context(), database.GetShowsFromYearAndLocationParams{
		Year:     int32(year),
		Location: location,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not get shows", err)
		return
	}

	var results []internal.ShowMeta
	for _, row := range showRows {
		results = append(results, internal.RowToShowMeta(row))
	}

	respondWithJSON(w, http.StatusOK, results)
}

func (s *server) handleGetShowsFromYear(w http.ResponseWriter, r *http.Request) {
	yearStr := r.URL.Query().Get("year")
	if yearStr == "" {
		respondWithError(w, http.StatusBadRequest, "Missing 'year' parameter", nil)
		return
	}

	year, err := strconv.Atoi(yearStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid year value", err)
		return
	}

	if year < 1965 || year > 1995 {
		respondWithError(w, http.StatusBadRequest, "Year must be between 1965-1995", nil)
		return
	}

	showRows, err := s.queries.GetShowsFromYear(r.Context(), int32(year))
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not get shows", err)
		return
	}

	var results []internal.ShowMeta
	for _, row := range showRows {
		results = append(results, internal.RowToShowMeta(row))
	}

	respondWithJSON(w, http.StatusOK, results)
}

func (s *server) handleGetShowsFromLocation(w http.ResponseWriter, r *http.Request) {
	location := r.URL.Query().Get("location")
	if location == "" {
		respondWithError(w, http.StatusBadRequest, "Missing 'location' parameter", nil)
		return
	}
	location = strings.ToLower(location)

	validLocations, err := s.queries.GetValidLocations(r.Context())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not get valid locations", err)
		return
	}
	if len(validLocations) == 0 {
		respondWithError(w, http.StatusInternalServerError, "Valid locations list empty", nil)
		return
	}

	if !slices.Contains(validLocations, location) {
		respondWithError(w, http.StatusBadRequest, "Invalid location", nil)
		return
	}

	showRows, err := s.queries.GetShowsFromLocation(r.Context(), location)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not get shows", err)
		return
	}

	var results []internal.ShowMeta
	for _, row := range showRows {
		results = append(results, internal.RowToShowMeta(row))
	}

	respondWithJSON(w, http.StatusOK, results)
}
