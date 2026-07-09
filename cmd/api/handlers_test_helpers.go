package main

import (
	"context"
	"database/sql"
	"time"

	"github.com/LunarDrift/deadabase/internal/database"
)

// fakeQuerier is a fake 'database' with all the methods required to satisfy ShowQuerier. Used so
// tests don't require a connection to the real database
type fakeQuerier struct {
	// Shows
	allShowIDs                   []int32
	allShowIDsErr                error
	showFromIDRows               []database.GetShowFromIDRow
	showFromIDErr                error
	showFromDateRows             []database.GetShowFromDateRow
	showFromDateErr              error
	showsBetweenDatesRows        []database.GetShowsBetweenDatesRow
	showsBetweenDatesErr         error
	showsFromSongNameRows        []database.GetShowsFromSongNameRow
	showsFromSongNameErr         error
	showsFromSetNameRows         []database.GetShowsFromSetNameRow
	showsFromSetNameErr          error
	showsFromVenueNameRows       []database.SearchByVenueRow
	showsFromVenueNameErr        error
	showsFromLocationRows        []database.GetShowsFromLocationRow
	showsFromLocationErr         error
	showsFromYearRows            []database.GetShowsFromYearRow
	showsFromYearErr             error
	showsFromYearAndLocationRows []database.GetShowsFromYearAndLocationRow
	showsFromYearAndLocationErr  error
	showsWithNotesRows           []database.ShowsWithShowNotesRow
	showsWithNotesErr            error
	showsWithoutNotesRows        []database.ShowsWithoutNotesRow
	showsWithoutNotesErr         error
	// Songs
	songStatsRow            database.SongStatsRow
	songStatsErr            error
	songsPlayedAtVenueRows  []database.AllSongsPlayedAtVenueRow
	songsPlayedAtVenueErr   error
	songsFromSetNameRows    []database.MostCommonSongsBySetNameRow
	songsFromSetNameErr     error
	songsMostPlayedRows     []database.MostPlayedSongsRow
	songsMostplayedErr      error
	songsUniquePerCityRows  []database.UniqueSongsPerCityRow
	songsUniquePerCityErr   error
	songsPlayedLessThanRows []database.SongsPlayedLessThanRow
	songsPlayedLessThanErr  error
	// Other
	validLocationRows []string
	validLocationErr  error
	footnoteRows      []database.GetFootnotesFromShowIDRow
	footnoteErr       error
}

func (f *fakeQuerier) GetAllShowIDs(ctx context.Context) ([]int32, error) {
	return f.allShowIDs, f.allShowIDsErr
}

func (f *fakeQuerier) GetShowFromID(ctx context.Context, showID int32) ([]database.GetShowFromIDRow, error) {
	return f.showFromIDRows, f.showFromIDErr
}

func (f *fakeQuerier) GetShowFromDate(ctx context.Context, showDate time.Time) ([]database.GetShowFromDateRow, error) {
	return f.showFromDateRows, f.showFromDateErr
}

func (f *fakeQuerier) GetShowsBetweenDates(ctx context.Context, arg database.GetShowsBetweenDatesParams) ([]database.GetShowsBetweenDatesRow, error) {
	return f.showsBetweenDatesRows, f.showsBetweenDatesErr
}

func (f *fakeQuerier) GetShowsFromSetName(ctx context.Context, setName string) ([]database.GetShowsFromSetNameRow, error) {
	return f.showsFromSetNameRows, f.showsFromSetNameErr
}

func (f *fakeQuerier) GetShowsFromSongName(ctx context.Context, rawEntry string) ([]database.GetShowsFromSongNameRow, error) {
	return f.showsFromSongNameRows, f.showsFromSongNameErr
}

func (f *fakeQuerier) GetShowsFromLocation(ctx context.Context, stateOrCountry string) ([]database.GetShowsFromLocationRow, error) {
	return f.showsFromLocationRows, f.showsFromLocationErr
}

func (f *fakeQuerier) GetShowsFromYear(ctx context.Context, year int32) ([]database.GetShowsFromYearRow, error) {
	return f.showsFromYearRows, f.showsFromYearErr
}

func (f *fakeQuerier) GetShowsFromYearAndLocation(ctx context.Context, arg database.GetShowsFromYearAndLocationParams) ([]database.GetShowsFromYearAndLocationRow, error) {
	return f.showsFromYearAndLocationRows, f.showsFromYearAndLocationErr
}

func (f *fakeQuerier) SearchByVenue(ctx context.Context, venue string) ([]database.SearchByVenueRow, error) {
	return f.showsFromVenueNameRows, f.showsFromVenueNameErr
}

func (f *fakeQuerier) ShowsWithShowNotes(ctx context.Context) ([]database.ShowsWithShowNotesRow, error) {
	return f.showsWithNotesRows, f.showsWithNotesErr
}

func (f *fakeQuerier) ShowsWithoutNotes(ctx context.Context) ([]database.ShowsWithoutNotesRow, error) {
	return f.showsWithoutNotesRows, f.showsWithoutNotesErr
}

func (f *fakeQuerier) SongStats(ctx context.Context, songName sql.NullString) (database.SongStatsRow, error) {
	return f.songStatsRow, f.songStatsErr
}

func (f *fakeQuerier) AllSongsPlayedAtVenue(ctx context.Context, venue string) ([]database.AllSongsPlayedAtVenueRow, error) {
	return f.songsPlayedAtVenueRows, f.songsPlayedAtVenueErr
}

func (f *fakeQuerier) MostCommonSongsBySetName(ctx context.Context, setName string) ([]database.MostCommonSongsBySetNameRow, error) {
	return f.songsFromSetNameRows, f.songsFromSetNameErr
}

func (f *fakeQuerier) MostPlayedSongs(ctx context.Context) ([]database.MostPlayedSongsRow, error) {
	return f.songsMostPlayedRows, f.songsMostplayedErr
}

func (f *fakeQuerier) SongsPlayedLessThan(ctx context.Context, number int32) ([]database.SongsPlayedLessThanRow, error) {
	return f.songsPlayedLessThanRows, f.songsPlayedLessThanErr
}

func (f *fakeQuerier) UniqueSongsPerCity(ctx context.Context) ([]database.UniqueSongsPerCityRow, error) {
	return f.songsUniquePerCityRows, f.songsUniquePerCityErr
}

func (f *fakeQuerier) GetFootnotesFromShowID(ctx context.Context, showID int32) ([]database.GetFootnotesFromShowIDRow, error) {
	return f.footnoteRows, f.footnoteErr
}

func (f *fakeQuerier) GetValidLocations(ctx context.Context) ([]string, error) {
	return f.validLocationRows, f.validLocationErr
}
