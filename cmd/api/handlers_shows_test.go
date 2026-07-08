package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/LunarDrift/deadabase/internal"
	"github.com/LunarDrift/deadabase/internal/database"
)

func TestHandleShowsFromPathVal_ByID(t *testing.T) {
	date, err := time.Parse(time.DateOnly, "1965-01-01")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	fake := &fakeQuerier{
		allShowIDs: []int32{42},
		showFromIDRows: []database.GetShowFromIDRow{
			{
				ShowID:   42,
				ShowDate: date,
				Venue:    "Fillmore West",
				City:     "San Francisco",
				State:    "CA",
				Notes:    sql.NullString{},
				SetName:  sql.NullString{String: "set_1", Valid: true},
				RawEntry: sql.NullString{String: "Dark Star", Valid: true},
			},
		},
		footnoteRows: []database.GetFootnotesFromShowIDRow{},
	}

	s := &server{queries: fake}

	req := httptest.NewRequest(http.MethodGet, "/shows/42", nil)
	req.SetPathValue("value", "42")
	w := httptest.NewRecorder()

	s.handleShowsFromPathVal(w, req)

	res := w.Result()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("status = %d; want 200", res.StatusCode)
	}

	var got internal.ShowResponse
	if err := json.NewDecoder(res.Body).Decode(&got); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if got.Venue != "Fillmore West" {
		t.Errorf("got.Venue = %q; want %q", got.Venue, "Fillmore West")
	}
}

func TestHandleShowsFromPathVal_ByID_EmptySetlist(t *testing.T) {
	date, err := time.Parse(time.DateOnly, "1965-01-01")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	fake := &fakeQuerier{
		allShowIDs: []int32{42},
		showFromIDRows: []database.GetShowFromIDRow{
			{
				ShowID:   42,
				ShowDate: date,
				Venue:    "Fillmore West",
				City:     "San Francisco",
				State:    "CA",
				Notes:    sql.NullString{},
				SetName:  sql.NullString{},
				RawEntry: sql.NullString{},
			},
		},
		footnoteRows: []database.GetFootnotesFromShowIDRow{},
	}

	s := &server{queries: fake}

	req := httptest.NewRequest(http.MethodGet, "/shows/42", nil)
	req.SetPathValue("value", "42")
	w := httptest.NewRecorder()

	s.handleShowsFromPathVal(w, req)

	res := w.Result()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("status = %d; want 200", res.StatusCode)
	}

	var got internal.ShowWithNoSetlist
	if err := json.NewDecoder(res.Body).Decode(&got); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	expectedMessage := "No setlist available for this show"
	if got.Message != expectedMessage {
		t.Errorf("got.Message = %q; want %q", got.Message, expectedMessage)
	}
}

func TestHandleShowsFromPathVal_ByDate(t *testing.T) {
	date, err := time.Parse(time.DateOnly, "1969-09-30")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	fake := &fakeQuerier{
		showFromDateRows: []database.GetShowFromDateRow{
			{
				ShowID:   int32(1949),
				ShowDate: date,
				Venue:    "Cafe Au Go Go",
				City:     "New York",
				State:    "NY",
				Notes:    sql.NullString{},
				SetName:  sql.NullString{String: "set_1", Valid: true},
				RawEntry: sql.NullString{String: "China Cat Sunflower > I Know You Rider", Valid: true},
			},
		},
		footnoteRows: []database.GetFootnotesFromShowIDRow{},
	}

	s := &server{queries: fake}

	req := httptest.NewRequest(http.MethodGet, "/shows/1969-09-30", nil)
	req.SetPathValue("value", "1969-09-30")
	w := httptest.NewRecorder()

	s.handleShowsFromPathVal(w, req)

	res := w.Result()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("status = %d; want 200", res.StatusCode)
	}

	var got []internal.ShowResponse
	if err := json.NewDecoder(res.Body).Decode(&got); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(got) != 1 {
		t.Errorf("len(got) = %v; want 1", len(got))
	}

	if got[0].Date != "1969-09-30" {
		t.Errorf("got.Date = %v; want 1969-09-30", got[0].Date)
	}
}

func TestHandleShowsFromPathVal_ByDate_EarlyLateShows(t *testing.T) {
	date, _ := time.Parse(time.DateOnly, "1969-09-30")

	fake := &fakeQuerier{
		showFromDateRows: []database.GetShowFromDateRow{
			{
				ShowID: 1949, ShowDate: date, Venue: "Cafe Au Go Go", City: "New York", State: "NY",
				SetName: sql.NullString{String: "set_1", Valid: true}, RawEntry: sql.NullString{String: "Early show song", Valid: true},
			},
			{
				ShowID: 1950, ShowDate: date, Venue: "Cafe Au Go Go", City: "New York", State: "NY",
				SetName: sql.NullString{String: "set_1", Valid: true}, RawEntry: sql.NullString{String: "Late show song", Valid: true},
			},
		},
		footnoteRows: []database.GetFootnotesFromShowIDRow{},
	}

	s := &server{queries: fake}
	req := httptest.NewRequest(http.MethodGet, "/shows/1969-09-30", nil)
	req.SetPathValue("value", "1969-09-30")
	w := httptest.NewRecorder()

	s.handleShowsFromPathVal(w, req)

	var got []internal.ShowResponse
	if err := json.NewDecoder(w.Result().Body).Decode(&got); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(got) != 2 {
		t.Errorf("len(got) = %d; want 2 (early + late show)", len(got))
	}

	if got[0].Date != got[1].Date {
		t.Errorf("got[0].Date = %q; want %q", got[0].Date, got[1].Date)
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
			{ShowID: 1, ShowDate: date, Venue: "test venue", City: "test city", State: "test state", Notes: sql.NullString{}},
			{ShowID: 2, ShowDate: date.Add(time.Hour * 48), Venue: "test venue 2", City: "test city 2", State: "test state 2", Notes: sql.NullString{}},
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

	var got []internal.ShowMeta
	if err := json.NewDecoder(res.Body).Decode(&got); err != nil {
		t.Fatalf("error decoding response: %v", err)
	}

	if len(got) != len(fake.showsFromSongNameRows) {
		t.Errorf("len(got) = %d; want %d", len(got), len(fake.showsFromSongNameRows))
	}

	if got[0].Date != "1995-01-01" {
		t.Errorf("got[0].Date = %q; want %q", got[0].Date, "1995-01-01")
	}

	if got[0].Venue != "test venue" {
		t.Errorf("got[0].Venue = %q; want 'test venue'", got[0].Venue)
	}
	if got[1].Venue != "test venue 2" {
		t.Errorf("got[1].Venue = %q; want 'test venue 2'", got[1].Venue)
	}
}

func TestHandleGetShowsFromSongName_Errors(t *testing.T) {
	tests := []struct {
		name       string
		url        string
		wantStatus int
	}{
		{"missing song name", "/shows?song=", http.StatusBadRequest},
		{"invalid song name", "/shows?song=123", http.StatusNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &server{queries: &fakeQuerier{}}
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
			{ShowID: 1, ShowDate: date, Venue: "Soldier Field", City: "Chicago", State: "test state", Notes: sql.NullString{}},
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

	var got []internal.ShowMeta
	if err := json.NewDecoder(res.Body).Decode(&got); err != nil {
		t.Fatalf("error decoding response: %v", err)
	}

	if got[0].ShowID != 1 {
		t.Errorf("got[0].ShowID = %q; want 1", got[0].ShowID)
	}
}

func TestHandleGetShowsFromSetName_Errors(t *testing.T) {
	tests := []struct {
		name       string
		url        string
		wantStatus int
	}{
		{"missing set name", "/shows?set_name=", http.StatusBadRequest},
		{"invalid set name", "/shows?set_name=hello", http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &server{queries: &fakeQuerier{}}
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
			{ShowID: 1, ShowDate: date1, Venue: "Soldier Field", City: "Chicago", State: "IL", Notes: sql.NullString{}},
			{ShowID: 2, ShowDate: date2, Venue: "Soldier Field", City: "Chicago", State: "IL", Notes: sql.NullString{}},
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

	var got []internal.ShowMeta
	if err := json.NewDecoder(res.Body).Decode(&got); err != nil {
		t.Fatalf("error decoding response: %v", err)
	}

	if len(got) != 2 {
		t.Errorf("len(got) = %d; want 2", len(got))
	}

	if got[0].Venue != "Soldier Field" {
		t.Errorf("got[0].Venue = %q; want %s", got[0].Venue, "Soldier Field")
	}

	if got[0].Venue != got[1].Venue {
		t.Errorf("different venues: got[0].Venue = %q, got[1].Venue = %q", got[0].Venue, got[1].Venue)
	}
}

func TestHandleGetShowsFromVenueName_Errors(t *testing.T) {
	tests := []struct {
		name       string
		url        string
		wantStatus int
	}{
		{"missing venue param", "/shows?venue=", http.StatusBadRequest},
		{"invalid venue param", "/shows?venue=hello_world", http.StatusNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &server{queries: &fakeQuerier{}}
			req := httptest.NewRequest(http.MethodGet, tt.url, nil)
			w := httptest.NewRecorder()
			s.handleGetShowsFromVenueName(w, req)
			if got := w.Result().StatusCode; got != tt.wantStatus {
				t.Errorf("status code = %d; want %d", got, tt.wantStatus)
			}
		})
	}
}

func TestHandleGetShowsFromState(t *testing.T) {
	date, _ := time.Parse(time.DateOnly, "1995-01-01")
	fake := &fakeQuerier{
		showsFromStateRows: []database.GetShowsFromStateRow{
			{ShowID: 1, ShowDate: date, Venue: "Soldier Field", City: "Chicago", State: "IL", Notes: sql.NullString{}},
			{ShowID: 2, ShowDate: date.Add(24 * time.Hour), Venue: "Soldier Field", City: "Chicago", State: "IL", Notes: sql.NullString{}},
		},
		validLocationRows: []string{"IL", "England"},
	}

	s := &server{queries: fake}
	req := httptest.NewRequest(http.MethodGet, "/shows?state=IL", nil)
	w := httptest.NewRecorder()

	s.handleGetShowsFromState(w, req)

	res := w.Result()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("status code = %d; want %d", res.StatusCode, http.StatusOK)
	}

	var got []internal.ShowMeta
	err := json.NewDecoder(res.Body).Decode(&got)
	if err != nil {
		t.Fatalf("error decoding response: %v", err)
	}

	if len(got) != 2 {
		t.Errorf("len(got) = %d; want 2", len(got))
	}

	if got[0].State != "IL" {
		t.Errorf("got[0].State = %q; want 'IL'", got[0].State)
	}

	if got[0].State != got[1].State {
		t.Errorf("different states: got[0].State = %q, got[1].State = %q", got[0].State, got[1].State)
	}
}

func TestHandleGetShowsFromState_CountryName(t *testing.T) {
	date, _ := time.Parse(time.DateOnly, "1995-01-01")
	fake := &fakeQuerier{
		showsFromStateRows: []database.GetShowsFromStateRow{
			{ShowID: 1, ShowDate: date, Venue: "Wembly Empire Pool", City: "London", State: "England", Notes: sql.NullString{}},
		},
		validLocationRows: []string{"England", "IL"},
	}

	s := &server{queries: fake}
	req := httptest.NewRequest(http.MethodGet, "/shows?state=England", nil)
	w := httptest.NewRecorder()

	s.handleGetShowsFromState(w, req)

	res := w.Result()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("status code = %d; want %d", res.StatusCode, http.StatusOK)
	}

	var got []internal.ShowMeta
	err := json.NewDecoder(res.Body).Decode(&got)
	if err != nil {
		t.Fatalf("error decoding response: %v", err)
	}

	if len(got) != 1 {
		t.Errorf("len(got) = %d; want 1", len(got))
	}

	if got[0].State != "England" {
		t.Errorf("got[0].State = %q; want 'England'", got[0].State)
	}
}

func TestHandleGetShowsFromState_Errors(t *testing.T) {
	tests := []struct {
		name       string
		url        string
		wantStatus int
	}{
		{"empty param", "/shows?state=", http.StatusBadRequest},
		{"invalid param", "/shows?state=hello", http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &server{queries: &fakeQuerier{
				validLocationRows: []string{"IL", "England"},
			}}
			req := httptest.NewRequest(http.MethodGet, tt.url, nil)
			w := httptest.NewRecorder()
			s.handleGetShowsFromState(w, req)
			if got := w.Result().StatusCode; got != tt.wantStatus {
				t.Errorf("status code = %d; want %d", got, tt.wantStatus)
			}
		})
	}
}

func TestHandleGetShowsFromYear(t *testing.T) {
	date, _ := time.Parse(time.DateOnly, "1995-01-01")
	fake := &fakeQuerier{
		showsFromYearRows: []database.GetShowsFromYearRow{
			{ShowID: 1, ShowDate: date, Venue: "Soldier Field", City: "Chicago", State: "IL", Notes: sql.NullString{}},
			{ShowID: 2, ShowDate: date.Add(time.Hour * 24), Venue: "Soldier Field", City: "Chicago", State: "IL", Notes: sql.NullString{}},
		},
	}

	s := &server{queries: fake}
	req := httptest.NewRequest(http.MethodGet, "/shows?year=1995", nil)
	w := httptest.NewRecorder()

	s.handleGetShowsFromYear(w, req)

	res := w.Result()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("status code = %d; want %d", res.StatusCode, http.StatusOK)
	}

	var got []internal.ShowMeta
	err := json.NewDecoder(res.Body).Decode(&got)
	if err != nil {
		t.Fatalf("error decoding response: %v", err)
	}

	if len(got) != 2 {
		t.Errorf("len(got) = %d; want 2", len(got))
	}

	if got[0].Date[:4] != "1995" {
		t.Errorf("got[0].Date year %q; want '1995'", got[0].Date[:4])
	}

	if got[0].Date != "1995-01-01" {
		t.Errorf("got[0].Date = %q; want '1995-01-01'", got[0].Date)
	}
}

func TestHandleGetShowsFromYear_Errors(t *testing.T) {
	tests := []struct {
		name       string
		url        string
		wantStatus int
	}{
		{"empty param", "/shows?year=", http.StatusBadRequest},
		{"invalid year", "/shows?year=1", http.StatusBadRequest},
		{"invalid year string", "/shows?year=hello", http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &server{queries: &fakeQuerier{}}
			req := httptest.NewRequest(http.MethodGet, tt.url, nil)
			w := httptest.NewRecorder()
			s.handleGetShowsFromYear(w, req)
			if got := w.Result().StatusCode; got != tt.wantStatus {
				t.Errorf("%q: status code = %d; want %d", tt.name, got, tt.wantStatus)
			}
		})
	}
}

func TestHandleGetShowsFromYearAndState(t *testing.T) {
	date, _ := time.Parse(time.DateOnly, "1995-01-01")
	fake := &fakeQuerier{
		showsFromYearAndStateRows: []database.GetShowsFromYearAndStateRow{
			{ShowID: 1, ShowDate: date, Venue: "Soldier Field", City: "Chicago", State: "IL", Notes: sql.NullString{}},
			{ShowID: 2, ShowDate: date.Add(time.Hour * 24), Venue: "Soldier Field", City: "Chicago", State: "IL", Notes: sql.NullString{}},
		},
	}

	s := &server{queries: fake}

	req := httptest.NewRequest(http.MethodGet, "/shows?year=1995&state=IL", nil)
	w := httptest.NewRecorder()

	s.handleGetShowsFromYearAndState(w, req)

	res := w.Result()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("status code = %d; want %d", res.StatusCode, http.StatusOK)
	}

	var got []internal.ShowMeta
	err := json.NewDecoder(res.Body).Decode(&got)
	if err != nil {
		t.Fatalf("error decoding response: %v", err)
	}

	if len(got) != 2 {
		t.Errorf("len(got) = %d; want 2", len(got))
	}

	if got[0].State != "IL" {
		t.Errorf("got[0].State = %q; want 'IL'", got[0].State)
	}

	if got[0].Date[:4] != "1995" {
		t.Errorf("got[0].Date year = %q; want '1995'", got[0].Date[:4])
	}
}

func TestHandleGetShowsFromYearAndState_CountryName(t *testing.T) {
	date, _ := time.Parse(time.DateOnly, "1995-01-01")
	fake := &fakeQuerier{
		showsFromYearAndStateRows: []database.GetShowsFromYearAndStateRow{
			{ShowID: 1, ShowDate: date, Venue: "Wembly Stadium", City: "London", State: "England", Notes: sql.NullString{}},
		},
	}

	s := &server{queries: fake}

	req := httptest.NewRequest(http.MethodGet, "/shows?year=1995&state=london", nil)
	w := httptest.NewRecorder()

	s.handleGetShowsFromYearAndState(w, req)

	res := w.Result()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("status code = %d; want %d", res.StatusCode, http.StatusOK)
	}

	var got []internal.ShowMeta
	err := json.NewDecoder(res.Body).Decode(&got)
	if err != nil {
		t.Fatalf("error decoding response: %v", err)
	}

	if len(got) != 1 {
		t.Errorf("len(got) = %d; want 1", len(got))
	}

	if got[0].State != "England" {
		t.Errorf("got[0].State = %q; want 'England'", got[0].State)
	}

	if got[0].Date[:4] != "1995" {
		t.Errorf("got[0].Date year = %q; want '1995'", got[0].Date[:4])
	}
}

func TestHandleGetShowsFromYearAndState_Errors(t *testing.T) {
	tests := []struct {
		name       string
		url        string
		wantStatus int
	}{
		{"empty year param", "/shows?year=&state=IL", http.StatusBadRequest},
		{"empty state param", "/shows?year=1995&state=", http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &server{queries: &fakeQuerier{}}
			req := httptest.NewRequest(http.MethodGet, tt.url, nil)
			w := httptest.NewRecorder()
			s.handleGetShowsFromYearAndState(w, req)
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
			{ShowID: 1, ShowDate: date, Venue: "Soldier Field", City: "Chicago", State: "IL", Notes: sql.NullString{String: "Final show", Valid: true}},
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

	var got []internal.ShowMeta
	err := json.NewDecoder(res.Body).Decode(&got)
	if err != nil {
		t.Fatalf("error decoding response: %v", err)
	}

	if len(got) != 1 {
		t.Errorf("len(got) = %d; want 1", len(got))
	}

	if got[0].Notes != "Final show" {
		t.Errorf("got[0].Notes = %q; want 'Final show", got[0].Notes)
	}
}

func TestHandleGetShowsFromNotes_WithoutNotes(t *testing.T) {
	date, _ := time.Parse(time.DateOnly, "1995-07-09")
	fake := &fakeQuerier{
		showsWithoutNotesRows: []database.ShowsWithoutNotesRow{
			{ShowID: 1, ShowDate: date, Venue: "Soldier Field", City: "Chicago", State: "IL"},
			{ShowID: 2, ShowDate: date.Add(time.Hour * 24), Venue: "Wembly", City: "London", State: "England"},
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

	var got []internal.ShowMeta
	err := json.NewDecoder(res.Body).Decode(&got)
	if err != nil {
		t.Fatalf("error decoding response: %v", err)
	}

	if len(got) != 2 {
		t.Errorf("len(got) = %d; want 2", len(got))
	}

	if got[0].Notes != "" {
		t.Errorf("got[0].Notes = %q; want ''", got[0].Notes)
	}
}

func TestHandleGetShowsFromNotes_Errors(t *testing.T) {
	tests := []struct {
		name       string
		url        string
		wantStatus int
	}{
		{"missing param", "/shows?has_notes=", http.StatusBadRequest},
		{"invalid param", "/shows?has_notes=yes", http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &server{queries: &fakeQuerier{}}
			req := httptest.NewRequest(http.MethodGet, tt.url, nil)
			w := httptest.NewRecorder()
			s.handleGetShowsFromNotes(w, req)
			if got := w.Result().StatusCode; got != tt.wantStatus {
				t.Errorf("%q: status code = %d; want %d", tt.name, got, tt.wantStatus)
			}
		})
	}
}
