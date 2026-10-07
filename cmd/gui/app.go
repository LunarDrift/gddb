package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"image/color"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/LunarDrift/deadabase/internal"
)

type showIDSearch struct {
	searchID *widget.Entry
	enterBtn *widget.Button
}

type showDateSearch struct {
	dayChoice   *widget.Select
	monthChoice *widget.Select
	yearChoice  *widget.Select
	enterBtn    *widget.Button
	day         int
	month       int
	year        int
}

type showsSongSearch struct {
	songEntry *widget.Entry
}

type searchPanel struct {
	textSearch showIDSearch
	dateSearch showDateSearch
	songSearch showsSongSearch
}

type resultsListPanel struct {
	results *widget.List
	nextBtn *widget.Button
	prevBtn *widget.Button
}

type singleShowResult struct {
	detailLabel    *widget.Label
	showIDLabel    *canvas.Text
	dateLabel      *canvas.Text
	locationLabel  *canvas.Text
	notesLabel     *canvas.Text
	setsLabel      *widget.Label
	footnotesLabel *widget.Label
}

type showDetailsPanel struct {
	container         *fyne.Container
	placeholderLayout *fyne.Container
	showDetails       singleShowResult
}

type App struct {
	fyneApp     fyne.App
	fyneWindow  fyne.Window
	search      *searchPanel
	resultsList resultsListPanel
	details     showDetailsPanel
}

func (a *App) setupResultCanvas() {
	a.details.showDetails.detailLabel = widget.NewLabel("Select an item to see details")
	a.details.showDetails.showIDLabel = canvas.NewText("", color.White)
	a.details.showDetails.dateLabel = canvas.NewText("", color.White)
	a.details.showDetails.locationLabel = canvas.NewText("", color.White)
	a.details.showDetails.notesLabel = canvas.NewText("", color.White)
	a.details.showDetails.setsLabel = widget.NewLabel("")
	a.details.showDetails.setsLabel.Wrapping = fyne.TextWrapWord
	a.details.showDetails.footnotesLabel = widget.NewLabel("")
	a.details.showDetails.footnotesLabel.Wrapping = fyne.TextWrapWord
}

func (a *App) dayChoiceFn(s string) {
	a.search.dateSearch.day = a.search.dateSearch.dayChoice.SelectedIndex() + 1
	// log.Println("Day set to:", a.search.dateSearch.day)
}

func (a *App) monthChoiceFn(s string) {
	a.search.dateSearch.month = a.search.dateSearch.monthChoice.SelectedIndex() + 1
	// log.Println("Month set to:", a.search.dateSearch.month)
}

func (a *App) yearChoiceFn(s string) {
	year, err := strconv.Atoi(s)
	if err != nil {
		dialog.ShowError(err, a.fyneWindow)
		return
	}
	a.search.dateSearch.year = year
	// log.Println("Year set to:", a.search.dateSearch.year)
}

func (a *App) searchByDate() {
	date := time.Date(a.search.dateSearch.year, time.Month(a.search.dateSearch.month), a.search.dateSearch.day, 0, 0, 0, 0, time.UTC)
	dateStr := date.Format(time.DateOnly)
	url := fmt.Sprintf("http://localhost:8080/shows/%s", dateStr)
	resp, err := http.Get(url)
	if err != nil {
		dialog.ShowError(err, a.fyneWindow)
		return
	}
	defer resp.Body.Close() //nolint:errcheck

	data := []internal.ShowResponse{}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		dialog.ShowError(err, a.fyneWindow)
		return
	}

	if len(data) == 0 {
		dialog.ShowError(errors.New("no show found on that date"), a.fyneWindow)
		return
	}
	// Update text of existing labels
	var setsText strings.Builder
	for _, set := range data[0].Sets {
		fmt.Fprintf(&setsText, "%s: %s\n\n", set.SetName, strings.Join(set.Songs, ", "))
	}
	var footnotesText strings.Builder
	keys := make([]string, 0, len(data[0].Footnotes))
	for k := range data[0].Footnotes {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		fmt.Fprintf(&footnotesText, "[%s] %s\n", k, data[0].Footnotes[k])
	}
	a.details.showDetails.showIDLabel.Text = fmt.Sprintf("Show ID: %v", data[0].ShowID)
	a.details.showDetails.dateLabel.Text = fmt.Sprintf("Date: %v", data[0].Date)
	a.details.showDetails.locationLabel.Text = fmt.Sprintf("Location: %v, %v", data[0].Venue, data[0].Location)
	a.details.showDetails.notesLabel.Text = fmt.Sprintf("Notes: %v", data[0].Notes)
	a.details.showDetails.setsLabel.SetText(setsText.String())
	a.details.showDetails.footnotesLabel.SetText(footnotesText.String())

	// Refresh so Fyne redraws with new text
	a.details.showDetails.showIDLabel.Refresh()
	a.details.showDetails.dateLabel.Refresh()
	a.details.showDetails.locationLabel.Refresh()
	a.details.showDetails.notesLabel.Refresh()
	a.details.showDetails.setsLabel.Refresh()
	a.details.showDetails.footnotesLabel.Refresh()
}

func (a *App) getShowFromID(showID string) {
	id, err := strconv.Atoi(showID)
	if err != nil {
		dialog.ShowError(err, a.fyneWindow)
		return
	}

	url := fmt.Sprintf("http://localhost:8080/shows/%d", id)

	res, err := http.Get(url)
	if err != nil {
		dialog.ShowError(err, a.fyneWindow)
		return
	}
	defer res.Body.Close() //nolint:errcheck

	show := internal.ShowResponse{}
	if err := json.NewDecoder(res.Body).Decode(&show); err != nil {
		fmt.Print(err)
	}

	// Update text of existing labels
	var setsText strings.Builder
	for _, set := range show.Sets {
		fmt.Fprintf(&setsText, "%s: %s\n\n", set.SetName, strings.Join(set.Songs, ", "))
	}
	var footnotesText strings.Builder
	keys := make([]string, 0, len(show.Footnotes))
	for k := range show.Footnotes {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		fmt.Fprintf(&footnotesText, "[%s] %s\n", k, show.Footnotes[k])
	}
	a.details.showDetails.showIDLabel.Text = fmt.Sprintf("Show ID: %v", show.ShowID)
	a.details.showDetails.dateLabel.Text = fmt.Sprintf("Date: %v", show.Date)
	a.details.showDetails.locationLabel.Text = fmt.Sprintf("Location: %v, %v", show.Venue, show.Location)
	a.details.showDetails.notesLabel.Text = fmt.Sprintf("Notes: %v", show.Notes)
	a.details.showDetails.setsLabel.SetText(setsText.String())
	a.details.showDetails.footnotesLabel.SetText(footnotesText.String())

	// Refresh so Fyne redraws with new text
	a.details.showDetails.showIDLabel.Refresh()
	a.details.showDetails.dateLabel.Refresh()
	a.details.showDetails.locationLabel.Refresh()
	a.details.showDetails.notesLabel.Refresh()
	a.details.showDetails.setsLabel.Refresh()
	a.details.showDetails.footnotesLabel.Refresh()

	a.search.textSearch.searchID.SetText("")
}

// fetchShowFromIDBtn: widget.Button can't use a func(string) - this was my initial idea to bypass that issue
func (a *App) fetchShowFromIDBtn() {
	showID := a.search.textSearch.searchID.Text
	a.getShowFromID(showID)
}

func searchBySongName(song string) (internal.Paginated[internal.ShowMeta], error) {
	u := "http://localhost:8080/shows?song=" + url.QueryEscape(song)
	resp, err := http.Get(u)
	if err != nil {
		return internal.Paginated[internal.ShowMeta]{}, err
	}
	defer resp.Body.Close() //nolint:errcheck

	if resp.StatusCode != http.StatusOK {
		return internal.Paginated[internal.ShowMeta]{}, fmt.Errorf("api returned %s", resp.Status)
	}
	var page internal.Paginated[internal.ShowMeta]
	if err := json.NewDecoder(resp.Body).Decode(&page); err != nil {
		return internal.Paginated[internal.ShowMeta]{}, err
	}
	return page, nil
}

func getShowFromDate(date string) (*internal.ShowResponse, error) {
	resp, err := http.Get("http://localhost:8080/shows/" + url.PathEscape(date))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close() //nolint:errcheck

	var shows []internal.ShowResponse
	if err := json.NewDecoder(resp.Body).Decode(&shows); err != nil {
		return nil, err
	}
	if len(shows) == 0 {
		return nil, fmt.Errorf("no show found for %s", date)
	}
	return &shows[0], nil
}
