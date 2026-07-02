package internal

import (
	"testing"
	"time"
)

func TestSetPosition(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  int
	}{
		{"acoustic first set", "acoustic_1", -2},
		{"acoustic second set", "acoustic_2", -1},
		{"acoustic", "acoustic", 0},
		{"electric", "electric", 1},
		{"encore sorts last", "encore", 999},
		{"regular set 1", "set_1", 1},
		{"regular set 2", "set_2", 2},
		{"unrecognized name sorts very last", "soundcheck", 1000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SetPosition(tt.input)
			if got != tt.want {
				t.Errorf("SetPosition(%q) = %d; want %d", tt.input, got, tt.want)
			}
		})
	}
}

func TestSortSetPositions(t *testing.T) {
	date, _ := time.Parse(time.DateOnly, "1977-05-08")

	input := []ShowSortInput{
		{ShowID: 1, Date: date, Venue: "Barton Hall", City: "Ithaca", State: "NY", SetName: "set_2", RawEntry: "Scarlet Begonias"},
		{ShowID: 1, Date: date, Venue: "Barton Hall", City: "Ithaca", State: "NY", SetName: "set_1", RawEntry: "Bertha"},
		{ShowID: 1, Date: date, Venue: "Barton Hall", City: "Ithaca", State: "NY", SetName: "encore", RawEntry: "One More Saturday Night"},
		{ShowID: 1, Date: date, Venue: "Barton Hall", City: "Ithaca", State: "NY", SetName: "set_1", RawEntry: "Me and My Uncle"},
	}

	got, err := SortSetPositions(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(got.Sets) != 3 { // set_1, set_2, encore
		t.Errorf("SortSetPositions(%v) = %v; want 3", input, got)
	}

	// set order check
	wantSetOrder := []string{"set_1", "set_2", "encore"}
	for i, want := range wantSetOrder {
		if got.Sets[i].SetName != want {
			t.Errorf("Sets[%d] = %q; want %q", i, got.Sets[i].SetName, want)
		}
	}

	// song order check
	wantSongOrder := []string{"Bertha", "Me and My Uncle"}
	if len(got.Sets[0].Songs) != len(wantSongOrder) {
		t.Fatalf("Sets[0].Songs = %v; want %v", got.Sets[0].Songs, wantSongOrder)
	}

	for i, want := range wantSongOrder {
		if got.Sets[0].Songs[i] != want {
			t.Errorf("Sets[0].Songs[%d] = %q; want %q", i, got.Sets[0].Songs[i], want)
		}
	}

	// venue check
	if got.Venue != "Barton Hall" {
		t.Errorf("ShowMeta.Venue = %q; want 'Barton Hall'", got.Venue)
	}
}

func TestSortSetPositions_EmptyInput(t *testing.T) {
	_, err := SortSetPositions([]ShowSortInput{})
	if err == nil {
		t.Error("expected error for empty input, got nil")
	}
}
