// Package internal .
package internal

import "time"

// Show is used as a temporary container for unmarshaling the
// json object, before distributing it across the other tables
type Show struct {
	ShowID    int                 `json:"show_id"`
	Date      string              `json:"date"`
	Day       string              `json:"day"`
	Venue     string              `json:"venue"`
	Location  string              `json:"location"`
	Notes     string              `json:"notes"`
	Setlist   map[string][]string `json:"setlist"`
	Footnotes map[string]string   `json:"footnotes"`
}

// Dataset is a type alias to be used when unmarshaling json file
type Dataset map[string][]Show

// ===================================================================================================================
// ===================================================================================================================
// ===================================================================================================================

// ShowMeta holds the shared elements that multiple show responses use
type ShowMeta struct {
	Date  string `json:"date"`
	Venue string `json:"venue"`
	City  string `json:"city"`
	State string `json:"state"`
	Notes string `json:"notes"`
}

// ShowResponse will be used as the payload sent in the server response
type ShowResponse struct {
	ShowMeta
	Sets      []SetResponse     `json:"sets"`
	Footnotes map[string]string `json:"footnotes"`
}

type ShowWithNoSetlist struct {
	ShowMeta
	Message string `json:"message"`
}

// SetResponse holds the set name (ie. set_1, set_2, encore, etc.) and list of songs
type SetResponse struct {
	SetName string   `json:"set_name"`
	Songs   []string `json:"songs"`
}

// ShowSortInput is used to hold the data from sqlc-generated structs while the setlist gets sorted
type ShowSortInput struct {
	ShowID   int32     `json:"show_id"`
	ShowDate time.Time `json:"show_date"`
	Venue    string    `json:"venue"`
	City     string    `json:"city"`
	State    string    `json:"state"`
	Notes    string    `json:"notes"`
	SetName  string    `json:"set_name"`
	RawEntry string    `json:"raw_entry"`
}

type ListOfShowsResult struct {
	ShowID int    `json:"show_id"`
	Date   string `json:"date"`
	Venue  string `json:"venue"`
	City   string `json:"city"`
	State  string `json:"state"`
}
