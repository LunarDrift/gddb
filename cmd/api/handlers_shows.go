package main

import (
	"context"
	"database/sql"
	"errors"
	"math/rand"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/LunarDrift/deadabase/internal"
	"github.com/LunarDrift/deadabase/internal/database"
)

// dateFilters holds the optional values for the handleSearchShowsFromDate endpoint
type dateFilters struct {
	year  int32
	month int32
	day   int32
}

// handlerShows parses the query parameter and chooses the appropriate endpoint
func (s *server) handleShowsFromQueryParam(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	switch {
	case query.Has("year") ||
		query.Has("month") ||
		query.Has("day") ||
		query.Has("location") ||
		query.Has("city") ||
		query.Has("song") ||
		query.Has("venue") ||
		query.Has("notes") ||
		query.Has("set_name") ||
		query.Has("has_notes"):
		s.handleSearchShows(w, r)

	case query.Has("date"):
		s.handleGetShowFromDate(w, r)

	case query.Has("id"):
		s.handleGetShowFromID(w, r)

	case query.Has("start_date") || query.Has("end_date"):
		s.handleGetShowsBetweenDates(w, r)

	default:
		s.handleGetAllShows(w, r)
	}
}

func (s *server) handleGetShowFromDate(w http.ResponseWriter, r *http.Request) {
	dateQuery := r.URL.Query().Get("date")
	if dateQuery == "" {
		respondWithError(w, http.StatusBadRequest, "Missing date query parameter", nil)
		return
	}
	date, err := time.Parse(time.DateOnly, dateQuery)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid date format. Want YYYY-MM-DD", err)
		return
	}
	showRows, err := s.queries.GetShowDetailsFromDate(r.Context(), date)
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
	groups := [][]internal.ShowSortInput{}
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

func (s *server) handleGetShowFromID(w http.ResponseWriter, r *http.Request) {
	queryID := r.URL.Query().Get("id")
	if queryID == "" {
		respondWithError(w, http.StatusBadRequest, "Missing ID parameter", nil)
		return
	}
	id, err := strconv.Atoi(queryID)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid ID parameter", err)
		return
	}

	showRows, err := s.queries.GetShowDetailsFromID(r.Context(), int32(id))
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not get shows", err)
		return
	}

	if len(showRows) == 0 {
		respondWithError(w, http.StatusNotFound, "No show with that ID", nil)
		return
	}

	parsedShow := []internal.ShowSortInput{}
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

	showRows, err := s.queries.GetShowDetailsFromID(r.Context(), int32(id))
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not get show", err)
		return
	}

	if len(showRows) == 0 {
		respondWithError(w, http.StatusInternalServerError, "Malformed show details", nil)
		return
	}

	parsedShow := []internal.ShowSortInput{}
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

	limit, offset, err := parsePagination(r)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error(), err)
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
		StartDate:  startDate,
		EndDate:    endDate,
		PageOffset: int32(offset),
		PageLimit:  int32(limit),
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not get shows between dates", err)
		return
	}

	if len(showRows) == 0 && offset > 0 {
		respondWithError(w, http.StatusBadRequest, "Offset out of range", nil)
		return
	}
	if len(showRows) == 0 && offset == 0 {
		respondWithError(w, http.StatusNotFound, "No shows between those dates", nil)
		return
	}

	next, prev := buildLinks(r, int(showRows[0].Count), offset, limit)
	showResults := internal.Paginated[internal.ShowMeta]{
		Count:    showRows[0].Count,
		Next:     next,
		Previous: prev,
		Results:  make([]internal.ShowMeta, 0, len(showRows)),
	}
	for _, row := range showRows {
		showResults.Results = append(showResults.Results, internal.RowToShowMeta(row))
	}
	respondWithJSON(w, http.StatusOK, showResults)
}

func (s *server) handleGetShowsFromTodaysDate(w http.ResponseWriter, r *http.Request) {
	limit, offset, err := parsePagination(r)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error(), err)
		return
	}

	showRows, err := s.queries.GetShowsFromTodaysDate(r.Context(), database.GetShowsFromTodaysDateParams{
		PageOffset: int32(offset),
		PageLimit:  int32(limit),
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not get shows", err)
		return
	}

	if len(showRows) == 0 && offset > 0 {
		respondWithError(w, http.StatusBadRequest, "Offset out of range", nil)
		return
	}
	if len(showRows) == 0 && offset == 0 {
		respondWithError(w, http.StatusNotFound, "No shows found on today's date", nil)
		return
	}

	next, prev := buildLinks(r, int(showRows[0].Count), offset, limit)

	results := internal.Paginated[internal.ShowMeta]{
		Count:    showRows[0].Count,
		Next:     next,
		Previous: prev,
		Results:  make([]internal.ShowMeta, 0, len(showRows)),
	}
	for _, row := range showRows {
		results.Results = append(results.Results, internal.RowToShowMeta(row))
	}
	respondWithJSON(w, http.StatusOK, results)
}

func (s *server) handleGetAcousticShows(w http.ResponseWriter, r *http.Request) {
	limit, offset, err := parsePagination(r)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error(), err)
		return
	}

	showRows, err := s.queries.GetAcousticShows(r.Context(), database.GetAcousticShowsParams{
		PageOffset: int32(offset),
		PageLimit:  int32(limit),
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not get shows", err)
		return
	}

	if len(showRows) == 0 && offset > 0 {
		respondWithError(w, http.StatusBadRequest, "Offset out of range", nil)
		return
	}
	if len(showRows) == 0 && offset == 0 {
		respondWithError(w, http.StatusInternalServerError, "Malformed show data", nil)
	}

	next, prev := buildLinks(r, int(showRows[0].Count), offset, limit)
	results := internal.Paginated[internal.ShowMeta]{
		Count:    showRows[0].Count,
		Next:     next,
		Previous: prev,
		Results:  make([]internal.ShowMeta, 0, len(showRows)),
	}
	for _, row := range showRows {
		results.Results = append(results.Results, internal.RowToShowMeta(row))
	}

	respondWithJSON(w, http.StatusOK, results)
}

func (s *server) handleGetShowsWithSongDebuts(w http.ResponseWriter, r *http.Request) {
	limit, offset, err := parsePagination(r)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error(), err)
		return
	}

	showRows, err := s.queries.GetSongDebuts(r.Context(), database.GetSongDebutsParams{
		PageOffset: int32(offset),
		PageLimit:  int32(limit),
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not get shows", err)
		return
	}
	if len(showRows) == 0 && offset > 0 {
		respondWithError(w, http.StatusBadRequest, "Offset out of range", nil)
		return
	}
	if len(showRows) == 0 && offset == 0 {
		respondWithError(w, http.StatusInternalServerError, "Malformed show data", nil)
		return
	}

	next, prev := buildLinks(r, int(showRows[0].Count), offset, limit)
	results := internal.Paginated[internal.ShowMeta]{
		Count:    showRows[0].Count,
		Next:     next,
		Previous: prev,
		Results:  make([]internal.ShowMeta, 0, len(showRows)),
	}
	for _, row := range showRows {
		results.Results = append(results.Results, internal.RowToShowMeta(row))
	}

	respondWithJSON(w, http.StatusOK, results)
}

func (s *server) handleGetAllShows(w http.ResponseWriter, r *http.Request) {
	limit, offset, err := parsePagination(r)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error(), err)
		return
	}

	showRows, err := s.queries.GetAllShows(r.Context(), database.GetAllShowsParams{
		PageOffset: int32(offset),
		PageLimit:  int32(limit),
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not get shows", err)
		return
	}

	if len(showRows) == 0 && offset > 0 {
		respondWithError(w, http.StatusBadRequest, "Offset out of range", nil)
		return
	}
	if len(showRows) == 0 && offset == 0 {
		respondWithError(w, http.StatusInternalServerError, "Could not get shows", nil)
		return
	}

	next, prev := buildLinks(r, int(showRows[0].Count), offset, limit)
	results := internal.Paginated[internal.ShowMeta]{
		Count:    showRows[0].Count,
		Next:     next,
		Previous: prev,
		Results:  make([]internal.ShowMeta, 0, len(showRows)),
	}
	for _, row := range showRows {
		results.Results = append(results.Results, internal.RowToShowMeta(row))
	}

	respondWithJSON(w, http.StatusOK, results)
}

func (s *server) isValidLocation(ctx context.Context, location string) (bool, error) {
	if location == "" {
		return true, nil
	}
	location = strings.ToLower(location)
	validLocations, err := s.queries.GetValidLocations(ctx)
	if err != nil {
		return false, err
	}
	if len(validLocations) == 0 {
		return false, errors.New("valid locations list empty")
	}
	if !slices.Contains(validLocations, location) {
		return false, errors.New("invalid location")
	}
	return true, nil
}

func (s *server) handleSearchShows(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	filters := internal.ShowSearchFilters{
		Year:     query.Get("year"),
		Month:    query.Get("month"),
		Day:      query.Get("day"),
		Location: query.Get("location"),
		City:     query.Get("city"),
		Song:     query.Get("song"),
		Venue:    query.Get("venue"),
		Notes:    query.Get("notes"),
		SetName:  query.Get("set_name"),
		HasNotes: query.Get("has_notes"),
	}

	dateParts, err := validateDateParts(filters.Year, filters.Month, filters.Day)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error(), err)
		return
	}

	location := strings.TrimSpace(strings.ToLower(filters.Location))
	ok, err := s.isValidLocation(r.Context(), location)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error(), err)
		return
	}
	if !ok {
		respondWithError(w, http.StatusBadRequest, "invalid location", nil)
		return
	}

	setName := strings.TrimSpace(strings.ToLower(filters.SetName))
	validSetName, err := isValidSetName(setName)
	if !validSetName || err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error(), err)
		return
	}

	hasNotes, err := strToNullBool(filters.HasNotes)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid has_notes parameter", err)
		return
	}

	limit, offset, err := parsePagination(r)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error(), err)
		return
	}

	showRows, err := s.queries.SearchShowsWithFilters(r.Context(), database.SearchShowsWithFiltersParams{
		Year:       sql.NullInt32{Int32: dateParts.year, Valid: filters.Year != ""},
		Month:      sql.NullInt32{Int32: dateParts.month, Valid: filters.Month != ""},
		Day:        sql.NullInt32{Int32: dateParts.day, Valid: filters.Day != ""},
		Location:   sql.NullString{String: location, Valid: filters.Location != ""},
		City:       sql.NullString{String: fuzzyPattern(filters.City), Valid: filters.City != ""},
		Song:       sql.NullString{String: fuzzyPattern(filters.Song), Valid: filters.Song != ""},
		Venue:      sql.NullString{String: fuzzyPattern(filters.Venue), Valid: filters.Venue != ""},
		Notes:      sql.NullString{String: fuzzyPattern(filters.Notes), Valid: filters.Notes != ""},
		SetName:    sql.NullString{String: setName, Valid: filters.SetName != ""},
		HasNotes:   hasNotes,
		PageOffset: int32(offset),
		PageLimit:  int32(limit),
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "could not get shows", err)
		return
	}

	if len(showRows) == 0 && offset > 0 {
		respondWithError(w, http.StatusBadRequest, "offset out of range", nil)
		return
	}
	if len(showRows) == 0 && offset == 0 {
		respondWithError(w, http.StatusNotFound, "no shows found", err)
		return
	}

	next, prev := buildLinks(r, int(showRows[0].Count), offset, limit)
	results := internal.Paginated[internal.ShowMeta]{
		Count:    showRows[0].Count,
		Next:     next,
		Previous: prev,
		Results:  make([]internal.ShowMeta, 0, len(showRows)),
	}
	for _, row := range showRows {
		results.Results = append(results.Results, internal.RowToShowMeta(row))
	}

	respondWithJSON(w, http.StatusOK, results)
}

func (s *server) handleGetVenueStats(w http.ResponseWriter, r *http.Request) {
	venueStatRows, err := s.queries.ListVenueStats(r.Context())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not get venue stats", err)
		return
	}
	results := []internal.ListVenueStatsRow{}
	for _, row := range venueStatRows {
		results = append(results, internal.ListVenueStatsRow{
			City:        row.City,
			Location:    row.Location,
			Venue:       row.Venue,
			ShowsPlayed: row.ShowsPlayed,
		})
	}
	respondWithJSON(w, http.StatusOK, results)
}

func validateDateParts(year, month, day string) (dateFilters, error) {
	df := dateFilters{}
	if year != "" {
		y, err := strconv.ParseInt(year, 10, 32)
		if err != nil {
			return dateFilters{}, errors.New("invalid year parameter")
		}
		if y < 1965 || y > 1995 {
			return dateFilters{}, errors.New("year out of range")
		}
		df.year = int32(y)
	}

	if month != "" {
		m, err := strconv.ParseInt(month, 10, 32)
		if err != nil {
			return dateFilters{}, errors.New("invalid month parameter")
		}
		if m < 1 || m > 12 {
			return dateFilters{}, errors.New("month out of range")
		}
		df.month = int32(m)
	}

	if day != "" {
		d, err := strconv.ParseInt(day, 10, 32)
		if err != nil {
			return dateFilters{}, errors.New("invalid day parameter")
		}
		if d < 1 || d > 31 {
			return dateFilters{}, errors.New("day out of range")
		}
		df.day = int32(d)
	}
	date := time.Date(
		int(df.year),
		time.Month(df.month),
		int(df.day),
		0, 0, 0, 0,
		time.UTC,
	)
	if df.year > 0 && df.month > 0 && df.day > 0 {
		if date.Year() != int(df.year) ||
			date.Month() != time.Month(df.month) ||
			date.Day() != int(df.day) {
			return dateFilters{}, errors.New("invalid date")
		}
	}
	return df, nil
}

func isValidSetName(setName string) (bool, error) {
	if setName == "" {
		return true, nil
	}
	setName = strings.ToLower(setName)
	validSetNames := []string{"set_1", "set_2", "set_3", "encore", "acoustic", "electric", "acoustic_1", "acoustic_2"}
	if !slices.Contains(validSetNames, setName) {
		return false, errors.New("invalid set name")
	}
	return true, nil
}

func strToNullBool(input string) (sql.NullBool, error) {
	if input == "" {
		return sql.NullBool{}, nil
	}

	b, err := strconv.ParseBool(input)
	if err != nil {
		return sql.NullBool{}, err
	}

	return sql.NullBool{Bool: b, Valid: true}, nil
}
