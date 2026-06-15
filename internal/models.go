// Package internal .
package internal

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

// Dataset is a type alias to be used when unmarshaling json file
type Dataset map[string][]Show

// ShowResponse will be used as the payload sent in the server response
type ShowResponse struct {
	Date  string        `json:"date"`
	Venue string        `json:"venue"`
	Sets  []SetResponse `json:"sets"`
}

// SetResponse holds the set name (ie. set_1, set_2, encore, etc.) and list of songs
type SetResponse struct {
	SetName string   `json:"set_name"`
	Songs   []string `json:"songs"`
}
