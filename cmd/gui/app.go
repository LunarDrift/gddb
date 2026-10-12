package main

import (
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/LunarDrift/deadabase/internal"
)

type App struct {
	win fyne.Window

	// Search panel
	idEntry       *widget.Entry
	yearSelect    *widget.Select
	monthSelect   *widget.Select
	daySelect     *widget.Select
	songEntry     *widget.Entry
	venueEntry    *widget.Entry
	cityEntry     *widget.Entry
	locationEntry *widget.Entry

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

func (a *App) searchByID(id string) {
	id = strings.TrimSpace(id)
	if id == "" {
		return
	}
	async(a, func() ([]internal.ShowResponse, error) {
		show, err := fetchShowByID(id)
		return []internal.ShowResponse{show}, err
	}, a.showShowDetails)
}

func (a *App) searchByDate() {
	date, err := buildDate(
		a.yearSelect.Selected,
		a.monthSelect.SelectedIndex()+1,
		a.daySelect.SelectedIndex()+1,
	)
	if err != nil {
		a.showError(err)
		return
	}
	a.loadShowsByDate(date)
}

func (a *App) searchBySong(song string) {
	async(a, func() (*showPage, error) { return searchSongs(song) }, a.setResults)
}

func (a *App) loadShowsByDate(date string) {
	async(a, func() ([]internal.ShowResponse, error) { return fetchShowsByDate(date) }, a.showShowDetails)
}

func (a *App) searchByVenue(venue string) {
	async(a, func() (*showPage, error) { return searchVenues(venue) }, a.setResults)
}

func (a *App) searchByCity(city string) {
	async(a, func() (*showPage, error) { return searchCities(city) }, a.setResults)
}

func (a *App) searchByLocation(location string) {
	async(a, func() (*showPage, error) { return searchLocations(location) }, a.setResults)
}

// ---------------- Results list ----------------

func (a *App) onResultSelected(id widget.ListItemID) {
	if id >= len(a.results.Results) {
		return
	}
	a.loadShowsByDate(a.results.Results[id].Date)
}

// setResults is the ONE place new results enter the app. Everything that
// produces a page (song search, next, previous) funnelsthrough here, so the
// list, selection, and details pane and buttons can never get out of sync.
func (a *App) setResults(p *showPage) {
	a.results = p
	a.resultsList.UnselectAll()
	a.resultsList.Refresh()
	a.resultsList.ScrollToTop()
	a.showPlaceholder()
	a.updateButtons()
}

func (a *App) loadPrev() {
	if a.results.Previous != nil {
		a.loadPage(*a.results.Previous)
	}
}

func (a *App) loadNext() {
	if a.results.Next != nil {
		a.loadPage(*a.results.Next)
	}
}

func (a *App) loadPage(u string) {
	// Disable while loading so a double-click can't fire two requests
	a.prevBtn.Disable()
	a.nextBtn.Disable()
	go func() {
		p, err := fetchPage(baseURL + u)
		fyne.Do(func() {
			if err != nil {
				a.showError(err)
				a.updateButtons()
				return
			}
			a.setResults(p)
		})
	}()
}

func (a *App) updateButtons() {
	setEnabled(a.prevBtn, a.results.Previous != nil)
	setEnabled(a.nextBtn, a.results.Next != nil)
}

func setEnabled(b *widget.Button, on bool) {
	if on {
		b.Enable()
	} else {
		b.Disable()
	}
}

// ---------------- Details panel ----------------

func (a *App) showShowDetails(data []internal.ShowResponse) {
	if len(data) == 0 {
		return
	}
	t := formatShows(data)
	a.dateLabel.SetText(t.date)
	a.locationLabel.SetText(t.location)
	a.setsLabel.SetText(t.sets)
	a.footnotesLabel.SetText(t.footnotes)
	a.placeholder.Hide()
	a.details.Show()
}

func (a *App) showPlaceholder() {
	a.details.Hide()
	a.placeholder.Show()
}
