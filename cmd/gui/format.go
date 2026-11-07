package main

import (
	"fmt"
	"maps"
	"slices"
	"strings"
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

	var date, sets, notes, location strings.Builder
	first := data[0]
	fmt.Fprintf(&date, "## Date: %v\n\n", first.Date)

	for i, show := range data {
		fmt.Fprintf(&date, "##### Show ID: %v\n", show.ShowID)
		// fmt.Fprintf(&sets, "### Show ID: %v\n\n", show.ShowID)
		fmt.Fprint(&location, "## Location(s):\n\n")
		fmt.Fprintf(&location, "**%v · %v, %v**\n\n", show.Venue, show.City, show.Location)
		if show.Notes != "" {
			fmt.Fprintf(&location, "\n\n\n**Notes:** %v\n\n", show.Notes)
		}
		fmt.Fprintf(&date, "[Listen on archive.org](%s)\n\n", buildArchiveOrgLink(first.Date))
		sets.WriteString("\n\n")

		if len(show.Sets) == 0 {
			sets.WriteString("*No setlist available for this show*\n")
		}
		for _, set := range show.Sets {
			fmt.Fprintf(&sets, "**%s:**\n\n%s\n\n", formatSetName(set.SetName), strings.Join(formatSongs(set), "\n\n\n"))
		}

		if len(show.Footnotes) > 0 {
			fmt.Fprintf(&notes, "#### Show %v Footnotes\n\n", show.ShowID)
			for _, k := range slices.Sorted(maps.Keys(show.Footnotes)) {
				fmt.Fprintf(&notes, "[%s]: *%v*\n\n", k, show.Footnotes[k])
			}
			notes.WriteString("\n")
		}

		if i < len(data)-1 {
			fmt.Fprintln(&sets, strings.Repeat("━", 60)+"\n")
		}
	}

	return showText{
		date:     strings.TrimSpace(date.String()),
		location: strings.TrimSpace(location.String()),
		// location:  fmt.Sprintf("Location: %v - %v, %v", first.Venue, first.City, first.Location),
		sets:      strings.TrimSpace(sets.String()),
		footnotes: strings.TrimSpace(notes.String()),
	}
}

// formatSetName turns "set_1" to "Set 1" etc...
func formatSetName(s string) string {
	words := strings.Split(strings.TrimSpace(s), "_")
	for i, word := range words {
		if word == "" {
			continue
		}
		r, size := utf8.DecodeRuneInString(word)
		words[i] = string(unicode.ToUpper(r)) + word[size:]
	}
	return strings.Join(words, " ")
}

func monthToNumString(name string) string {
	switch name {
	case "January":
		return "1"
	case "February":
		return "2"
	case "March":
		return "3"
	case "April":
		return "4"
	case "May":
		return "5"
	case "June":
		return "6"
	case "July":
		return "7"
	case "August":
		return "8"
	case "September":
		return "9"
	case "October":
		return "10"
	case "November":
		return "11"
	case "December":
		return "12"
	default:
		return ""
	}
}

func buildArchiveOrgLink(date string) string {
	if date == "" {
		return ""
	}
	return fmt.Sprintf("https://archive.org/details/GratefulDead?tab=collection&query=%s", date)
}

func formatSongs(data internal.SetResponse) []string {
	if len(data.Songs) == 0 {
		return nil
	}
	out := []string{}
	for i, song := range data.Songs {
		out = append(out, fmt.Sprintf("%d. %s\n\n", i+1, song))
	}
	return out
}
