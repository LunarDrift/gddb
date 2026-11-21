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
	req := httptest.NewRequest(http.MethodGet, "/shows/", nil)
	req.SetPathValue("showID", "1")
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
		{"missing id", "/shows/", http.StatusBadRequest, &fakeQuerier{}},
		{"invalid id", "/shows/six", http.StatusBadRequest, &fakeQuerier{}},
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
	show := []database.SearchShowsWithFiltersRow{{ShowID: 1, ShowDate: date, Weekday: sql.NullString{String: "fri", Valid: true}, Venue: "Venue", City: "New York", Location: "NY", Notes: sql.NullString{String: "Mime Troupe Benefit", Valid: true}}}
	fake := &fakeQuerier{searchShowsFiltersRows: show}
	tests := []struct {
		name string
		url  string
		fake *fakeQuerier
	}{
		{name: "year only", url: "/shows?year=1970", fake: fake},
		{name: "month only", url: "/shows?month=10", fake: fake},
		{name: "day only", url: "/shows?day=10", fake: fake},
		{name: "weekday only", url: "/shows?weekday=fri", fake: fake},
		{name: "year and month", url: "/shows?year=1970&month=3", fake: fake},
		{name: "month and day", url: "/shows?month=3&day=10", fake: fake},
		{name: "year and day", url: "/shows?year=1970&day=10", fake: fake},
		{name: "song search", url: "/shows?song=althea", fake: fake},
		{name: "city search", url: "/shows?city=york", fake: fake},
		{name: "notes search", url: "/shows?notes=benefit", fake: fake},
		{name: "year and city", url: "/shows?year=1970&city=york", fake: fake},
		{name: "song and city", url: "/shows?city=york&song=althea", fake: fake},
		{name: "set name", url: "/shows?set_name=acoustic", fake: fake},
		{name: "has_notes=true", url: "/shows?has_notes=true", fake: fake},
		{name: "has_notes=false", url: "/shows?has_notes=false", fake: fake},
		{name: "no parameters given", url: "/shows?year=&month=&day=&weekday=&location=&city=&song=&venue=&notes=&set_name=&has_notes=", fake: fake},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &server{queries: tt.fake}
			req := httptest.NewRequest(http.MethodGet, tt.url, nil)
			w := httptest.NewRecorder()
			s.handleSearchShows(w, req)
			res := w.Result()
			if res.StatusCode != http.StatusOK {
				t.Fatalf("status code = %d; want %d", res.StatusCode, http.StatusOK)
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

func TestHandleSearchShows_Errors(t *testing.T) {
	tests := []struct {
		name       string
		url        string
		wantStatus int
		fake       *fakeQuerier
	}{
		{name: "no shows found / too specific search", url: "/shows?year=1980&month=4&weekday=tue&city=chicago&song=dark+star", wantStatus: http.StatusNotFound, fake: &fakeQuerier{}},
		{name: "invalid month parameter", url: "/shows?month=april", wantStatus: http.StatusBadRequest, fake: &fakeQuerier{}},
		{name: "year out of range", url: "/shows?year=2000", wantStatus: http.StatusBadRequest, fake: &fakeQuerier{}},
		{name: "invalid weekday", url: "/shows?weekday=monday", wantStatus: http.StatusBadRequest, fake: &fakeQuerier{}},
		{name: "server error", url: "/shows?year=1982&city=new+york", wantStatus: http.StatusInternalServerError, fake: &fakeQuerier{searchShowsFiltersErr: errors.New("db exploded")}},
	}

	for _, tt := range tests {
		s := &server{queries: tt.fake}
		req := httptest.NewRequest(http.MethodGet, tt.url, nil)
		w := httptest.NewRecorder()
		s.handleSearchShows(w, req)
		res := w.Result()
		if res.StatusCode != tt.wantStatus {
			t.Errorf("status code = %d; want %d", res.StatusCode, tt.wantStatus)
		}
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
