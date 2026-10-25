package main

import (
	"regexp"
	"strings"
)

var markerRe = regexp.MustCompile(`\s*(\*+)\s*$`)

func parseEntry(raw string) []string {
	parts := strings.Split(raw, ">")
	songs := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(markerRe.ReplaceAllString(part, ""))
		if part != "" {
			songs = append(songs, part)
		}
	}
	return songs
}
