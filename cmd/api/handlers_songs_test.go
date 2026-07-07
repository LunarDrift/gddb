package main

import (
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
	err := json.NewDecoder(res.Body).Decode(&got)
	if err != nil {
		t.Fatalf("error decoding response: %v", err)
	}

	if got.FirstPlayed != "1990-01-01" {
		t.Errorf("got.FirstPlayed = %q; want '1990-01-01'", got.FirstPlayed)
	}
	if got.LastPlayed != "1990-01-02" {
		t.Errorf("got.LastPlayed = %q; want '1990-01-02", got.LastPlayed)
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
