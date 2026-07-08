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

func TestHandleGetSongStats(t *testing.T) {
	firstPlayed, _ := time.Parse(time.DateOnly, "1990-01-01")
	lastPlayed := firstPlayed.Add(time.Hour * 24)
	fake := &fakeQuerier{
		songStatsRow: database.SongStatsRow{
			TimesPlayed: 10, FirstPlayed: firstPlayed, LastPlayed: lastPlayed,
		},
	}

	s := &server{queries: fake}
	req := httptest.NewRequest(http.MethodGet, "/songs/althea", nil)
	req.SetPathValue("song", "althea")
	w := httptest.NewRecorder()

	s.handleGetSongStats(w, req)

	res := w.Result()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("status code = %d; want %d", res.StatusCode, http.StatusOK)
	}

	var got internal.SongStats
	if err := json.NewDecoder(res.Body).Decode(&got); err != nil {
		t.Fatalf("error decoding response: %v", err)
	}

	if got.FirstPlayed != "1990-01-01" {
		t.Errorf("got.FirstPlayed = %q; want '1990-01-01'", got.FirstPlayed)
	}
	if got.LastPlayed != "1990-01-02" {
		t.Errorf("got.LastPlayed = %q; want '1990-01-02'", got.LastPlayed)
	}
	if got.TimesPlayed != 10 {
		t.Errorf("got.TimesPlayed = %d; want 10", got.TimesPlayed)
	}
}

func TestHandleGetSongStats_MissingPathParam(t *testing.T) {
	fake := &fakeQuerier{}
	s := &server{queries: fake}
	req := httptest.NewRequest(http.MethodGet, "/songs/", nil)
	w := httptest.NewRecorder()

	s.handleGetSongStats(w, req)
	res := w.Result()
	if res.StatusCode != http.StatusBadRequest {
		t.Errorf("status code = %d; want %d", res.StatusCode, http.StatusBadRequest)
	}
}

func TestHandleGetSongsPlayedAtVenue(t *testing.T) {
	fake := &fakeQuerier{
		songsPlayedAtVenueRows: []database.AllSongsPlayedAtVenueRow{
			{SongName: sql.NullString{String: "Althea", Valid: true}, Venue: "Soldier Field", City: "Chicago", State: "IL"},
			{SongName: sql.NullString{String: "Dark Star", Valid: true}, Venue: "Soldier Field", City: "Chicago", State: "IL"},
		},
	}

	s := &server{queries: fake}
	req := httptest.NewRequest(http.MethodGet, "/songs?venue=soldier_field", nil)
	w := httptest.NewRecorder()

	s.handleGetSongsPlayedAtVenue(w, req)

	res := w.Result()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("status code = %d; want %d", res.StatusCode, http.StatusOK)
	}

	var got []internal.SongsFromVenue
	if err := json.NewDecoder(res.Body).Decode(&got); err != nil {
		t.Fatalf("error decoding response: %v", err)
	}

	if len(got) != 2 {
		t.Errorf("len(got) = %d; want 2", len(got))
	}

	if got[0].Venue != "Soldier Field" {
		t.Errorf("got[0].Venue = %q; want 'Soldier Field'", got[0].Venue)
	}

	if got[1].Venue != "Soldier Field" {
		t.Errorf("got[1].Venue = %q; want 'Soldier Field'", got[1].Venue)
	}

	if got[0].Venue != got[1].Venue {
		t.Errorf("different venues: got[0].Venue = %q; got[1].Venue = %q", got[0].Venue, got[1].Venue)
	}
}

func TestHandleGetSongsPlayedAtVenue_Errors(t *testing.T) {
	tests := []struct {
		name       string
		url        string
		wantStatus int
	}{
		{"missing venue param", "/songs?venue=", http.StatusBadRequest},
		{"invalid venue param", "/songs?venue=hello_world", http.StatusNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &server{queries: &fakeQuerier{}}
			req := httptest.NewRequest(http.MethodGet, tt.url, nil)
			w := httptest.NewRecorder()
			s.handleGetSongsPlayedAtVenue(w, req)
			if got := w.Result().StatusCode; got != tt.wantStatus {
				t.Errorf("status = %d; want %d", got, tt.wantStatus)
			}
		})
	}
}

func TestHandleGetMostPlayedSongsBySetName(t *testing.T) {
	fake := &fakeQuerier{
		songsFromSetNameRows: []database.MostCommonSongsBySetNameRow{
			{Song: sql.NullString{String: "Dark Star", Valid: true}, TimesPlayed: 10},
			{Song: sql.NullString{String: "Althea", Valid: true}, TimesPlayed: 8},
		},
	}

	s := &server{queries: fake}
	req := httptest.NewRequest(http.MethodGet, "/songs?sort=most_played&set_name=encore", nil)
	w := httptest.NewRecorder()

	s.handleGetMostPlayedSongsBySetName(w, req)

	res := w.Result()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("status = %d; want %d", res.StatusCode, http.StatusOK)
	}

	var got []internal.SongsTimesPlayed
	if err := json.NewDecoder(res.Body).Decode(&got); err != nil {
		t.Fatalf("error decoding response: %v", err)
	}

	if len(got) != 2 {
		t.Errorf("len(got) = %d; want 2", len(got))
	}

	if got[0].TimesPlayed != 10 {
		t.Errorf("got[0].TimesPlayed = %d; want 10", got[0].TimesPlayed)
	}
	if got[1].TimesPlayed != 8 {
		t.Errorf("got[1].TimesPlayed = %d; want 8", got[1].TimesPlayed)
	}
}

func TestHandleGetMostPlayedSongsBySetName_Errors(t *testing.T) {
	tests := []struct {
		name       string
		url        string
		wantStatus int
	}{
		{"missing set name param", "/songs?set_name=", http.StatusBadRequest},
		{"invalid set name param", "/songs?set_name=hello", http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &server{queries: &fakeQuerier{}}
			req := httptest.NewRequest(http.MethodGet, tt.url, nil)
			w := httptest.NewRecorder()
			s.handleGetMostPlayedSongsBySetName(w, req)
			if got := w.Result().StatusCode; got != tt.wantStatus {
				t.Errorf("status code = %d; want %d", got, tt.wantStatus)
			}
		})
	}
}

func TestHandleGetMostPlayedSongs(t *testing.T) {
	fake := &fakeQuerier{
		songsMostPlayedRows: []database.MostPlayedSongsRow{
			{Song: sql.NullString{String: "Althea", Valid: true}, TimesPlayed: 100},
			{Song: sql.NullString{String: "Dark Star", Valid: true}, TimesPlayed: 90},
		},
	}

	s := &server{queries: fake}
	req := httptest.NewRequest(http.MethodGet, "/songs?sort=most_played", nil)
	w := httptest.NewRecorder()

	s.handleGetMostPlayedSongs(w, req)

	res := w.Result()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("status code = %d; want %d", res.StatusCode, http.StatusOK)
	}

	var got []internal.SongsTimesPlayed
	if err := json.NewDecoder(res.Body).Decode(&got); err != nil {
		t.Fatalf("error decoding response: %v", err)
	}

	if len(got) != 2 {
		t.Errorf("len(got) = %d; want 2", len(got))
	}

	if got[0].Song != "Althea" {
		t.Errorf("got[0].Song = %q; want 'Althea'", got[0].Song)
	}
	if got[1].Song != "Dark Star" {
		t.Errorf("got[1].Song = %q; want 'Dark Star'", got[1].Song)
	}
}

func TestHandleUniqueSongsPerCity(t *testing.T) {
	fake := &fakeQuerier{
		songsUniquePerCityRows: []database.UniqueSongsPerCityRow{
			{City: "Chicago", StateOrCountry: "IL", UniqueSongCount: 50},
			{City: "London", StateOrCountry: "England", UniqueSongCount: 24},
		},
	}

	s := &server{queries: fake}
	req := httptest.NewRequest(http.MethodGet, "/stats/songs-per-city", nil)
	w := httptest.NewRecorder()

	s.handleGetUniqueSongsPerCity(w, req)

	res := w.Result()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("status code = %d; want %d", res.StatusCode, http.StatusOK)
	}

	var got []internal.UniqueSongsPerCity
	if err := json.NewDecoder(res.Body).Decode(&got); err != nil {
		t.Fatalf("error decoding response: %v", err)
	}

	if len(got) != 2 {
		t.Errorf("len(got) = %d; want 2", len(got))
	}

	if got[0].City != "Chicago" {
		t.Errorf("got[0].City = %q; want 'Chicago'", got[0].City)
	}

	if got[1].StateOrCountry != "England" {
		t.Errorf("got[1].StateOrCountry = %q; want 'England'", got[1].StateOrCountry)
	}
	if got[1].UniqueSongCount != 24 {
		t.Errorf("got[1].UniqueSongCount = %d; want 24", got[1].UniqueSongCount)
	}
}

func TestHandleSongsPlayedLessThanNTimes(t *testing.T) {
	fake := &fakeQuerier{
		songsPlayedLessThanRows: []database.SongsPlayedLessThanRow{
			{Song: sql.NullString{String: "Althea", Valid: true}, TimesPlayed: 19},
			{Song: sql.NullString{String: "Dark Star", Valid: true}, TimesPlayed: 10},
		},
	}

	s := &server{queries: fake}
	req := httptest.NewRequest(http.MethodGet, "/songs?played_lt=20", nil)
	w := httptest.NewRecorder()

	s.handleGetSongsPlayedLessThanNTimes(w, req)

	res := w.Result()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("status code = %d; want %d", res.StatusCode, http.StatusOK)
	}

	var got []internal.SongsTimesPlayed
	if err := json.NewDecoder(res.Body).Decode(&got); err != nil {
		t.Fatalf("error decoding response: %v", err)
	}

	if len(got) != 2 {
		t.Errorf("len(got) = %d; want 2", len(got))
	}

	if got[0].Song != "Althea" {
		t.Errorf("got[0].Song = %q; want 'Althea'", got[0].Song)
	}
	if got[0].TimesPlayed != 19 {
		t.Errorf("got[0].TimesPlayed = %d; want 19", got[0].TimesPlayed)
	}

	if got[1].Song != "Dark Star" {
		t.Errorf("got[1].Song = %q; want 'Dark Star'", got[1].Song)
	}
}

func TestHandleSongsPlayedLessThanNTimes_Errors(t *testing.T) {
	tests := []struct {
		name       string
		url        string
		wantStatus int
	}{
		{"missing set name param", "/songs?played_lt=", http.StatusBadRequest},
		{"invalid set name param", "/songs?set_name=five", http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &server{queries: &fakeQuerier{}}
			req := httptest.NewRequest(http.MethodGet, tt.url, nil)
			w := httptest.NewRecorder()
			s.handleGetMostPlayedSongsBySetName(w, req)
			if got := w.Result().StatusCode; got != tt.wantStatus {
				t.Errorf("status code = %d; want %d", got, tt.wantStatus)
			}
		})
	}
}
