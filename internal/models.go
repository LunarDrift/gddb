// Package internal .
package internal

import (
	"context"
	"database/sql"
	"time"

	"github.com/LunarDrift/deadabase/internal/database"
)

// ===================================================================================================================
// FOR IMPORTING
// ===================================================================================================================

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
// Also used in responses where a list of shows is returned
type ShowMeta struct {
	ShowID   int32  `json:"show_id"`
	Date     string `json:"date"`
	Venue    string `json:"venue"`
	City     string `json:"city"`
	Location string `json:"location"`
	Notes    string `json:"notes"`
}

// ShowResponse will be used as the payload sent in the server response for a single show
type ShowResponse struct {
	ShowMeta
	Sets      []SetResponse     `json:"sets"`
	Footnotes map[string]string `json:"footnotes"`
}

// SetResponse holds the set name (i.e. set_1, set_2, encore, etc.) and list of songs
type SetResponse struct {
	SetName string   `json:"set_name"`
	Songs   []string `json:"songs"`
}

// ShowWithNoSetlist has all other ShowMeta details with a custom message informing about no set list available
type ShowWithNoSetlist struct {
	ShowMeta
	Message string `json:"message"`
}

// ShowSortInput is used to hold the data from sqlc-generated structs while the setlist gets sorted
type ShowSortInput struct {
	ShowID   int32     `json:"show_id"`
	Date     time.Time `json:"date"`
	Venue    string    `json:"venue"`
	City     string    `json:"city"`
	Location string    `json:"location"`
	Notes    string    `json:"notes"`
	SetName  string    `json:"set_name"`
	RawEntry string    `json:"raw_entry"`
}

type SongsTimesPlayed struct {
	Song        string `json:"song"`
	TimesPlayed int    `json:"times_played"`
}

type SongStats struct {
	TimesPlayed int    `json:"times_played"`
	FirstPlayed string `json:"first_played"`
	LastPlayed  string `json:"last_played"`
}

type SongsFromVenue struct {
	SongName string `json:"song"`
	Venue    string `json:"venue"`
	City     string `json:"city"`
	Location string `json:"location"`
}

type UniqueSongsPerCity struct {
	City            string `json:"city"`
	Location        string `json:"location"`
	UniqueSongCount int    `json:"unique_song_count"`
}

// ShowQuerier is needed in order to be able to make unit tests for the endpoints
// without relying on a connection to the database
type ShowQuerier interface {
	GetAllShowIDs(ctx context.Context) ([]int32, error)
	GetShowFromDate(ctx context.Context, showDate time.Time) ([]database.GetShowFromDateRow, error)
	GetShowFromID(ctx context.Context, showID int32) ([]database.GetShowFromIDRow, error)
	GetShowsBetweenDates(ctx context.Context, arg database.GetShowsBetweenDatesParams) ([]database.GetShowsBetweenDatesRow, error)
	GetShowsFromSetName(ctx context.Context, setName string) ([]database.GetShowsFromSetNameRow, error)
	GetShowsFromSongName(ctx context.Context, rawEntry string) ([]database.GetShowsFromSongNameRow, error)
	GetShowsFromLocation(ctx context.Context, location string) ([]database.GetShowsFromLocationRow, error)
	GetShowsFromCity(ctx context.Context, city string) ([]database.GetShowsFromCityRow, error)
	GetShowsFromYear(ctx context.Context, year int32) ([]database.GetShowsFromYearRow, error)
	GetShowsFromYearAndLocation(ctx context.Context, arg database.GetShowsFromYearAndLocationParams) ([]database.GetShowsFromYearAndLocationRow, error)
	SearchByVenue(ctx context.Context, venue string) ([]database.SearchByVenueRow, error)
	ShowsWithShowNotes(ctx context.Context) ([]database.ShowsWithShowNotesRow, error)
	ShowsWithoutNotes(ctx context.Context) ([]database.ShowsWithoutNotesRow, error)
	SongStats(ctx context.Context, songName sql.NullString) (database.SongStatsRow, error)
	AllSongsPlayedAtVenue(ctx context.Context, venue string) ([]database.AllSongsPlayedAtVenueRow, error)
	MostCommonSongsBySetName(ctx context.Context, setName string) ([]database.MostCommonSongsBySetNameRow, error)
	MostPlayedSongs(ctx context.Context) ([]database.MostPlayedSongsRow, error)
	SongsPlayedLessThan(ctx context.Context, number int32) ([]database.SongsPlayedLessThanRow, error)
	UniqueSongsPerCity(ctx context.Context) ([]database.UniqueSongsPerCityRow, error)
	GetFootnotesFromShowID(ctx context.Context, showID int32) ([]database.GetFootnotesFromShowIDRow, error)
	GetValidLocations(ctx context.Context) ([]string, error)
}
