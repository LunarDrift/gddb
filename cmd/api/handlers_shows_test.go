package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/LunarDrift/deadabase/internal"
	"github.com/LunarDrift/deadabase/internal/database"
)

func TestHandleGetShowFromID(t *testing.T) {
	date, _ := time.Parse(time.DateOnly, "1970-04-20")
	fake := &fakeQuerier{
		showFromIDRows: []database.GetShowDetailsFromIDRow{{ShowID: 1, ShowDate: date, Venue: "venue", City: "city", Location: "location", Notes: sql.NullString{}}},
	}

	s := &server{queries: fake}
	req := httptest.NewRequest(http.MethodGet, "/shows?id=1", nil)
	w := httptest.NewRecorder()
	s.handleGetShowFromID(w, req)
	res := w.Result()
	if res.StatusCode != http.StatusOK {
		t.Errorf("status code = %d; want %d", res.StatusCode, http.StatusOK)
	}

	got := internal.ShowResponse{}
	if err := json.NewDecoder(res.Body).Decode(&got); err != nil {
		t.Fatalf("error decoding response: %v", err)
	}

	if got.ShowID != 1 {
		t.Errorf("got.ShowID = %d; want 1", got.ShowID)
	}
}

func TestHandleGetShowFromID_Errors(t *testing.T) {
	tests := []struct {
		name       string
		url        string
		wantStatus int
		fake       *fakeQuerier
	}{
		{"missing id", "/shows?id=", http.StatusBadRequest, &fakeQuerier{}},
		{"invalid id", "/shows?id=six", http.StatusBadRequest, &fakeQuerier{}},
		{"server error", "/shows?id=555", http.StatusInternalServerError, &fakeQuerier{showFromIDErr: errors.New("db exploded")}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &server{queries: tt.fake}
			req := httptest.NewRequest(http.MethodGet, tt.url, nil)
			w := httptest.NewRecorder()
			s.handleGetShowFromID(w, req)
			if got := w.Result().StatusCode; got != tt.wantStatus {
				t.Errorf("status code = %d; want %d", got, tt.wantStatus)
			}
		})
	}
}

func TestHandleGetShowsFromTodaysDate(t *testing.T) {
	date1, _ := time.Parse(time.DateOnly, "1973-11-05")
	date2, _ := time.Parse(time.DateOnly, "1978-11-05")

	fake := &fakeQuerier{
		showsFromTodaysDateRows: []database.GetShowsFromTodaysDateRow{
			{ShowID: 1, ShowDate: date1, Venue: "Soldier Field", City: "Chicago", Location: "IL", Notes: sql.NullString{}},
			{ShowID: 2, ShowDate: date2, Venue: "Another venue", City: "New York", Location: "NY", Notes: sql.NullString{}},
		},
	}

	s := &server{queries: fake}
	req := httptest.NewRequest(http.MethodGet, "/shows/today", nil)
	w := httptest.NewRecorder()
	s.handleGetShowsFromTodaysDate(w, req)
	res := w.Result()
	if res.StatusCode != http.StatusOK {
		t.Errorf("status code = %d; want %d", res.StatusCode, http.StatusOK)
	}

	got := internal.Paginated[internal.ShowMeta]{}
	if err := json.NewDecoder(res.Body).Decode(&got); err != nil {
		t.Fatalf("error decoding response: %v", err)
	}

	if len(got.Results) != 2 {
		t.Errorf("len(got.Results) = %d; want 2", len(got.Results))
	}
}

func TestHandleGetShowsBetweenDates(t *testing.T) {
	startDate, _ := time.Parse(time.DateOnly, "1970-01-01")
	endDate, _ := time.Parse(time.DateOnly, "1970-01-31")
	fake := &fakeQuerier{
		showsBetweenDatesRows: []database.GetShowsBetweenDatesRow{
			{ShowID: 1, ShowDate: startDate, Venue: "test venue", City: "test city", Location: "test location", Notes: sql.NullString{}},
			{ShowID: 2, ShowDate: endDate, Venue: "test venue", City: "test city", Location: "test location", Notes: sql.NullString{}},
		},
	}

	s := &server{queries: fake}
	req := httptest.NewRequest(http.MethodGet, "/shows?start_date=1970-01-01&end_date=1970-01-31", nil)
	w := httptest.NewRecorder()

	s.handleGetShowsBetweenDates(w, req)

	res := w.Result()
	if res.StatusCode != http.StatusOK {
		t.Errorf("status code = %d; want %d", res.StatusCode, http.StatusOK)
	}

	got := internal.Paginated[internal.ShowMeta]{}
	if err := json.NewDecoder(res.Body).Decode(&got); err != nil {
		t.Fatalf("error decoding response: %v", err)
	}

	if len(got.Results) != len(fake.showsBetweenDatesRows) {
		t.Errorf("len(got) = %d; want %d", len(got.Results), len(fake.showsBetweenDatesRows))
	}

	if got.Results[0].Date != "1970-01-01" {
		t.Errorf("got.Results[0].Date = %q; want '1970-01-01'", got.Results[0].Date)
	}
}

func TestHandleGetShowsBetweenDates_Errors(t *testing.T) {
	tests := []struct {
		name       string
		url        string
		wantStatus int
		fake       *fakeQuerier
	}{
		{"start date after end date", "/shows?start_date=1980-09-09&end_date=1965-06-03", http.StatusBadRequest, &fakeQuerier{}},
		{"missing start date", "/shows?start_date=&end_date=1990-01-01", http.StatusBadRequest, &fakeQuerier{}},
		{"missing end date", "/shows?start_date=1970-01-01&end_date=", http.StatusBadRequest, &fakeQuerier{}},
		{"server error", "/shows?start_date=1990-01-01&end_date=1990-05-01", http.StatusInternalServerError, &fakeQuerier{showsBetweenDatesErr: errors.New("db exploded")}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &server{queries: tt.fake}
			req := httptest.NewRequest(http.MethodGet, tt.url, nil)
			w := httptest.NewRecorder()
			s.handleGetShowsBetweenDates(w, req)
			if got := w.Result().StatusCode; got != tt.wantStatus {
				t.Errorf("status code = %d; want %d", got, tt.wantStatus)
			}
		})
	}
}

func TestHandleGetRandomShow_NoShowsAvailable(t *testing.T) {
	fake := &fakeQuerier{
		allShowIDs: []int32{},
	}
	s := &server{queries: fake}

	req := httptest.NewRequest(http.MethodGet, "/shows/random", nil)
	w := httptest.NewRecorder()

	s.handleGetRandomShow(w, req)

	res := w.Result()
	if res.StatusCode != http.StatusInternalServerError {
		t.Errorf("status code = %d; want %d", res.StatusCode, http.StatusInternalServerError)
	}
}

func TestHandleSearchShows(t *testing.T) {
	date, _ := time.Parse(time.DateOnly, "1970-03-10")
	show := []database.SearchShowsWithFiltersRow{{ShowID: 1, ShowDate: date, Venue: "Venue", City: "New York", Location: "NY", Notes: sql.NullString{String: "Mime Troupe Benefit", Valid: true}}}
	fake := &fakeQuerier{searchShowsFiltersRows: show}
	tests := []struct {
		name       string
		url        string
		wantStatus int
		fake       *fakeQuerier
	}{
		{"year only", "/shows?year=1970", http.StatusOK, fake},
		{"month only", "/shows?month=10", http.StatusOK, fake},
		{"day only", "/shows?day=10", http.StatusOK, fake},
		{"year and month", "/shows?year=1970&month=3", http.StatusOK, fake},
		{"month and day", "/shows?month=3&day=10", http.StatusOK, fake},
		{"year and day", "/shows?year=1970&day=10", http.StatusOK, fake},
		{"song search", "/shows?song=althea", http.StatusOK, fake},
		{"city search", "/shows?city=york", http.StatusOK, fake},
		{"notes search", "/shows?notes=benefit", http.StatusOK, fake},
		{"year and city", "/shows?year=1970&city=york", http.StatusOK, fake},
		{"song and city", "/shows?city=york&song=althea", http.StatusOK, fake},
		{"set name", "/shows?set_name=acoustic", http.StatusOK, fake},
		{"has_notes=true", "/shows?has_notes=true", http.StatusOK, fake},
		{"has_notes=false", "/shows?has_notes=false", http.StatusOK, fake},
		{"no parameters given", "/shows?year=&month=&day=&location=&city=&song=&venue=&notes=&set_name=&has_notes=", http.StatusOK, fake},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &server{queries: tt.fake}
			req := httptest.NewRequest(http.MethodGet, tt.url, nil)
			w := httptest.NewRecorder()
			s.handleSearchShows(w, req)
			res := w.Result()
			if res.StatusCode != tt.wantStatus {
				t.Fatalf("status code = %d; want %d", res.StatusCode, tt.wantStatus)
			}

			got := internal.Paginated[internal.ShowMeta]{}
			if err := json.NewDecoder(res.Body).Decode(&got); err != nil {
				t.Fatalf("error decoding response: %v", err)
			}

			if len(got.Results) != len(show) {
				t.Errorf("len(got.Results) = %d; want %d", len(got.Results), len(show))
			}
			if got.Results[0].ShowID != show[0].ShowID {
				t.Errorf("got.ShowID = %d; want %d", got.Results[0].ShowID, show[0].ShowID)
			}
		})
	}
}

func TestHandleGetShowsWithSongDebuts(t *testing.T) {
	date, _ := time.Parse(time.DateOnly, "1995-09-04")
	fake := &fakeQuerier{
		getSongDebutsRows: []database.GetSongDebutsRow{
			{ShowID: 1, ShowDate: date, Venue: "Venue", City: "New York", Location: "NY", Notes: sql.NullString{}},
		},
	}
	s := &server{queries: fake}
	req := httptest.NewRequest(http.MethodGet, "/shows/song-debuts", nil)
	w := httptest.NewRecorder()
	s.handleGetShowsWithSongDebuts(w, req)
	res := w.Result()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("status code = %d; want %d", res.StatusCode, http.StatusOK)
	}

	got := internal.Paginated[internal.ShowMeta]{}
	if err := json.NewDecoder(res.Body).Decode(&got); err != nil {
		t.Fatalf("error decoding response: %v", err)
	}

	if len(got.Results) != 1 {
		t.Errorf("len(got.Results) = %d; want 1", len(got.Results))
	}

	if got.Results[0].ShowID != 1 {
		t.Errorf("got.Results[0].ShowID = %d; want 1", got.Results[0].ShowID)
	}
}

func TestHandleGetAllAcousticShows(t *testing.T) {
	date, _ := time.Parse(time.DateOnly, "1995-09-04")
	fake := &fakeQuerier{
		getAcousticShowsRows: []database.GetAcousticShowsRow{
			{ShowID: 1, ShowDate: date, Venue: "Venue", City: "New York", Location: "NY", Notes: sql.NullString{}},
		},
	}
	s := &server{queries: fake}
	req := httptest.NewRequest(http.MethodGet, "/shows/acoustic", nil)
	w := httptest.NewRecorder()
	s.handleGetAcousticShows(w, req)
	res := w.Result()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("status code = %d; want %d", res.StatusCode, http.StatusOK)
	}

	got := internal.Paginated[internal.ShowMeta]{}
	if err := json.NewDecoder(res.Body).Decode(&got); err != nil {
		t.Fatalf("error decoding response: %v", err)
	}

	if len(got.Results) != 1 {
		t.Errorf("len(got.Results) = %d; want 1", len(got.Results))
	}
}
