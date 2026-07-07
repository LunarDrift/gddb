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
	allShowIDs                []int32
	allShowIDsErr             error
	showFromIDRows            []database.GetShowFromIDRow
	showFromIDErr             error
	footnoteRows              []database.GetFootnotesFromShowIDRow
	footnoteErr               error
	showFromDateRows          []database.GetShowFromDateRow
	showFromDateErr           error
	showsBetweenDatesRows     []database.GetShowsBetweenDatesRow
	showsBetweenDatesErr      error
	showsFromSongNameRows     []database.GetShowsFromSongNameRow
	showsFromSongNameErr      error
	showsFromSetNameRows      []database.GetShowsFromSetNameRow
	showsFromSetNameErr       error
	showsFromVenueNameRows    []database.SearchByVenueRow
	showsFromVenueNameErr     error
	showsFromStateRows        []database.GetShowsFromStateRow
	showsFromStateErr         error
	showsFromYearRows         []database.GetShowsFromYearRow
	showsFromYearErr          error
	showsFromYearAndStateRows []database.GetShowsFromYearAndStateRow
	showsFromYearAndStateErr  error
	showsWithNotesRows        []database.ShowsWithShowNotesRow
	showsWithNotesErr         error
	showsWithoutNotesRows     []database.ShowsWithoutNotesRow
	showsWithoutNotesErr      error
	// Songs
	songStatsRow           database.SongStatsRow
	songStatsErr           error
	songsPlayedAtVenueRows []database.AllSongsPlayedAtVenueRow
	songsPlayedAtVenueErr  error
	songsFromSetNameRows   []database.MostCommonSongsBySetNameRow
	songsFromSetNameErr    error
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

func (f *fakeQuerier) GetShowsFromState(ctx context.Context, stateOrCountry string) ([]database.GetShowsFromStateRow, error) {
	return f.showsFromStateRows, f.showsFromStateErr
}

func (f *fakeQuerier) GetShowsFromYear(ctx context.Context, year int32) ([]database.GetShowsFromYearRow, error) {
	return f.showsFromYearRows, f.showsFromYearErr
}

func (f *fakeQuerier) GetShowsFromYearAndState(ctx context.Context, arg database.GetShowsFromYearAndStateParams) ([]database.GetShowsFromYearAndStateRow, error) {
	return f.showsFromYearAndStateRows, f.showsFromYearAndStateErr
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
	return nil, nil
}

func (f *fakeQuerier) SongsPlayedLessThan(ctx context.Context, number int32) ([]database.SongsPlayedLessThanRow, error) {
	return nil, nil
}

func (f *fakeQuerier) UniqueSongsPerCity(ctx context.Context) ([]database.UniqueSongsPerCityRow, error) {
	return nil, nil
}

func (f *fakeQuerier) GetFootnotesFromShowID(ctx context.Context, showID int32) ([]database.GetFootnotesFromShowIDRow, error) {
	return f.footnoteRows, f.footnoteErr
}
