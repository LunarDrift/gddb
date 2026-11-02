package importer

import (
	"slices"
	"testing"
)

func TestParseEntry(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{
			name:  "song chain w/ footnote markers",
			input: "Help On The Way ** > Slipknot! ** > Franklin's Tower > Playing In The Band > drums > space > Throwing Stones > Not Fade Away > Black Peter > Sugar Magnolia",
			want:  []string{"Help On The Way", "Slipknot!", "Franklin's Tower", "Playing In The Band", "drums", "space", "Throwing Stones", "Not Fade Away", "Black Peter", "Sugar Magnolia"},
		},
		{
			name:  "single song",
			input: "Alabama Getaway",
			want:  []string{"Alabama Getaway"},
		},
		{
			name:  "Comma in title shouldn't get split",
			input: "It's All Over Now, Baby Blue",
			want:  []string{"It's All Over Now, Baby Blue"},
		},
		{
			name:  "single song with trailing marker",
			input: "Alabama Getaway **",
			want:  []string{"Alabama Getaway"},
		},
		{
			name:  "marker on last song of chain",
			input: "Sugar Magnolia > Sunshine Daydream **",
			want:  []string{"Sugar Magnolia", "Sunshine Daydream"},
		},
		{
			name:  "extra whitespace",
			input: "  Cassidy   >   Deal  ",
			want:  []string{"Cassidy", "Deal"},
		},
		{
			name:  "empty input",
			input: "",
			want:  []string{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := parseEntry(tc.input)
			if !slices.Equal(got, tc.want) {
				t.Errorf("parseEntry(%q) = %q; want %q", tc.input, got, tc.want)
			}
		})
	}
}
