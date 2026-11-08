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
	allShowIDs              []int32
	allShowIDsErr           error
	allVenues               []string
	allVenuesErr            error
	allLocations            []string
	allLocationsErr         error
	allCities               []string
	allCitiesErr            error
	searchShowsFiltersRows  []database.SearchShowsWithFiltersRow
	searchShowsFiltersErr   error
	showFromIDRows          []database.GetShowDetailsFromIDRow
	showFromIDErr           error
	showFromDateRows        []database.GetShowDetailsFromDateRow
	showFromDateErr         error
	showsFromTodaysDateRows []database.GetShowsFromTodaysDateRow
	showsFromTodaysDateErr  error
	showsBetweenDatesRows   []database.GetShowsBetweenDatesRow
	showsBetweenDatesErr    error
	listVenueStats          []database.ListVenueStatsRow
	listVenueStatsErr       error
	allShowsRows            []database.GetAllShowsRow
	allShowsErr             error
	getAcousticShowsRows    []database.GetAcousticShowsRow
	getAcousticShowsErr     error
	getSongDebutsRows       []database.GetSongDebutsRow
	getSongDebutsErr        error
	// Songs
	songStatsRow            database.GetSongStatsRow
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
	listSongStatsRow        []database.ListSongStatsRow
	listSongStatsErr        error
	// Other
	validLocationRows []string
	validLocationErr  error
	footnoteRows      []database.GetFootnotesFromShowIDRow
	footnoteErr       error
}

func (f *fakeQuerier) SearchShowsWithFilters(ctx context.Context, arg database.SearchShowsWithFiltersParams) ([]database.SearchShowsWithFiltersRow, error) {
	return f.searchShowsFiltersRows, f.searchShowsFiltersErr
}

func (f *fakeQuerier) GetSongDebuts(ctx context.Context, arg database.GetSongDebutsParams) ([]database.GetSongDebutsRow, error) {
	return f.getSongDebutsRows, f.getSongDebutsErr
}

func (f *fakeQuerier) GetAcousticShows(ctx context.Context, arg database.GetAcousticShowsParams) ([]database.GetAcousticShowsRow, error) {
	return f.getAcousticShowsRows, f.getAcousticShowsErr
}

func (f *fakeQuerier) GetAllShows(ctx context.Context, arg database.GetAllShowsParams) ([]database.GetAllShowsRow, error) {
	return f.allShowsRows, f.allShowsErr
}

func (f *fakeQuerier) ListVenueStats(ctx context.Context) ([]database.ListVenueStatsRow, error) {
	return f.listVenueStats, f.listVenueStatsErr
}

func (f *fakeQuerier) GetAllVenues(ctx context.Context) ([]string, error) {
	return f.allVenues, f.allVenuesErr
}

func (f *fakeQuerier) GetAllLocations(ctx context.Context) ([]string, error) {
	return f.allLocations, f.allLocationsErr
}

func (f *fakeQuerier) GetAllShowIDs(ctx context.Context) ([]int32, error) {
	return f.allShowIDs, f.allShowIDsErr
}

func (f *fakeQuerier) GetAllCities(ctx context.Context) ([]string, error) {
	return f.allCities, f.allCitiesErr
}

func (f *fakeQuerier) GetShowDetailsFromID(ctx context.Context, showID int32) ([]database.GetShowDetailsFromIDRow, error) {
	return f.showFromIDRows, f.showFromIDErr
}

func (f *fakeQuerier) GetShowDetailsFromDate(ctx context.Context, showDate time.Time) ([]database.GetShowDetailsFromDateRow, error) {
	return f.showFromDateRows, f.showFromDateErr
}

func (f *fakeQuerier) GetShowsFromTodaysDate(ctx context.Context, arg database.GetShowsFromTodaysDateParams) ([]database.GetShowsFromTodaysDateRow, error) {
	return f.showsFromTodaysDateRows, f.showsFromTodaysDateErr
}

func (f *fakeQuerier) GetShowsBetweenDates(ctx context.Context, arg database.GetShowsBetweenDatesParams) ([]database.GetShowsBetweenDatesRow, error) {
	return f.showsBetweenDatesRows, f.showsBetweenDatesErr
}

func (f *fakeQuerier) GetSongStats(ctx context.Context, songName sql.NullString) (database.GetSongStatsRow, error) {
	return f.songStatsRow, f.songStatsErr
}

func (f *fakeQuerier) AllSongsPlayedAtVenue(ctx context.Context, arg database.AllSongsPlayedAtVenueParams) ([]database.AllSongsPlayedAtVenueRow, error) {
	return f.songsPlayedAtVenueRows, f.songsPlayedAtVenueErr
}

func (f *fakeQuerier) MostCommonSongsBySetName(ctx context.Context, arg database.MostCommonSongsBySetNameParams) ([]database.MostCommonSongsBySetNameRow, error) {
	return f.songsFromSetNameRows, f.songsFromSetNameErr
}

func (f *fakeQuerier) MostPlayedSongs(ctx context.Context, arg database.MostPlayedSongsParams) ([]database.MostPlayedSongsRow, error) {
	return f.songsMostPlayedRows, f.songsMostplayedErr
}

func (f *fakeQuerier) SongsPlayedLessThan(ctx context.Context, arg database.SongsPlayedLessThanParams) ([]database.SongsPlayedLessThanRow, error) {
	return f.songsPlayedLessThanRows, f.songsPlayedLessThanErr
}

func (f *fakeQuerier) UniqueSongsPerCity(ctx context.Context, arg database.UniqueSongsPerCityParams) ([]database.UniqueSongsPerCityRow, error) {
	return f.songsUniquePerCityRows, f.songsUniquePerCityErr
}

func (f *fakeQuerier) GetFootnotesFromShowID(ctx context.Context, showID int32) ([]database.GetFootnotesFromShowIDRow, error) {
	return f.footnoteRows, f.footnoteErr
}

func (f *fakeQuerier) GetValidLocations(ctx context.Context) ([]string, error) {
	return f.validLocationRows, f.validLocationErr
}

func (f *fakeQuerier) ListSongStats(ctx context.Context) ([]database.ListSongStatsRow, error) {
	return f.listSongStatsRow, f.listSongStatsErr
}
