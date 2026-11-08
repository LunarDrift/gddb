package main

import (
	"database/sql"
	"testing"
)

func TestFuzzyPattern(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"Hello world", "%Hello%world%"},
		{"Althea", "%Althea%"},
		{"Dark Star > drums > space > Dark Star", "%Dark%Star%>%drums%>%space%>%Dark%Star%"},
		{"", ""},
		{"Dark  Star", "%Dark%Star%"},  // two spaces
		{" Dark Star ", "%Dark%Star%"}, // leading and trailing whitespace
		{"a", "%a%"},
		{"100% Fun", "%100%%Fun%"},
	}

	for _, tt := range tests {
		got := fuzzyPattern(tt.in)
		if got != tt.want {
			t.Errorf("fuzzyPattern(%q) = %q; want %q", tt.in, got, tt.want)
		}
	}
}

func TestValidateDateParts(t *testing.T) {
	tests := []struct {
		name  string
		year  string
		month string
		day   string
		want  dateFilters
	}{
		{name: "full date", year: "1970", month: "09", day: "24", want: dateFilters{year: 1970, month: 9, day: 24}},
		{name: "year only", year: "1970", month: "", day: "", want: dateFilters{year: 1970}},
		{name: "month only", year: "", month: "12", day: "", want: dateFilters{month: 12}},
		{name: "day only", year: "", month: "", day: "31", want: dateFilters{day: 31}},
		{name: "year & month", year: "1980", month: "5", day: "", want: dateFilters{year: 1980, month: 5}},
		{name: "year & day", year: "1980", month: "", day: "10", want: dateFilters{year: 1980, day: 10}},
		{name: "month & day", year: "", month: "09", day: "24", want: dateFilters{month: 9, day: 24}},
		{name: "empty input", year: "", month: "", day: "", want: dateFilters{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := validateDateParts(tt.year, tt.month, tt.day)
			if err != nil {
				t.Fatalf("error: %v", err)
			}
			if got.year != tt.want.year {
				t.Errorf("got.year = %d; want %d", got.year, tt.want.year)
			}
			if got.month != tt.want.month {
				t.Errorf("got.month = %d; want %d", got.month, tt.want.month)
			}
			if got.day != tt.want.day {
				t.Errorf("got.day = %d; want %d", got.day, tt.want.day)
			}
		})
	}
}

func TestStrToNullBool(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  sql.NullBool
	}{
		{name: "empty input", input: "", want: sql.NullBool{}},
		{name: "true", input: "true", want: sql.NullBool{Bool: true, Valid: true}},
		{name: "false", input: "false", want: sql.NullBool{Bool: false, Valid: true}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := strToNullBool(tt.input)
			if err != nil {
				t.Fatalf("error: %v", err)
			}
			if got.Bool != tt.want.Bool {
				t.Errorf("got.Bool = %v; want %v", got.Bool, tt.want.Bool)
			}
			if got.Valid != tt.want.Valid {
				t.Errorf("got.Valid = %v; want %v", got.Valid, tt.want.Valid)
			}
		})
	}
}
