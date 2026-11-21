package main

import (
	"errors"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
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

	// Quick Actions
	acousticShowsBtnClicked bool
	todayBtnClicked         bool
	songDebutsBtnClicked    bool

	// Search panel
	activeFilters  *showFilters
	idEntry        *widget.Entry
	yearSelect     *xwidget.CompletionEntry
	monthSelect    *xwidget.CompletionEntry
	daySelect      *xwidget.CompletionEntry
	weekdaySelect  *xwidget.CompletionEntry
	songSearch     *xwidget.CompletionEntry
	venueSearch    *xwidget.CompletionEntry
	citySearch     *xwidget.CompletionEntry
	locationSearch *xwidget.CompletionEntry
	setNameSearch  *xwidget.CompletionEntry
	hasNotesSearch *xwidget.CompletionEntry
	notesSearch    *widget.Entry

	// Results panel
	results           *showPage
	resultsList       *widget.List
	resultsListHeader *widget.RichText
	prevBtn           *widget.Button
	nextBtn           *widget.Button

	// Details panel
	placeholder    *fyne.Container
	details        *fyne.Container
	setlist        *container.Scroll
	dateLabel      *widget.RichText
	locationLabel  *widget.RichText
	setsLabel      *widget.RichText
	footnotesLabel *widget.RichText
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

func (s *showsTab) searchShows() {
	async(s.app, func() (*showPage, error) {
		return searchShows(
			s.yearSelect.Text, monthToNumString(s.monthSelect.Text), s.daySelect.Text,
			s.weekdaySelect.Text,
			s.locationSearch.Text,
			s.citySearch.Text,
			s.songSearch.Text,
			s.venueSearch.Text,
			s.notesSearch.Text,
			s.setNameSearch.Text,
			s.hasNotesSearch.Text)
	}, s.setResults)
}

func (s *showsTab) searchShowsByDate(date string) {
	async(s.app, func() ([]internal.ShowResponse, error) { return getShowsByDate(date) }, s.showShowDetails)
}

func (s *showsTab) showsFromTodaysDate() {
	async(s.app, func() (*showPage, error) { return getShowsFromTodaysDate() }, s.setResults)
}

func (s *showsTab) randomShow() {
	async(s.app, func() ([]internal.ShowResponse, error) { return getRandomShow() }, s.showShowDetails)
}

func (s *showsTab) getAcousticShows() {
	async(s.app, func() (*showPage, error) { return getAcousticShows() }, s.setResults)
}

func (s *showsTab) showsWithSongDebuts() {
	async(s.app, func() (*showPage, error) { return getShowsWithSongDebuts() }, s.setResults)
}

// ---------------- Results list ----------------

// keep track of active filters to use in the resultsListHeader
// display what produced the results
type showFilters struct {
	// Quick Actions
	AcousticShows string
	Today         string
	SongDebuts    string
	// Search filters
	Year, Month, Day string
	Weekday          string
	Song, Venue      string
	Location, City   string
	Notes            string
	SetName          string
	HasNotes         string
}

// read widgets once, when Search is clicked/Enter pressed
func (s *showsTab) currentFilters() *showFilters {
	filters := &showFilters{
		Year:     strings.TrimSpace(s.yearSelect.Text),
		Month:    strings.TrimSpace(s.monthSelect.Text),
		Day:      strings.TrimSpace(s.daySelect.Text),
		Weekday:  strings.TrimSpace(s.weekdaySelect.Text),
		Song:     strings.TrimSpace(s.songSearch.Text),
		Venue:    strings.TrimSpace(s.venueSearch.Text),
		Location: strings.TrimSpace(s.locationSearch.Text),
		City:     strings.TrimSpace(s.citySearch.Text),
		Notes:    strings.TrimSpace(s.notesSearch.Text),
		SetName:  strings.TrimSpace(s.setNameSearch.Text),
		HasNotes: strings.TrimSpace(s.hasNotesSearch.Text),
	}
	if s.acousticShowsBtnClicked {
		filters.AcousticShows = "Acoustic Shows"
	}
	if s.todayBtnClicked {
		filters.Today = "Today in GD History"
	}
	if s.songDebutsBtnClicked {
		filters.SongDebuts = "Song Debuts"
	}
	return filters
}

// describe skips empty fields
func (f showFilters) Describe() string {
	var parts []string
	add := func(label, v string) {
		if v != "" {
			parts = append(parts, label+v)
		}
	}

	// Quick action buttons
	add("", f.AcousticShows)
	add("", f.Today)
	add("", f.SongDebuts)

	date := strings.Join(nonEmpty(f.Year, f.Month, f.Day), " ")
	add("", date)
	add("**Weekday:** ", f.Weekday)
	add("**Song:** ", f.Song)
	add("**Location:** ", f.Location)
	add("**City:** ", f.City)
	add("**Venue:** ", f.Venue)
	add("**Notes:** ", f.Notes)
	add("**Set:** ", f.SetName)
	add("**Has Notes:** ", f.HasNotes)
	return strings.Join(parts, " **·** ")
}

func nonEmpty(year, month, day string) []string {
	out := []string{}
	if year != "" {
		out = append(out, year)
	}
	if month != "" {
		out = append(out, month)
	}
	if day != "" {
		out = append(out, day)
	}
	return out
}

func (s *showsTab) onResultSelected(id widget.ListItemID) {
	if id >= len(s.results.Results) {
		return
	}
	s.searchShowsByDate(s.results.Results[id].Date)
	s.setlist.ScrollToTop()
}

// setResults is the ONE place new results enter the app. Everything that
// produces a page (song search, next, previous) funnelsthrough here, so the
// list, selection, and details pane and buttons can never get out of sync.
func (s *showsTab) setResults(p *showPage) {
	s.results = p
	s.activeFilters = s.currentFilters()
	s.resultsListHeader.ParseMarkdown("**" + strconv.FormatInt(p.Count, 10) + " shows found**\n\n" + s.activeFilters.Describe())
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
		p, err := fetchPage[showPage](u)
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
	s.dateLabel.ParseMarkdown(t.header)
	s.locationLabel.ParseMarkdown(t.showNotes)
	s.setsLabel.ParseMarkdown(t.setlist)
	s.footnotesLabel.ParseMarkdown(t.songNotes)
	s.placeholder.Hide()
	s.details.Show()
}

func (s *showsTab) showPlaceholder() {
	s.details.Hide()
	s.placeholder.Show()
}
