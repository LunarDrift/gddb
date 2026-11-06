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

func TestHandleSearchShowsFromDate(t *testing.T) {
	date, _ := time.Parse(time.DateOnly, "1995-01-01")
	tests := []struct {
		name string
		url  string
		fake *fakeQuerier
	}{
		{
			"full date", "/shows?year=1995&month=1&day=1", &fakeQuerier{
				showsFromDateRows: []database.GetShowsFromDateRow{{ShowID: 1, ShowDate: date, Venue: "Soldier Field", City: "Chicago", Location: "IL", Notes: sql.NullString{}}},
			},
		},
		{
			"year only", "/shows?year=1995&month=&date=", &fakeQuerier{
				showsFromDateRows: []database.GetShowsFromDateRow{{ShowID: 1, ShowDate: date, Venue: "Soldier Field", City: "Chicago", Location: "IL", Notes: sql.NullString{}}},
			},
		},
		{
			"month only", "/shows?year=&month=1&day=", &fakeQuerier{
				showsFromDateRows: []database.GetShowsFromDateRow{{ShowID: 1, ShowDate: date, Venue: "Soldier Field", City: "Chicago", Location: "IL", Notes: sql.NullString{}}},
			},
		},
		{
			"day only", "/shows?year=&month=&day=1", &fakeQuerier{
				showsFromDateRows: []database.GetShowsFromDateRow{{ShowID: 1, ShowDate: date, Venue: "Soldier Field", City: "Chicago", Location: "IL", Notes: sql.NullString{}}},
			},
		},
		{
			"year and month", "/shows?year=1995&month=1&day=", &fakeQuerier{
				showsFromDateRows: []database.GetShowsFromDateRow{{ShowID: 1, ShowDate: date, Venue: "Soldier Field", City: "Chicago", Location: "IL", Notes: sql.NullString{}}},
			},
		},
		{
			"year and day", "/shows?year=1995&month=&day=1", &fakeQuerier{
				showsFromDateRows: []database.GetShowsFromDateRow{{ShowID: 1, ShowDate: date, Venue: "Soldier Field", City: "Chicago", Location: "IL", Notes: sql.NullString{}}},
			},
		},
		{
			"month and day", "/shows?year=&month=1&day=1", &fakeQuerier{
				showsFromDateRows: []database.GetShowsFromDateRow{{ShowID: 1, ShowDate: date, Venue: "Soldier Field", City: "Chicago", Location: "IL", Notes: sql.NullString{}}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &server{queries: tt.fake}
			req := httptest.NewRequest(http.MethodGet, tt.url, nil)
			w := httptest.NewRecorder()
			s.handleSearchShowsFromDate(w, req)

			res := w.Result()
			if res.StatusCode != http.StatusOK {
				t.Errorf("status code = %d; want %d", res.StatusCode, http.StatusOK)
			}

			got := internal.Paginated[internal.ShowMeta]{}
			if err := json.NewDecoder(res.Body).Decode(&got); err != nil {
				t.Fatalf("error decoding response: %v", err)
			}

			if len(got.Results) != len(tt.fake.showsFromDateRows) {
				t.Errorf("len(got) = %d; want %d", len(got.Results), len(tt.fake.showsFromDateRows))
			}

			show := got.Results[0]
			if show.Date != "1995-01-01" {
				t.Errorf("show.Date = %q; want '1995-01-01", show.Date)
			}
		})
	}
}

func TestHandleSearchShowsFromDate_Errors(t *testing.T) {
	tests := []struct {
		name       string
		url        string
		wantStatus int
		fake       *fakeQuerier
	}{
		{"invalid year parameter", "/shows?year=abc&month=&day=", http.StatusBadRequest, &fakeQuerier{}},
		{"invalid month parameter", "/shows?year=&month=15&day=", http.StatusBadRequest, &fakeQuerier{}},
		{"invalid day parameter", "/shows?year=&month=&day=420", http.StatusBadRequest, &fakeQuerier{}},
		{"server error", "/shows?year=1970&month=&day=", http.StatusInternalServerError, &fakeQuerier{showsFromDateErr: errors.New("db exploded")}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &server{queries: tt.fake}
			req := httptest.NewRequest(http.MethodGet, tt.url, nil)
			w := httptest.NewRecorder()
			s.handleSearchShowsFromDate(w, req)
			if got := w.Result().StatusCode; got != tt.wantStatus {
				t.Errorf("status code = %d; want %d", got, tt.wantStatus)
			}
		})
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

func TestHandleShowsBetweenDates_StartDateAfterEndDate(t *testing.T) {
	// empty querier - validation should reject the request before ever getting to query step
	fake := &fakeQuerier{}

	s := &server{queries: fake}
	req := httptest.NewRequest(http.MethodGet, "/shows?start_date=1980-09-01&end_date=1980-02-01", nil)
	w := httptest.NewRecorder()

	s.handleGetShowsBetweenDates(w, req)

	res := w.Result()
	if res.StatusCode != http.StatusBadRequest {
		t.Errorf("status code = %d; want %d", res.StatusCode, http.StatusBadRequest)
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

func TestHandleGetShowsFromSongName(t *testing.T) {
	date, _ := time.Parse(time.DateOnly, "1995-01-01")
	fake := &fakeQuerier{
		showsFromSongNameRows: []database.GetShowsFromSongNameRow{
			{ShowID: 1, ShowDate: date, Venue: "test venue", City: "test city", Location: "test location", Notes: sql.NullString{}},
			{ShowID: 2, ShowDate: date.Add(time.Hour * 48), Venue: "test venue 2", City: "test city 2", Location: "test location 2", Notes: sql.NullString{}},
		},
	}

	s := &server{queries: fake}
	req := httptest.NewRequest(http.MethodGet, "/shows?song=althea", nil)
	w := httptest.NewRecorder()

	s.handleGetShowsFromSongName(w, req)

	res := w.Result()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("status code = %d; want %d", res.StatusCode, http.StatusOK)
	}

	got := internal.Paginated[internal.ShowMeta]{}
	if err := json.NewDecoder(res.Body).Decode(&got); err != nil {
		t.Fatalf("error decoding response: %v", err)
	}

	if len(got.Results) != len(fake.showsFromSongNameRows) {
		t.Errorf("len(got) = %d; want %d", len(got.Results), len(fake.showsFromSongNameRows))
	}

	if got.Results[0].Date != "1995-01-01" {
		t.Errorf("got.Results[0].Date = %q; want %q", got.Results[0].Date, "1995-01-01")
	}

	if got.Results[0].Venue != "test venue" {
		t.Errorf("got.Results[0].Venue = %q; want 'test venue'", got.Results[0].Venue)
	}
	if got.Results[1].Venue != "test venue 2" {
		t.Errorf("got.Results[1].Venue = %q; want 'test venue 2'", got.Results[1].Venue)
	}
}

func TestHandleGetShowsFromSongName_Errors(t *testing.T) {
	tests := []struct {
		name       string
		url        string
		wantStatus int
		fake       *fakeQuerier
	}{
		{"missing song name", "/shows?song=", http.StatusBadRequest, &fakeQuerier{}},
		{"invalid song name", "/shows?song=Drak_Star", http.StatusNotFound, &fakeQuerier{}},
		{"server error", "/shows?song=althea", http.StatusInternalServerError, &fakeQuerier{showsFromSongNameErr: errors.New("db exploded")}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &server{queries: tt.fake}
			req := httptest.NewRequest(http.MethodGet, tt.url, nil)
			w := httptest.NewRecorder()
			s.handleGetShowsFromSongName(w, req)
			if got := w.Result().StatusCode; got != tt.wantStatus {
				t.Errorf("status code = %d; want %d", got, tt.wantStatus)
			}
		})
	}
}

func TestHandleGetShowsFromSetName(t *testing.T) {
	date, _ := time.Parse(time.DateOnly, "1995-01-01")
	fake := &fakeQuerier{
		showsFromSetNameRows: []database.GetShowsFromSetNameRow{
			{ShowID: 1, ShowDate: date, Venue: "Soldier Field", City: "Chicago", Location: "test location", Notes: sql.NullString{}},
		},
	}

	s := &server{queries: fake}
	req := httptest.NewRequest(http.MethodGet, "/shows?set_name=encore", nil)
	w := httptest.NewRecorder()

	s.handleGetShowsFromSetName(w, req)

	res := w.Result()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("status code = %d; want %d", res.StatusCode, http.StatusOK)
	}

	got := internal.Paginated[internal.ShowMeta]{}
	if err := json.NewDecoder(res.Body).Decode(&got); err != nil {
		t.Fatalf("error decoding response: %v", err)
	}

	if got.Results[0].ShowID != 1 {
		t.Errorf("got.Results[0].ShowID = %q; want 1", got.Results[0].ShowID)
	}
}

func TestHandleGetShowsFromSetName_Errors(t *testing.T) {
	tests := []struct {
		name       string
		url        string
		wantStatus int
		fake       *fakeQuerier
	}{
		{"missing set name", "/shows?set_name=", http.StatusBadRequest, &fakeQuerier{}},
		{"invalid set name", "/shows?set_name=hello", http.StatusBadRequest, &fakeQuerier{}},
		{"server error", "/shows?set_name=set_1", http.StatusInternalServerError, &fakeQuerier{showsFromSetNameErr: errors.New("db exploded")}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &server{queries: tt.fake}
			req := httptest.NewRequest(http.MethodGet, tt.url, nil)
			w := httptest.NewRecorder()
			s.handleGetShowsFromSetName(w, req)
			if got := w.Result().StatusCode; got != tt.wantStatus {
				t.Errorf("status code = %d; want %d", got, tt.wantStatus)
			}
		})
	}
}

func TestHandleGetShowsFromVenueName(t *testing.T) {
	date1, _ := time.Parse(time.DateOnly, "1995-01-01")
	date2, _ := time.Parse(time.DateOnly, "1995-01-02")
	fake := &fakeQuerier{
		showsFromVenueNameRows: []database.SearchByVenueRow{
			{ShowID: 1, ShowDate: date1, Venue: "Soldier Field", City: "Chicago", Location: "IL", Notes: sql.NullString{}},
			{ShowID: 2, ShowDate: date2, Venue: "Soldier Field", City: "Chicago", Location: "IL", Notes: sql.NullString{}},
		},
	}

	s := &server{queries: fake}
	req := httptest.NewRequest(http.MethodGet, "/shows?venue=soldier", nil)
	w := httptest.NewRecorder()

	s.handleGetShowsFromVenueName(w, req)

	res := w.Result()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("status code = %d; want %d", res.StatusCode, http.StatusOK)
	}

	got := internal.Paginated[internal.ShowMeta]{}
	if err := json.NewDecoder(res.Body).Decode(&got); err != nil {
		t.Fatalf("error decoding response: %v", err)
	}

	if len(got.Results) != 2 {
		t.Errorf("len(got) = %d; want 2", len(got.Results))
	}

	if got.Results[0].Venue != "Soldier Field" {
		t.Errorf("got.Results[0].Venue = %q; want %s", got.Results[0].Venue, "Soldier Field")
	}

	if got.Results[0].Venue != got.Results[1].Venue {
		t.Errorf("different venues: got.Results[0].Venue = %q, got.Results[1].Venue = %q", got.Results[0].Venue, got.Results[1].Venue)
	}
}

func TestHandleGetShowsFromVenueName_Errors(t *testing.T) {
	tests := []struct {
		name       string
		url        string
		wantStatus int
		fake       *fakeQuerier
	}{
		{"missing venue param", "/shows?venue=", http.StatusBadRequest, &fakeQuerier{}},
		{"invalid venue param", "/shows?venue=hello_world", http.StatusNotFound, &fakeQuerier{}},
		{"server error", "/shows?venue=deer_creek", http.StatusInternalServerError, &fakeQuerier{showsFromVenueNameErr: errors.New("db exploded")}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &server{queries: tt.fake}
			req := httptest.NewRequest(http.MethodGet, tt.url, nil)
			w := httptest.NewRecorder()
			s.handleGetShowsFromVenueName(w, req)
			if got := w.Result().StatusCode; got != tt.wantStatus {
				t.Errorf("status code = %d; want %d", got, tt.wantStatus)
			}
		})
	}
}

func TestHandleGetShowsFromLocation(t *testing.T) {
	date, _ := time.Parse(time.DateOnly, "1995-01-01")
	fake := &fakeQuerier{
		showsFromLocationRows: []database.GetShowsFromLocationRow{
			{ShowID: 1, ShowDate: date, Venue: "Soldier Field", City: "Chicago", Location: "IL", Notes: sql.NullString{}},
			{ShowID: 2, ShowDate: date.Add(24 * time.Hour), Venue: "Soldier Field", City: "Chicago", Location: "IL", Notes: sql.NullString{}},
		},
		validLocationRows: []string{"il", "england"},
	}

	s := &server{queries: fake}
	req := httptest.NewRequest(http.MethodGet, "/shows?location=IL", nil)
	w := httptest.NewRecorder()

	s.handleGetShowsFromLocation(w, req)

	res := w.Result()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("status code = %d; want %d", res.StatusCode, http.StatusOK)
	}

	got := internal.Paginated[internal.ShowMeta]{}
	if err := json.NewDecoder(res.Body).Decode(&got); err != nil {
		t.Fatalf("error decoding response: %v", err)
	}

	if len(got.Results) != 2 {
		t.Errorf("len(got) = %d; want 2", len(got.Results))
	}

	if got.Results[0].Location != "IL" {
		t.Errorf("got.Results[0].Location = %q; want 'IL'", got.Results[0].Location)
	}

	if got.Results[0].Location != got.Results[1].Location {
		t.Errorf("different locations: got.Results[0].Location = %q, got.Results[1].Location = %q", got.Results[0].Location, got.Results[1].Location)
	}
}

func TestHandleGetShowsFromLocation_CountryName(t *testing.T) {
	date, _ := time.Parse(time.DateOnly, "1995-01-01")
	fake := &fakeQuerier{
		showsFromLocationRows: []database.GetShowsFromLocationRow{
			{ShowID: 1, ShowDate: date, Venue: "Wembly Empire Pool", City: "London", Location: "England", Notes: sql.NullString{}},
		},
		validLocationRows: []string{"england", "il"},
	}

	s := &server{queries: fake}
	req := httptest.NewRequest(http.MethodGet, "/shows?location=England", nil)
	w := httptest.NewRecorder()

	s.handleGetShowsFromLocation(w, req)

	res := w.Result()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("status code = %d; want %d", res.StatusCode, http.StatusOK)
	}

	got := internal.Paginated[internal.ShowMeta]{}
	if err := json.NewDecoder(res.Body).Decode(&got); err != nil {
		t.Fatalf("error decoding response: %v", err)
	}

	if len(got.Results) != 1 {
		t.Errorf("len(got) = %d; want 1", len(got.Results))
	}

	if got.Results[0].Location != "England" {
		t.Errorf("got.Results[0].Location = %q; want 'England'", got.Results[0].Location)
	}
}

func TestHandleGetShowsFromLocation_Errors(t *testing.T) {
	tests := []struct {
		name       string
		url        string
		wantStatus int
		fake       *fakeQuerier
	}{
		{"empty param", "/shows?location=", http.StatusBadRequest, &fakeQuerier{validLocationRows: []string{"IL", "England"}}},
		{"invalid param", "/shows?location=hello", http.StatusBadRequest, &fakeQuerier{validLocationRows: []string{"IL", "England"}}},
		{"server error", "/shows?location=IL", http.StatusInternalServerError, &fakeQuerier{showsFromLocationErr: errors.New("db exploded")}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &server{queries: tt.fake}
			req := httptest.NewRequest(http.MethodGet, tt.url, nil)
			w := httptest.NewRecorder()
			s.handleGetShowsFromLocation(w, req)
			if got := w.Result().StatusCode; got != tt.wantStatus {
				t.Errorf("status code = %d; want %d", got, tt.wantStatus)
			}
		})
	}
}

func TestHandleGetShowsFromCity(t *testing.T) {
	date, _ := time.Parse(time.DateOnly, "1995-01-01")
	fake := &fakeQuerier{
		showsFromCityRows: []database.GetShowsFromCityRow{
			{ShowID: 1, ShowDate: date, Venue: "Soldier Field", City: "Chicago", Location: "IL", Notes: sql.NullString{}},
			{ShowID: 2, ShowDate: date.Add(24 * time.Hour), Venue: "Soldier Field", City: "Chicago", Location: "IL", Notes: sql.NullString{}},
		},
	}

	s := &server{queries: fake}
	req := httptest.NewRequest(http.MethodGet, "/shows?city=chicago", nil)
	w := httptest.NewRecorder()

	s.handleGetShowsFromCity(w, req)

	res := w.Result()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("status code = %d; want %d", res.StatusCode, http.StatusOK)
	}

	got := internal.Paginated[internal.ShowMeta]{}
	if err := json.NewDecoder(res.Body).Decode(&got); err != nil {
		t.Fatalf("error decoding response: %v", err)
	}

	if len(got.Results) != 2 {
		t.Errorf("len(got) = %d; want 2", len(got.Results))
	}

	if got.Results[0].City != "Chicago" {
		t.Errorf("got.Results[0].City = %q; want %q", got.Results[0].City, "Chicago")
	}
	if got.Results[1].City != "Chicago" {
		t.Errorf("got.Results[1].City = %q; want %q", got.Results[1].City, "Chicago")
	}
}

func TestHandleGetShowsFromCity_Errors(t *testing.T) {
	tests := []struct {
		name       string
		url        string
		wantStatus int
		fake       *fakeQuerier
	}{
		{"empty param", "/shows?city=", http.StatusBadRequest, &fakeQuerier{}},
		{"city not found", "/shows?city=abc", http.StatusNotFound, &fakeQuerier{}},
		{"server error", "/shows?city=reno", http.StatusInternalServerError, &fakeQuerier{showsFromCityErr: errors.New("db exploded")}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &server{queries: tt.fake}
			req := httptest.NewRequest(http.MethodGet, tt.url, nil)
			w := httptest.NewRecorder()
			s.handleGetShowsFromCity(w, req)
			if got := w.Result().StatusCode; got != tt.wantStatus {
				t.Errorf("status code = %d; want %d", got, tt.wantStatus)
			}
		})
	}
}

// func TestHandleGetShowsFromYear(t *testing.T) {
// 	date, _ := time.Parse(time.DateOnly, "1995-01-01")
// 	fake := &fakeQuerier{
// 		showsFromYearRows: []database.GetShowsFromYearRow{
// 			{ShowID: 1, ShowDate: date, Venue: "Soldier Field", City: "Chicago", Location: "IL", Notes: sql.NullString{}},
// 			{ShowID: 2, ShowDate: date.Add(time.Hour * 24), Venue: "Soldier Field", City: "Chicago", Location: "IL", Notes: sql.NullString{}},
// 		},
// 	}
//
// 	s := &server{queries: fake}
// 	req := httptest.NewRequest(http.MethodGet, "/shows?year=1995", nil)
// 	w := httptest.NewRecorder()
//
// 	s.handleGetShowsFromYear(w, req)
//
// 	res := w.Result()
// 	if res.StatusCode != http.StatusOK {
// 		t.Fatalf("status code = %d; want %d", res.StatusCode, http.StatusOK)
// 	}
//
// 	got := internal.Paginated[internal.ShowMeta]{}
// 	if err := json.NewDecoder(res.Body).Decode(&got); err != nil {
// 		t.Fatalf("error decoding response: %v", err)
// 	}
//
// 	if len(got.Results) != 2 {
// 		t.Errorf("len(got) = %d; want 2", len(got.Results))
// 	}
//
// 	if got.Results[0].Date[:4] != "1995" {
// 		t.Errorf("got.Results[0].Date year %q; want '1995'", got.Results[0].Date[:4])
// 	}
//
// 	if got.Results[0].Date != "1995-01-01" {
// 		t.Errorf("got.Results[0].Date = %q; want '1995-01-01'", got.Results[0].Date)
// 	}
// }
//
// func TestHandleGetShowsFromYear_Errors(t *testing.T) {
// 	tests := []struct {
// 		name       string
// 		url        string
// 		wantStatus int
// 		fake       *fakeQuerier
// 	}{
// 		{"empty year param", "/shows?year=", http.StatusBadRequest, &fakeQuerier{}},
// 		{"invalid year", "/shows?year=1", http.StatusBadRequest, &fakeQuerier{}},
// 		{"invalid year string", "/shows?year=hello", http.StatusBadRequest, &fakeQuerier{}},
// 		{"server error", "/shows?year=1967", http.StatusInternalServerError, &fakeQuerier{showsFromYearErr: errors.New("db exploded")}},
// 	}
//
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			s := &server{queries: tt.fake}
// 			req := httptest.NewRequest(http.MethodGet, tt.url, nil)
// 			w := httptest.NewRecorder()
// 			s.handleGetShowsFromYear(w, req)
// 			if got := w.Result().StatusCode; got != tt.wantStatus {
// 				t.Errorf("%q: status code = %d; want %d", tt.name, got, tt.wantStatus)
// 			}
// 		})
// 	}
// }

func TestHandleGetShowsFromYearAndLocation(t *testing.T) {
	date, _ := time.Parse(time.DateOnly, "1995-01-01")
	fake := &fakeQuerier{
		showsFromYearAndLocationRows: []database.GetShowsFromYearAndLocationRow{
			{ShowID: 1, ShowDate: date, Venue: "Soldier Field", City: "Chicago", Location: "IL", Notes: sql.NullString{}},
			{ShowID: 2, ShowDate: date.Add(time.Hour * 24), Venue: "Soldier Field", City: "Chicago", Location: "IL", Notes: sql.NullString{}},
		},
		validLocationRows: []string{"england", "il"},
	}

	s := &server{queries: fake}

	req := httptest.NewRequest(http.MethodGet, "/shows?year=1995&location=IL", nil)
	w := httptest.NewRecorder()

	s.handleGetShowsFromYearAndLocation(w, req)

	res := w.Result()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("status code = %d; want %d", res.StatusCode, http.StatusOK)
	}

	got := internal.Paginated[internal.ShowMeta]{}
	if err := json.NewDecoder(res.Body).Decode(&got); err != nil {
		t.Fatalf("error decoding response: %v", err)
	}

	if len(got.Results) != 2 {
		t.Errorf("len(got) = %d; want 2", len(got.Results))
	}

	if got.Results[0].Location != "IL" {
		t.Errorf("got.Results[0].Location = %q; want 'IL'", got.Results[0].Location)
	}

	if got.Results[0].Date[:4] != "1995" {
		t.Errorf("got.Results[0].Date year = %q; want '1995'", got.Results[0].Date[:4])
	}
}

func TestHandleGetShowsFromYearAndLocation_CountryName(t *testing.T) {
	date, _ := time.Parse(time.DateOnly, "1995-01-01")
	fake := &fakeQuerier{
		showsFromYearAndLocationRows: []database.GetShowsFromYearAndLocationRow{
			{ShowID: 1, ShowDate: date, Venue: "Wembly Stadium", City: "London", Location: "England", Notes: sql.NullString{}},
		},
		validLocationRows: []string{"england", "il"},
	}

	s := &server{queries: fake}

	req := httptest.NewRequest(http.MethodGet, "/shows?year=1995&location=England", nil)
	w := httptest.NewRecorder()

	s.handleGetShowsFromYearAndLocation(w, req)

	res := w.Result()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("status code = %d; want %d", res.StatusCode, http.StatusOK)
	}

	got := internal.Paginated[internal.ShowMeta]{}
	if err := json.NewDecoder(res.Body).Decode(&got); err != nil {
		t.Fatalf("error decoding response: %v", err)
	}

	if len(got.Results) != 1 {
		t.Errorf("len(got) = %d; want 1", len(got.Results))
	}

	if got.Results[0].Location != "England" {
		t.Errorf("got.Results[0].Location = %q; want 'England'", got.Results[0].Location)
	}

	if got.Results[0].Date[:4] != "1995" {
		t.Errorf("got.Results[0].Date year = %q; want '1995'", got.Results[0].Date[:4])
	}
}

func TestHandleGetShowsFromYearAndLocation_Errors(t *testing.T) {
	tests := []struct {
		name       string
		url        string
		wantStatus int
		fake       *fakeQuerier
	}{
		{"empty year param", "/shows?year=&location=IL", http.StatusBadRequest, &fakeQuerier{}},
		{"empty location param", "/shows?year=1995&location=", http.StatusBadRequest, &fakeQuerier{}},
		{"server error", "/shows?year=1965&location=CA", http.StatusInternalServerError, &fakeQuerier{showsFromYearAndLocationErr: errors.New("db exploded")}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &server{queries: tt.fake}
			req := httptest.NewRequest(http.MethodGet, tt.url, nil)
			w := httptest.NewRecorder()
			s.handleGetShowsFromYearAndLocation(w, req)
			if got := w.Result().StatusCode; got != tt.wantStatus {
				t.Errorf("%q: status code = %d; want %d", tt.name, got, tt.wantStatus)
			}
		})
	}
}

func TestHandleGetShowsFromNotes_WithNotes(t *testing.T) {
	date, _ := time.Parse(time.DateOnly, "1995-07-09")
	fake := &fakeQuerier{
		showsWithNotesRows: []database.ShowsWithShowNotesRow{
			{ShowID: 1, ShowDate: date, Venue: "Soldier Field", City: "Chicago", Location: "IL", Notes: sql.NullString{String: "Final show", Valid: true}},
		},
	}

	s := &server{queries: fake}
	req := httptest.NewRequest(http.MethodGet, "/shows?has_notes=true", nil)
	w := httptest.NewRecorder()

	s.handleGetShowsFromNotes(w, req)

	res := w.Result()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("status code = %d; want %d", res.StatusCode, http.StatusOK)
	}

	got := internal.Paginated[internal.ShowMeta]{}
	if err := json.NewDecoder(res.Body).Decode(&got); err != nil {
		t.Fatalf("error decoding response: %v", err)
	}

	if len(got.Results) != 1 {
		t.Errorf("len(got) = %d; want 1", len(got.Results))
	}

	if got.Results[0].Notes != "Final show" {
		t.Errorf("got.Results[0].Notes = %q; want 'Final show'", got.Results[0].Notes)
	}
}

func TestHandleGetShowsFromNotes_WithoutNotes(t *testing.T) {
	date, _ := time.Parse(time.DateOnly, "1995-07-09")
	fake := &fakeQuerier{
		showsWithoutNotesRows: []database.ShowsWithoutNotesRow{
			{ShowID: 1, ShowDate: date, Venue: "Soldier Field", City: "Chicago", Location: "IL"},
			{ShowID: 2, ShowDate: date.Add(time.Hour * 24), Venue: "Wembly", City: "London", Location: "England"},
		},
	}

	s := &server{queries: fake}
	req := httptest.NewRequest(http.MethodGet, "/shows?has_notes=false", nil)
	w := httptest.NewRecorder()

	s.handleGetShowsFromNotes(w, req)

	res := w.Result()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("status code = %d; want %d", res.StatusCode, http.StatusOK)
	}

	got := internal.Paginated[internal.ShowMeta]{}
	if err := json.NewDecoder(res.Body).Decode(&got); err != nil {
		t.Fatalf("error decoding response: %v", err)
	}

	if len(got.Results) != 2 {
		t.Errorf("len(got) = %d; want 2", len(got.Results))
	}

	if got.Results[0].Notes != "" {
		t.Errorf("got.Results[0].Notes = %q; want ''", got.Results[0].Notes)
	}
}

func TestHandleGetShowsFromNotes_Errors(t *testing.T) {
	tests := []struct {
		name       string
		url        string
		wantStatus int
		fake       *fakeQuerier
	}{
		{"missing notes param", "/shows?has_notes=", http.StatusBadRequest, &fakeQuerier{}},
		{"invalid notes param", "/shows?has_notes=yes", http.StatusBadRequest, &fakeQuerier{}},
		{"server error", "/shows?has_notes=false", http.StatusInternalServerError, &fakeQuerier{showsWithoutNotesErr: errors.New("without notes: db exploded")}},
		{"server error", "/shows?has_notes=true", http.StatusInternalServerError, &fakeQuerier{showsWithNotesErr: errors.New("with notes: db exploded")}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &server{queries: tt.fake}
			req := httptest.NewRequest(http.MethodGet, tt.url, nil)
			w := httptest.NewRecorder()
			s.handleGetShowsFromNotes(w, req)
			if got := w.Result().StatusCode; got != tt.wantStatus {
				t.Errorf("%q: status code = %d; want %d", tt.name, got, tt.wantStatus)
			}
		})
	}
}

func TestHandleSearchShowNotes(t *testing.T) {
	date, _ := time.Parse(time.DateOnly, "1995-09-04")
	wantedNote := "This is a show note..."
	fake := &fakeQuerier{
		searchShowNotesRows: []database.SearchShowNotesRow{
			{ShowID: 1, ShowDate: date, Venue: "Soldier Field", City: "Chicago", Location: "IL", Notes: sql.NullString{String: wantedNote, Valid: true}},
		},
	}
	s := &server{queries: fake}
	req := httptest.NewRequest(http.MethodGet, "/shows?notes=this", nil)
	w := httptest.NewRecorder()

	s.handleSearchShowNotes(w, req)

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

	if got.Results[0].Notes != wantedNote {
		t.Errorf("got.Results[0].Notes = %q; want %q", got.Results[0].Notes, wantedNote)
	}
}

func TestHandleSearchShowNotes_Errors(t *testing.T) {
	tests := []struct {
		name       string
		url        string
		wantStatus int
		fake       *fakeQuerier
	}{
		{"missing notes param", "/shows?notes=", http.StatusBadRequest, &fakeQuerier{}},
		{"no show results", "/shows?notes=pikachu", http.StatusNotFound, &fakeQuerier{}},
		{"server error", "/shows?notes=benefit", http.StatusInternalServerError, &fakeQuerier{searchShowNotesErr: errors.New("search show notes: db exploded")}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &server{queries: tt.fake}
			req := httptest.NewRequest(http.MethodGet, tt.url, nil)
			w := httptest.NewRecorder()
			s.handleSearchShowNotes(w, req)
			if got := w.Result().StatusCode; got != tt.wantStatus {
				t.Errorf("%q: status code = %d; want %d", tt.name, got, tt.wantStatus)
			}
		})
	}
}
