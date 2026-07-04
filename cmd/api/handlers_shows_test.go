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

type fakeQuerier struct {
	allShowIDs     []int32
	allShowIDsErr  error
	showFromIDRows []database.GetShowFromIDRow
	showFromIDErr  error
	footnoteRows   []database.GetFootnotesFromShowIDRow
	footnoteErr    error
}

func (f *fakeQuerier) GetAllShowIDs(ctx context.Context) ([]int32, error) {
	return f.allShowIDs, f.allShowIDsErr
}

func (f *fakeQuerier) GetShowFromID(ctx context.Context, showID int32) ([]database.GetShowFromIDRow, error) {
	return f.showFromIDRows, f.showFromIDErr
}

func (f *fakeQuerier) GetShowFromDate(ctx context.Context, showDate time.Time) ([]database.GetShowFromDateRow, error) {
	return nil, nil
}

func (f *fakeQuerier) GetShowsBetweenDates(ctx context.Context, arg database.GetShowsBetweenDatesParams) ([]database.GetShowsBetweenDatesRow, error) {
	return nil, nil
}

func (f *fakeQuerier) GetShowsFromSetName(ctx context.Context, setName string) ([]database.GetShowsFromSetNameRow, error) {
	return nil, nil
}

func (f *fakeQuerier) GetShowsFromSongName(ctx context.Context, rawEntry string) ([]database.GetShowsFromSongNameRow, error) {
	return nil, nil
}

func (f *fakeQuerier) GetShowsFromState(ctx context.Context, stateOrCountry string) ([]database.GetShowsFromStateRow, error) {
	return nil, nil
}

func (f *fakeQuerier) GetShowsFromYear(ctx context.Context, year int32) ([]database.GetShowsFromYearRow, error) {
	return nil, nil
}

func (f *fakeQuerier) GetShowsFromYearAndState(ctx context.Context, arg database.GetShowsFromYearAndStateParams) ([]database.GetShowsFromYearAndStateRow, error) {
	return nil, nil
}

func (f *fakeQuerier) SearchByVenue(ctx context.Context, venue string) ([]database.SearchByVenueRow, error) {
	return nil, nil
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
	return nil, nil
}

// ----------------------------------------------------------------------------------------------------------------------------------
// ----------------------------------------------------------------------------------------------------------------------------------
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
		t.Errorf("status = %d; want 200", res.StatusCode)
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
		t.Errorf("status = %d; want 200", res.StatusCode)
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
