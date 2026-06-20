// Package internal .
package internal

import "time"

// Show is used as a temporary container for unmarshaling the
// json object, before distributing it across the other tables
// ================================================================
// NO LONGER NEEDED. WAS USED FOR IMPORTING JSON INTO DATABASE
// ================================================================
type Show struct {
	ShowID   int                 `json:"show_id"`
	Date     string              `json:"date"`
	Day      string              `json:"day"`
	Venue    string              `json:"venue"`
	Location string              `json:"location"`
	Notes    string              `json:"notes"`
	Setlist  map[string][]string `json:"setlist"`
}

// Dataset is a type alias to be used when unmarshaling json file
// ================================================================
// NO LONGER NEEDED. WAS USED FOR IMPORTING JSON INTO DATABASE
// ================================================================
type Dataset map[string][]Show

// ===================================================================================================================
// ===================================================================================================================
// ===================================================================================================================

// ShowResponse will be used as the payload sent in the server response
type ShowResponse struct {
	Date  string        `json:"date"`
	Venue string        `json:"venue"`
	City  string        `json:"city"`
	State string        `json:"state"`
	Notes string        `json:"notes"`
	Sets  []SetResponse `json:"sets"`
}

// SetResponse holds the set name (ie. set_1, set_2, encore, etc.) and list of songs
type SetResponse struct {
	SetName string   `json:"set_name"`
	Songs   []string `json:"songs"`
}

// ShowSortInput is used to hold the data from sqlc-generated structs while the setlist gets sorted
type ShowSortInput struct {
	ShowDate time.Time `json:"show_date"`
	Venue    string    `json:"venue"`
	City     string    `json:"city"`
	State    string    `json:"state"`
	Notes    string    `json:"notes"`
	SetName  string    `json:"set_name"`
	RawEntry string    `json:"raw_entry"`
}

type VenueSearchResult struct {
	ShowID int    `json:"show_id"`
	Date   string `json:"date"`
	Venue  string `json:"venue"`
	City   string `json:"city"`
	State  string `json:"state"`
}

type ShowWithNoSetlist struct {
	Date    string `json:"date"`
	Venue   string `json:"venue"`
	City    string `json:"city"`
	State   string `json:"state"`
	Notes   string `json:"notes"`
	Message string `json:"message"`
}
