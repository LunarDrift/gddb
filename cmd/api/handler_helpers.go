package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/LunarDrift/deadabase/internal"
)

// pageURL, buildLinks, and parsePagination are helpers for pagination process

// pageURL: Add 'offset' and 'limit' query parameters to URL
func pageURL(r *http.Request, offset, limit int) *string {
	q := r.URL.Query()
	q.Set("offset", strconv.Itoa(offset))
	q.Set("limit", strconv.Itoa(limit))
	s := r.URL.Path + "?" + q.Encode()
	return &s
}

// buildLinks: Builds the Paginated response 'Next' and 'Previous' links
// based on the total number of results ('count') and current 'offset' and 'limit'
func buildLinks(r *http.Request, count, offset, limit int) (next, prev *string) {
	if offset+limit < count {
		next = pageURL(r, offset+limit, limit)
	}
	if offset > 0 {
		prev = pageURL(r, max(offset-limit, 0), limit)
	}
	return next, prev
}

// parsePagination: Sets default 'limit' and 'offset' query parameters if none are given.
// Default limit of 20. Max limit of 100.
func parsePagination(r *http.Request) (limit, offset int, err error) {
	limit, offset = 20, 0
	q := r.URL.Query()

	if v := q.Get("limit"); v != "" {
		if limit, err = strconv.Atoi(v); err != nil || limit < 1 {
			return 0, 0, errors.New("invalid 'limit' parameter")
		}
		limit = min(limit, 100) // cap so server doesn't send everything at once
	}
	if v := q.Get("offset"); v != "" {
		if offset, err = strconv.Atoi(v); err != nil || offset < 0 {
			return 0, 0, errors.New("invalid 'offset' parameter")
		}
	}
	return limit, offset, nil
}

// fuzzyPattern wraps and inserts a '%' between every word of `input` to be used during SQL ILIKE query searches
// e.g. "Help on the way > Slipknot! > Franklin's Tower" becomes "%Help%On%The%Way%>..."
func fuzzyPattern(input string) string {
	words := strings.Fields(input)
	if len(words) == 0 {
		return ""
	}
	return "%" + strings.Join(words, "%") + "%"
}

func respondWithJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		slog.SetLogLoggerLevel(slog.LevelError)
		log.Println("respondWithJSON - could not encode payload:", err)
	}
}

func respondWithError(w http.ResponseWriter, statusCode int, message string, err error) {
	if err != nil {
		slog.SetLogLoggerLevel(slog.LevelError)
		log.Println(err)
	}
	type errorResponse struct {
		Status string `json:"status"`
		Code   int    `json:"code"`
		Error  string `json:"error"`
	}
	respondWithJSON(w, statusCode, errorResponse{Status: strings.ToUpper(http.StatusText(statusCode)), Code: statusCode, Error: message})
}

// buildShowResponse takes a slice of ShowSortInput rows for a single show and returns
// a 'no setlist available' message if the show has no songs, otherwise sorting the
// sets + songs and attaching any footnotes, returning the completed show response.
func (s *server) buildShowResponse(r *http.Request, parsedShowRows []internal.ShowSortInput) (any, error) {
	if len(parsedShowRows) > 0 && parsedShowRows[0].RawEntry == "" {
		row := parsedShowRows[0]
		return internal.ShowWithNoSetlist{
			ShowID:   row.ShowID,
			Date:     row.Date.Format(time.DateOnly),
			Venue:    row.Venue,
			City:     row.City,
			Location: row.Location,
			Notes:    row.Notes,
			Message:  "No setlist available for this show",
		}, nil
	}

	showResp, err := internal.SortSetPositions(parsedShowRows)
	if err != nil {
		return nil, err
	}

	footnoteRows, err := s.queries.GetFootnotesFromShowID(r.Context(), parsedShowRows[0].ShowID)
	if err != nil {
		return nil, err
	}
	showResp.Footnotes = make(map[string]string)
	for _, f := range footnoteRows {
		showResp.Footnotes[f.Marker] = f.NoteText
	}

	return showResp, nil
}

func validateDateParts(year, month, day string) (dateFilters, error) {
	df := dateFilters{}

	if year != "" {
		y, err := strconv.ParseInt(year, 10, 32)
		if err != nil {
			return dateFilters{}, errors.New("invalid year parameter: " + year)
		}
		if y < 1965 || y > 1995 {
			return dateFilters{}, errors.New("year out of range: " + year)
		}
		df.year = int32(y)
	}

	if month != "" {
		m, err := strconv.ParseInt(month, 10, 32)
		if err != nil {
			return dateFilters{}, errors.New("invalid month parameter: " + month)
		}
		if m < 1 || m > 12 {
			return dateFilters{}, errors.New("month out of range: " + month)
		}
		df.month = int32(m)
	}

	if day != "" {
		d, err := strconv.ParseInt(day, 10, 32)
		if err != nil {
			return dateFilters{}, errors.New("invalid day parameter: " + day)
		}
		if d < 1 || d > 31 {
			return dateFilters{}, errors.New("day out of range: " + day)
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
			return dateFilters{}, fmt.Errorf("invalid date: %v", date)
		}
	}

	return df, nil
}

func isValidSetName(setName string) bool {
	// no set_name parameter is still valid - becomes NULL in SQL query and doesn't filter based on set name
	if setName == "" {
		return true
	}
	validSetNames := []string{"set_1", "set_2", "set_3", "encore", "acoustic", "electric", "acoustic_1", "acoustic_2"}
	return slices.Contains(validSetNames, setName)
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

func (s *server) isValidLocation(ctx context.Context, location string) (bool, error) {
	// no location parameter is still valid - becomes NULL in SQL query and doesn't filter based on location
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
		return false, nil
	}
	return true, nil
}
