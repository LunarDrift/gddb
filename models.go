package main

// ShowData is just a type alias that will be used when unmarshaling
var ShowData map[string][]Show

// Show is used as a temporary container for unmarshaling the
// json object, before distributing it across the other tables
type Show struct {
	ShowID   int                 `json:"show_id"`
	Date     string              `json:"date"`
	Day      string              `json:"day"`
	Venue    string              `json:"venue"`
	Location string              `json:"location"`
	Notes    string              `json:"notes"`
	Setlist  map[string][]string `json:"setlist"`
}
