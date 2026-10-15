package main

import (
	"errors"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/LunarDrift/deadabase/internal"
)

type App struct {
	win    fyne.Window
	player *widget.Label
	shows  *showsTab
	songs  *songsTab
}

type showsTab struct {
	app *App
	// Search panel
	idEntry       *widget.Entry
	yearSelect    *widget.Select
	monthSelect   *widget.Select
	daySelect     *widget.Select
	songEntry     *widget.Entry
	venueEntry    *widget.Entry
	cityEntry     *widget.Entry
	locationEntry *widget.Entry
	setNameSelect *widget.Select
	notesSelect   *widget.Select

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
	app  *App
	page *songPage
	// Search panel
	mostPlayed *widget.Button
	// Results list
	list       *widget.List
	prev, next *widget.Button
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

// ---------------- Searches ----------------

func (s *showsTab) searchByID(id string) {
	id = strings.TrimSpace(id)
	if id == "" {
		return
	}
	async(s.app, func() ([]internal.ShowResponse, error) {
		show, err := fetchShowByID(id)
		return []internal.ShowResponse{show}, err
	}, s.showShowDetails)
}

func (s *showsTab) searchByDate() {
	date, err := buildDate(
		s.yearSelect.Selected,
		s.monthSelect.SelectedIndex()+1,
		s.daySelect.SelectedIndex()+1,
	)
	if err != nil {
		s.app.showError(err)
		return
	}
	s.loadShowsByDate(date)
}

func (s *showsTab) searchBySetName() {
	i := s.setNameSelect.SelectedIndex()
	if i < 0 {
		s.app.showError(errors.New("pick a set"))
		return
	}
	setName := validSetNames[i]
	async(s.app, func() (*showPage, error) { return searchSetNames(setName) }, s.setResults)
}

func (s *showsTab) searchByNotes() {
	options := []string{"true", "false"}
	i := s.notesSelect.SelectedIndex()
	if i < 0 {
		s.app.showError(errors.New("pick an option"))
		return
	}
	option := options[i]
	async(s.app, func() (*showPage, error) { return searchNotes(option) }, s.setResults)
}

func (s *showsTab) searchBySong(song string) {
	async(s.app, func() (*showPage, error) { return searchSongs(song) }, s.setResults)
}

func (s *showsTab) loadShowsByDate(date string) {
	async(s.app, func() ([]internal.ShowResponse, error) { return fetchShowsByDate(date) }, s.showShowDetails)
}

func (s *showsTab) searchByVenue(venue string) {
	async(s.app, func() (*showPage, error) { return searchVenues(venue) }, s.setResults)
}

func (s *showsTab) searchByCity(city string) {
	async(s.app, func() (*showPage, error) { return searchCities(city) }, s.setResults)
}

func (s *showsTab) searchByLocation(location string) {
	async(s.app, func() (*showPage, error) { return searchLocations(location) }, s.setResults)
}

func (s *songsTab) mostPlayedSongs() {
	async(s.app, func() (*songPage, error) { return getMostPlayedSongs() }, s.setPage)
}

// ---------------- Results list ----------------

func (s *showsTab) onResultSelected(id widget.ListItemID) {
	if id >= len(s.results.Results) {
		return
	}
	s.loadShowsByDate(s.results.Results[id].Date)
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

func (s *songsTab) setPage(p *songPage) {
	s.page = p
	s.list.UnselectAll()
	s.list.Refresh()
	s.list.ScrollToTop()
	s.updateButtons()
}

func (s *songsTab) loadPrev() {
	if s.page.Previous != nil {
		s.loadPage(*s.page.Previous)
	}
}

func (s *songsTab) loadNext() {
	if s.page.Next != nil {
		s.loadPage(*s.page.Next)
	}
}

func (s *songsTab) loadPage(u string) {
	s.prev.Disable()
	s.next.Disable()
	go func() {
		p, err := fetchPage[songPage](baseURL + u)
		fyne.Do(func() {
			if err != nil {
				s.app.showError(err)
				s.updateButtons()
				return
			}
			s.setPage(p)
		})
	}()
}

func (s *songsTab) updateButtons() {
	setEnabled(s.prev, s.page.Previous != nil)
	setEnabled(s.next, s.page.Next != nil)
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
		p, err := fetchPage[showPage](baseURL + u)
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
