package main

import (
	"errors"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	xwidget "fyne.io/x/fyne/widget"
	"github.com/LunarDrift/deadabase/internal"
)

type App struct {
	win    fyne.Window
	shows  *showsTab
	songs  *songsTab
	venues *venuesTab

	allCities    []string
	allVenues    []string
	allLocations []string
	allSongs     []string
}

type showsTab struct {
	app *App
	// Search panel
	idEntry        *widget.Entry
	yearSelect     *widget.Select
	monthSelect    *widget.Select
	daySelect      *widget.Select
	songSearch     *xwidget.CompletionEntry
	venueSearch    *xwidget.CompletionEntry
	citySearch     *xwidget.CompletionEntry
	locationSearch *xwidget.CompletionEntry
	setNameSelect  *widget.Select
	notesSelect    *widget.Select
	notesSearch    *widget.Entry

	// Results panel
	results     *showPage
	resultsList *widget.List
	prevBtn     *widget.Button
	nextBtn     *widget.Button

	// Details panel
	placeholder    *fyne.Container
	details        *fyne.Container
	dateLabel      *widget.Label
	locationLabel  *widget.Label
	setsLabel      *widget.Label
	footnotesLabel *widget.Label
}

type songsTab struct {
	app          *App
	allSongStats []songStats
	visibleSongs []songStats
	songSearch   *widget.Entry
	table        *widget.Table
}

type venuesTab struct {
	app           *App
	allVenueStats []venueStats
	visibleVenues []venueStats
	venueSearch   *widget.Entry
	table         *widget.Table
}

// async runs work off the UI thread, then runs done back on it. Fyne's rule:
// slow things (HTTP) in a goroutine, anything touching widgets inside
// fyne.Do. On error it shows a dialog and skips done.
func async[T any](a *App, work func() (T, error), done func(T)) {
	go func() {
		v, err := work()
		fyne.Do(func() {
			if err != nil {
				a.showError(err)
				return
			}
			done(v)
		})
	}()
}

func (a *App) showError(err error) {
	dialog.ShowError(err, a.win)
}

func (a *App) loadLookups() {
	venues, err := getAllVenues()
	if err != nil {
		fyne.Do(func() { a.showError(err) })
		return
	}
	cities, err := getAllCities()
	if err != nil {
		fyne.Do(func() { a.showError(err) })
		return
	}
	locations, err := getAllLocations()
	if err != nil {
		fyne.Do(func() { a.showError(err) })
		return
	}
	songs, err := getAllSongNames()
	if err != nil {
		fyne.Do(func() { a.showError(err) })
		return
	}

	fyne.Do(func() {
		a.allVenues = venues
		a.allCities = cities
		a.allLocations = locations
		a.allSongs = songs
		a.shows.locationSearch.SetOptions(locations)
	})
}

// ---------------- Searches ----------------

func (s *showsTab) searchByID(id string) {
	id = strings.TrimSpace(id)
	if id == "" {
		s.app.showError(errors.New("enter a show ID"))
		return
	}
	async(s.app, func() ([]internal.ShowResponse, error) {
		show, err := getShowByID(id)
		return []internal.ShowResponse{show}, err
	}, s.showShowDetails)
}

func (s *showsTab) searchBySetName() {
	i := s.setNameSelect.SelectedIndex()
	if i < 0 {
		s.app.showError(errors.New("choose a set"))
		return
	}
	setName := validSetNames[i]
	async(s.app, func() (*showPage, error) { return searchSetNames(setName) }, s.setResults)
}

func (s *showsTab) searchByNotes() {
	options := []string{"true", "false"}
	i := s.notesSelect.SelectedIndex()
	if i < 0 {
		s.app.showError(errors.New("choose an option"))
		return
	}
	option := options[i]
	async(s.app, func() (*showPage, error) { return searchNotes(option) }, s.setResults)
}

func (s *showsTab) searchBySong(song string) {
	async(s.app, func() (*showPage, error) { return searchSongs(song) }, s.setResults)
}

func (s *showsTab) searchShowsByDate(date string) {
	async(s.app, func() ([]internal.ShowResponse, error) { return getShowsByDate(date) }, s.showShowDetails)
}

func (s *showsTab) showsFromTodaysDate() {
	async(s.app, func() (*showPage, error) { return getShowsFromTodaysDate() }, s.setResults)
}

func (s *showsTab) searchByVenue(venue string) {
	async(s.app, func() (*showPage, error) { return searchVenues(venue) }, s.setResults)
	// s.venueSelect.SetText("")
}

func (s *showsTab) searchByCity(city string) {
	async(s.app, func() (*showPage, error) { return searchCities(city) }, s.setResults)
}

func (s *showsTab) searchByLocation(location string) {
	async(s.app, func() (*showPage, error) { return searchLocations(location) }, s.setResults)
}

func (s *showsTab) searchByDate(year, month, day string) {
	async(s.app, func() (*showPage, error) { return getShowsFromDate(year, month, day) }, s.setResults)
}

func (s *showsTab) randomShow() {
	async(s.app, func() ([]internal.ShowResponse, error) { return getRandomShow() }, s.showShowDetails)
}

func (s *showsTab) searchShowNotes(query string) {
	if query == "" {
		s.app.showError(errors.New("enter a search term"))
		return
	}
	async(s.app, func() (*showPage, error) { return getShowsFromNotesQuery(query) }, s.setResults)
}

func (s *showsTab) getAcousticShows() {
	async(s.app, func() (*showPage, error) { return getAcousticShows() }, s.setResults)
}

func (s *showsTab) showsWithSongDebuts() {
	async(s.app, func() (*showPage, error) { return getShowsWithSongDebuts() }, s.setResults)
}

// ---------------- Results list ----------------

func (s *showsTab) onResultSelected(id widget.ListItemID) {
	if id >= len(s.results.Results) {
		return
	}
	s.searchShowsByDate(s.results.Results[id].Date)
}

// setResults is the ONE place new results enter the app. Everything that
// produces a page (song search, next, previous) funnelsthrough here, so the
// list, selection, and details pane and buttons can never get out of sync.
func (s *showsTab) setResults(p *showPage) {
	s.results = p
	s.resultsList.UnselectAll()
	s.resultsList.Refresh()
	s.resultsList.ScrollToTop()
	s.showPlaceholder()
	s.updateButtons()
}

func (s *showsTab) loadPrev() {
	if s.results.Previous != nil {
		s.loadPage(*s.results.Previous)
	}
}

func (s *showsTab) loadNext() {
	if s.results.Next != nil {
		s.loadPage(*s.results.Next)
	}
}

func (s *showsTab) loadPage(u string) {
	// Disable while loading so a double-click can't fire two requests
	s.prevBtn.Disable()
	s.nextBtn.Disable()
	go func() {
		p, err := fetchPage[showPage](absoluteURL(u))
		fyne.Do(func() {
			if err != nil {
				s.app.showError(err)
				s.updateButtons()
				return
			}
			s.setResults(p)
		})
	}()
}

func (s *showsTab) updateButtons() {
	setEnabled(s.prevBtn, s.results.Previous != nil)
	setEnabled(s.nextBtn, s.results.Next != nil)
}

func (s *songsTab) setVisible(rows []songStats) {
	s.visibleSongs = rows
	s.table.Refresh()
	s.table.ScrollToTop()
}

func (v *venuesTab) setVisible(rows []venueStats) {
	v.visibleVenues = rows
	v.table.Refresh()
	v.table.ScrollToTop()
}

func setEnabled(b *widget.Button, on bool) {
	if on {
		b.Enable()
	} else {
		b.Disable()
	}
}

// ---------------- Details panel ----------------

func (s *showsTab) showShowDetails(data []internal.ShowResponse) {
	if len(data) == 0 {
		return
	}
	t := formatShows(data)
	s.dateLabel.SetText(t.date)
	s.locationLabel.SetText(t.location)
	s.setsLabel.SetText(t.sets)
	s.footnotesLabel.SetText(t.footnotes)
	s.placeholder.Hide()
	s.details.Show()
}

func (s *showsTab) showPlaceholder() {
	s.details.Hide()
	s.placeholder.Show()
}
