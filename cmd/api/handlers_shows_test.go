package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/LunarDrift/deadabase/internal"
	"github.com/LunarDrift/deadabase/internal/database"
)

// fakeQuerier is a fake 'database' with all the methods required to satisfy ShowQuerier. Used so
// tests don't require a connection to the real database
type fakeQuerier struct {
	allShowIDs             []int32
	allShowIDsErr          error
	showFromIDRows         []database.GetShowFromIDRow
	showFromIDErr          error
	footnoteRows           []database.GetFootnotesFromShowIDRow
	footnoteErr            error
	showFromDateRows       []database.GetShowFromDateRow
	showFromDateErr        error
	showsBetweenDatesRows  []database.GetShowsBetweenDatesRow
	showsBetweenDatesErr   error
	showsFromSongNameRows  []database.GetShowsFromSongNameRow
	showsFromSongNameErr   error
	showsFromSetNameRows   []database.GetShowsFromSetNameRow
	showsFromSetNameErr    error
	showsFromVenueNameRows []database.SearchByVenueRow
	showsFromVenueNameErr  error
	showsFromStateRows     []database.GetShowsFromStateRow
	showsFromStateErr      error
	showsFromYearRows      []database.GetShowsFromYearRow
	showsFromYearErr       error
}

func (f *fakeQuerier) GetAllShowIDs(ctx context.Context) ([]int32, error) {
	return f.allShowIDs, f.allShowIDsErr
}

func (f *fakeQuerier) GetShowFromID(ctx context.Context, showID int32) ([]database.GetShowFromIDRow, error) {
	return f.showFromIDRows, f.showFromIDErr
}

func (f *fakeQuerier) GetShowFromDate(ctx context.Context, showDate time.Time) ([]database.GetShowFromDateRow, error) {
	return f.showFromDateRows, f.showFromDateErr
}

func (f *fakeQuerier) GetShowsBetweenDates(ctx context.Context, arg database.GetShowsBetweenDatesParams) ([]database.GetShowsBetweenDatesRow, error) {
	return f.showsBetweenDatesRows, f.showsBetweenDatesErr
}

func (f *fakeQuerier) GetShowsFromSetName(ctx context.Context, setName string) ([]database.GetShowsFromSetNameRow, error) {
	return f.showsFromSetNameRows, f.showsFromSetNameErr
}

func (f *fakeQuerier) GetShowsFromSongName(ctx context.Context, rawEntry string) ([]database.GetShowsFromSongNameRow, error) {
	return f.showsFromSongNameRows, f.showsFromSongNameErr
}

func (f *fakeQuerier) GetShowsFromState(ctx context.Context, stateOrCountry string) ([]database.GetShowsFromStateRow, error) {
	return f.showsFromStateRows, f.showsFromStateErr
}

func (f *fakeQuerier) GetShowsFromYear(ctx context.Context, year int32) ([]database.GetShowsFromYearRow, error) {
	return f.showsFromYearRows, f.showsFromYearErr
}

func (f *fakeQuerier) GetShowsFromYearAndState(ctx context.Context, arg database.GetShowsFromYearAndStateParams) ([]database.GetShowsFromYearAndStateRow, error) {
	return nil, nil
}

func (f *fakeQuerier) SearchByVenue(ctx context.Context, venue string) ([]database.SearchByVenueRow, error) {
	return f.showsFromVenueNameRows, f.showsFromVenueNameErr
}

func (f *fakeQuerier) ShowsWithShowNotes(ctx context.Context) ([]database.ShowsWithShowNotesRow, error) {
	return nil, nil
}

func (f *fakeQuerier) ShowsWithoutNotes(ctx context.Context) ([]database.ShowsWithoutNotesRow, error) {
	return nil, nil
}

func (f *fakeQuerier) SongStats(ctx context.Context, songName sql.NullString) (database.SongStatsRow, error) {
	return database.SongStatsRow{}, nil
}

func (f *fakeQuerier) AllSongsPlayedAtVenue(ctx context.Context, venue string) ([]database.AllSongsPlayedAtVenueRow, error) {
	return nil, nil
}

func (f *fakeQuerier) MostCommonSongsBySetName(ctx context.Context, setName string) ([]database.MostCommonSongsBySetNameRow, error) {
	return nil, nil
}

func (f *fakeQuerier) MostPlayedSongs(ctx context.Context) ([]database.MostPlayedSongsRow, error) {
	return nil, nil
}

func (f *fakeQuerier) SongsPlayedLessThan(ctx context.Context, number int32) ([]database.SongsPlayedLessThanRow, error) {
	return nil, nil
}

func (f *fakeQuerier) UniqueSongsPerCity(ctx context.Context) ([]database.UniqueSongsPerCityRow, error) {
	return nil, nil
}

func (f *fakeQuerier) GetFootnotesFromShowID(ctx context.Context, showID int32) ([]database.GetFootnotesFromShowIDRow, error) {
	return f.footnoteRows, f.footnoteErr
}

// ----------------------------------------------------------------------------------------------------------------------------------
// -------------------------------------------------------------- TESTS -------------------------------------------------------------
// ----------------------------------------------------------------------------------------------------------------------------------

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

func TestHandleGetShowsFromSongName_EmptySongName(t *testing.T) {
	fake := &fakeQuerier{}
	s := &server{queries: fake}
	req := httptest.NewRequest(http.MethodGet, "/shows?song=", nil)
	w := httptest.NewRecorder()

	s.handleGetShowsFromSongName(w, req)

	res := w.Result()
	if res.StatusCode != http.StatusBadRequest {
		t.Errorf("status code = %d; want %d", res.StatusCode, http.StatusBadRequest)
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

func TestHandleGetShowsFromSetName_InvalidSetName(t *testing.T) {
	fake := &fakeQuerier{}
	s := &server{queries: fake}

	req := httptest.NewRequest(http.MethodGet, "/shows?set_name=hello", nil)
	w := httptest.NewRecorder()

	s.handleGetShowsFromSetName(w, req)

	res := w.Result()
	if res.StatusCode != http.StatusBadRequest {
		t.Errorf("status code = %d; want %d", res.StatusCode, http.StatusBadRequest)
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

func TestHandleGetShowsFromVenueName_EmptyVenueParam(t *testing.T) {
	fake := &fakeQuerier{}
	s := &server{queries: fake}

	req := httptest.NewRequest(http.MethodGet, "/shows?venue=", nil)
	w := httptest.NewRecorder()

	s.handleGetShowsFromVenueName(w, req)

	res := w.Result()
	if res.StatusCode != http.StatusBadRequest {
		t.Errorf("status code = %d; want %d", res.StatusCode, http.StatusBadRequest)
	}
}

func TestHandleGetShowsFromState(t *testing.T) {
	date, _ := time.Parse(time.DateOnly, "1995-01-01")
	fake := &fakeQuerier{
		showsFromStateRows: []database.GetShowsFromStateRow{
			{ShowID: 1, ShowDate: date, Venue: "Soldier Field", City: "Chicago", State: "IL", Notes: sql.NullString{}},
			{ShowID: 2, ShowDate: date.Add(24 * time.Hour), Venue: "Soldier Field", City: "Chicago", State: "IL", Notes: sql.NullString{}},
		},
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

func TestHandleGetShowsFromState_EmptyStateParam(t *testing.T) {
	fake := &fakeQuerier{}
	s := &server{queries: fake}

	req := httptest.NewRequest(http.MethodGet, "/shows?state=", nil)
	w := httptest.NewRecorder()

	s.handleGetShowsFromState(w, req)

	res := w.Result()
	if res.StatusCode != http.StatusBadRequest {
		t.Errorf("status code = %d; want %d", res.StatusCode, http.StatusBadRequest)
	}
}

func TestHandleGetShowsFromState_CountryName(t *testing.T) {
	date, _ := time.Parse(time.DateOnly, "1995-01-01")
	fake := &fakeQuerier{
		showsFromStateRows: []database.GetShowsFromStateRow{
			{ShowID: 1, ShowDate: date, Venue: "Wembly Empire Pool", City: "London", State: "England", Notes: sql.NullString{}},
		},
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

func TestHandleGetShowsFromYear_MissingYearParam(t *testing.T) {
	fake := &fakeQuerier{}
	s := &server{queries: fake}

	req := httptest.NewRequest(http.MethodGet, "/shows?year=", nil)
	w := httptest.NewRecorder()

	s.handleGetShowsFromYear(w, req)

	res := w.Result()
	if res.StatusCode != http.StatusBadRequest {
		t.Errorf("status code = %d; want %d", res.StatusCode, http.StatusBadRequest)
	}
}

func TestHandleGetShowsFromYear_InvalidYearNumber(t *testing.T) {
	fake := &fakeQuerier{}
	s := &server{queries: fake}

	req := httptest.NewRequest(http.MethodGet, "/shows?year=1", nil)
	w := httptest.NewRecorder()

	s.handleGetShowsFromYear(w, req)

	res := w.Result()
	if res.StatusCode != http.StatusBadRequest {
		t.Errorf("status code = %d; want %d", res.StatusCode, http.StatusBadRequest)
	}
}

func TestHandleGetShowsFromYear_InvalidYearString(t *testing.T) {
	fake := &fakeQuerier{}
	s := server{queries: fake}

	req := httptest.NewRequest(http.MethodGet, "/shows?year=hello", nil)
	w := httptest.NewRecorder()

	s.handleGetShowsFromYear(w, req)

	res := w.Result()
	if res.StatusCode != http.StatusBadRequest {
		t.Errorf("status code = %d; want %d", res.StatusCode, http.StatusBadRequest)
	}
}
