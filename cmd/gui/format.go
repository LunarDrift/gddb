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
	header    string
	showNotes string
	setlist   string
	songNotes string
}

// formatShows renders one or more shows (same date) for display.
func formatShows(data []internal.ShowResponse) showText {
	if len(data) == 0 {
		return showText{}
	}

	var header, setlist, songNotes, showNotes strings.Builder
	first := data[0]
	fmt.Fprintf(&header, "# %v\n", first.Date)

	for _, show := range data {
		fmt.Fprintf(&header, "**Show ID:** %v\n\n", show.ShowID)
		if show.Notes != "" {
			fmt.Fprintf(&showNotes, "%v\n", show.Notes)
		}

		if len(show.Sets) == 0 {
			setlist.WriteString("\n*No setlist available for this show*\n")
		}
		for _, set := range show.Sets {
			fmt.Fprintf(&setlist, "### **%s:**\n\n", formatSetName(set.SetName))
			fmt.Fprintf(&setlist, "%s\n", strings.Join(formatSongs(set), "\n"))
		}
		if len(show.Sets) > 0 && len(data) == 2 {
			fmt.Fprint(&setlist, "\n---\n")
		}

		if len(show.Footnotes) > 0 {
			fmt.Fprintf(&songNotes, "#### Show %v Footnotes\n\n", show.ShowID)
			for _, k := range slices.Sorted(maps.Keys(show.Footnotes)) {
				fmt.Fprintf(&songNotes, "[%s]: *%v*\n\n", k, show.Footnotes[k])
			}
		}
	}
	fmt.Fprintf(&header, "**%v · %v, %v**\n\n", first.Venue, first.City, first.Location)
	fmt.Fprintf(&header, "[Listen on archive.org](%s)", buildArchiveOrgLink(first.Date))

	return showText{
		header:    strings.TrimSpace(header.String()),
		showNotes: strings.TrimSpace(showNotes.String()),
		setlist:   strings.TrimSpace(setlist.String()),
		songNotes: strings.TrimSpace(songNotes.String()),
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
