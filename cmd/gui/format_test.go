package main

import (
	"testing"

	"github.com/LunarDrift/deadabase/internal"
)

func TestFormatSetName(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"set name with underscore", "set_1", "Set 1"},
		{"no underscore", "acoustic", "Acoustic"},
		{"extra whitespace", "   encore ", "Encore"},
		{"empty input", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatSetName(tt.input)
			if got != tt.want {
				t.Errorf("got %q; want %q", got, tt.want)
			}
		})
	}
}

func TestFormatShows(t *testing.T) {
	normalInput := []internal.ShowResponse{
		{
			ShowID:   1,
			Date:     "1990-07-09",
			Venue:    "Soldier Field",
			City:     "Chicago",
			Location: "IL",
			Notes:    "Last show",
			Sets: []internal.SetResponse{
				{
					SetName: "set_1",
					Songs:   []string{"song 1", "song 2", "song 3"},
				},
				{
					SetName: "set_2",
					Songs:   []string{"song 4", "song 5", "song 6 > song 7 > song 6"},
				},
			},
		},
	}
	noSongs := []internal.ShowResponse{
		{
			ShowID:   1,
			Date:     "1990-07-09",
			Venue:    "Soldier Field",
			City:     "Chicago",
			Location: "IL",
			Notes:    "",
			Sets:     []internal.SetResponse{},
		},
	}

	tests := []struct {
		name string
		in   []internal.ShowResponse
		want showText
	}{
		{name: "typical show input", in: normalInput, want: showText{
			date:      "Date: 1990-07-09",
			location:  "Location(s): Soldier Field · Chicago, IL",
			sets:      "Show ID: 1\nNotes: Last show\n\nSet 1:\nsong 1\nsong 2\nsong 3\n\nSet 2:\nsong 4\nsong 5\nsong 6 > song 7 > song 6",
			footnotes: "",
		}},
		{name: "show without set list", in: noSongs, want: showText{
			date:      "Date: 1990-07-09",
			location:  "Location(s): Soldier Field · Chicago, IL",
			sets:      "Show ID: 1\n\nNo setlist available for this show",
			footnotes: "",
		}},
		{"empty show response", []internal.ShowResponse{}, showText{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatShows(tt.in)
			if got.date != tt.want.date {
				t.Errorf("got.date = %q; want %q", got.date, tt.want.date)
			}
			if got.location != tt.want.location {
				t.Errorf("got.location = %q; want %q", got.location, tt.want.location)
			}
			if got.sets != tt.want.sets {
				t.Errorf("got.sets = %q; want %q", got.sets, tt.want.sets)
			}
			if got.footnotes != tt.want.footnotes {
				t.Errorf("got.footnotes = %q; want %q", got.footnotes, tt.want.footnotes)
			}
		})
	}
}
