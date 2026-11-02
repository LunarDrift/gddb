package main

import (
	"errors"
	"fmt"
	"maps"
	"slices"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/LunarDrift/deadabase/internal"
)

// showText is everything the details panel needs, already formatted.
type showText struct {
	date      string
	location  string
	sets      string
	footnotes string
}

// formatShows renders one or more shows (same date) for display.
func formatShows(data []internal.ShowResponse) showText {
	if len(data) == 0 {
		return showText{}
	}

	var sets, notes, location strings.Builder
	for i, show := range data {
		fmt.Fprintf(&sets, "Show ID: %v\n", show.ShowID)
		fmt.Fprintf(&location, "Location(s): %v - %v, %v\n", show.Venue, show.City, show.Location)
		if show.Notes != "" {
			fmt.Fprintf(&sets, "Notes: %v\n", show.Notes)
		}
		sets.WriteString("\n")

		if len(show.Sets) == 0 {
			sets.WriteString("No setlist available for this show\n")
		}
		for _, set := range show.Sets {
			fmt.Fprintf(&sets, "%s:\n%s\n\n", formatSetName(set.SetName), strings.Join(set.Songs, "\n"))
		}

		if len(show.Footnotes) > 0 {
			fmt.Fprintf(&notes, "━━━ Show %v Footnotes ━━━\n", show.ShowID)
			for _, k := range slices.Sorted(maps.Keys(show.Footnotes)) {
				fmt.Fprintf(&notes, "[%s] %v\n", k, show.Footnotes[k])
			}
			notes.WriteString("\n")
		}

		if i < len(data)-1 {
			fmt.Fprintln(&sets, strings.Repeat("━", 60)+"\n")
		}
	}

	first := data[0]
	return showText{
		date:     fmt.Sprintf("Date: %v", first.Date),
		location: strings.TrimSpace(location.String()),
		// location:  fmt.Sprintf("Location: %v - %v, %v", first.Venue, first.City, first.Location),
		sets:      strings.TrimSpace(sets.String()),
		footnotes: strings.TrimSpace(notes.String()),
	}
}

// formatSetName turns "set_1" to "Set 1" etc...
func formatSetName(s string) string {
	words := strings.Split(s, "_")
	for i, word := range words {
		if word == "" {
			continue
		}
		r, size := utf8.DecodeRuneInString(word)
		words[i] = string(unicode.ToUpper(r)) + word[size:]
	}
	return strings.Join(words, " ")
}

// buildDate validates the date picker values and returns YYYY-MM-DD.
// month and day are 1-based; pass 0 (or less) for "nothing selected".
func buildDate(yearStr string, month, day int) (string, error) {
	year, err := strconv.Atoi(yearStr)
	if err != nil || month < 1 || day < 1 {
		return "", errors.New("choose a year, month, and day")
	}
	t := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
	if t.Day() != day { // Feb 31 silently rolls over to Mar 3
		return "", errors.New("that date doesn't exist")
	}
	return t.Format(time.DateOnly), nil
}
