package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne/v2"
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
	enterBtn  *widget.Button
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
	showIDLabel    *widget.Label
	dateLabel      *widget.Label
	locationLabel  *widget.Label
	notesLabel     *widget.Label
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

func (a *App) setupSearchPanel(days, months, years []string) {
	a.search.textSearch.searchID.SetPlaceHolder("Enter a show ID")
	a.search.textSearch.searchID.OnSubmitted = a.getShowFromID
	a.search.textSearch.enterBtn = widget.NewButton("Search by ID", a.fetchShowFromIDBtn)
	a.search.dateSearch = showDateSearch{
		dayChoice:   widget.NewSelect(days, a.dayChoiceFn),
		monthChoice: widget.NewSelect(months, a.monthChoiceFn),
		yearChoice:  widget.NewSelect(years, a.yearChoiceFn),
	}
	a.search.dateSearch.enterBtn = widget.NewButton("Search by date", a.searchByDate)
	a.search.dateSearch.dayChoice.PlaceHolder = "Day"
	a.search.dateSearch.monthChoice.PlaceHolder = "Month"
	a.search.dateSearch.yearChoice.PlaceHolder = "Year"
	a.search.songSearch.enterBtn = widget.NewButton("Search by song", a.searchBySongNameBtn)
}

func (a *App) setupResultCanvas() {
	a.details.showDetails.detailLabel = widget.NewLabel("Select an item to see details")
	a.details.showDetails.showIDLabel = widget.NewLabel("")
	a.details.showDetails.dateLabel = widget.NewLabel("")
	a.details.showDetails.locationLabel = widget.NewLabel("")
	a.details.showDetails.notesLabel = widget.NewLabel("")
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

func (a *App) updateSingleShowDetails(data internal.ShowResponse) {
	var setsText, footnotesText strings.Builder
	for _, set := range data.Sets {
		fmt.Fprintf(&setsText, "%s:\n%s\n\n", set.SetName, strings.Join(set.Songs, "\n"))
	}
	keys := make([]string, 0, len(data.Footnotes))
	for k := range data.Footnotes {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		fmt.Fprintf(&footnotesText, "[%s] %s\n", k, data.Footnotes[k])
	}
	a.details.showDetails.showIDLabel.Text = fmt.Sprintf("Show ID: %v", data.ShowID)
	a.details.showDetails.dateLabel.Text = fmt.Sprintf("Date: %v", data.Date)
	a.details.showDetails.locationLabel.Text = fmt.Sprintf("Location: %v, %v", data.Venue, data.Location)
	a.details.showDetails.notesLabel.Text = fmt.Sprintf("Notes: %v", data.Notes)
	a.details.showDetails.setsLabel.SetText(setsText.String())
	a.details.showDetails.footnotesLabel.SetText(footnotesText.String())
	a.details.placeholderLayout.Hide()

	// Refresh so Fyne redraws with new text
	a.details.container.Refresh()
}

func (a *App) updateMultiShowDetails(data []internal.ShowResponse) {
	if len(data) == 0 {
		return
	}

	// Set shared top-level info once
	a.details.showDetails.dateLabel.SetText(fmt.Sprintf("Date: %v", data[0].Date))
	a.details.showDetails.locationLabel.SetText(fmt.Sprintf("Location: %v, %v", data[0].Venue, data[0].Location))

	var footnotesText, masterShowText strings.Builder
	for i := range data {
		fmt.Fprintf(&masterShowText, "Show ID: %v\n", data[i].ShowID)
		if data[i].Notes != "" {
			fmt.Fprintf(&masterShowText, "Notes: %v\n", data[i].Notes)
		}
		fmt.Fprintln(&masterShowText, "") // Padding

		for _, set := range data[i].Sets {
			fmt.Fprintf(&masterShowText, "%s: \n%s\n\n", set.SetName, strings.Join(set.Songs, "\n"))
		}

		if len(data[i].Footnotes) > 0 {
			fmt.Fprintf(&footnotesText, "━━━ Show %v Footnotes ━━━\n", data[i].ShowID)

			keys := make([]string, 0, len(data[i].Footnotes))
			for k := range data[i].Footnotes {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			for _, k := range keys {
				fmt.Fprintf(&footnotesText, "[%s] %s\n", k, data[i].Footnotes[k])
			}
			fmt.Fprintln(&footnotesText, "")
		}

		// divider between shows
		if i < len(data)-1 {
			fmt.Fprintln(&masterShowText, strings.Repeat("━", 30)+"\n")
		}

	}
	a.details.showDetails.showIDLabel.SetText("")
	a.details.showDetails.notesLabel.SetText("")

	a.details.showDetails.setsLabel.SetText(strings.TrimSpace(masterShowText.String()))
	a.details.showDetails.footnotesLabel.SetText(strings.TrimSpace(footnotesText.String()))
	a.details.placeholderLayout.Hide()
	a.details.container.Show()
	// Refresh so Fyne redraws with new text
	a.details.container.Refresh()
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
		dialog.ShowError(errors.New("no show found on that date"), a.fyneWindow)
		return
	}

	if len(data) == 0 {
		dialog.ShowError(errors.New("no show found on that date"), a.fyneWindow)
		return
	}
	// Update text of existing labels
	a.updateMultiShowDetails(data)
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

	a.updateSingleShowDetails(show)
}

// fetchShowFromIDBtn: widget.Button can't use a func(string) - this was my initial idea to bypass that issue
func (a *App) fetchShowFromIDBtn() {
	showID := a.search.textSearch.searchID.Text
	a.getShowFromID(showID)
}

func (a *App) searchBySongNameBtn() {
	// NOTE: this doesn't work...
	song := a.search.songSearch.songEntry.Text
	shows, err := searchBySongName(song)
	if err != nil {
		dialog.ShowError(err, a.fyneWindow)
	}
	a.resultsList.results = widget.NewList(
		func() int { return len(shows.Results) },
		func() fyne.CanvasObject { return widget.NewLabel("template") },
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			obj.(*widget.Label).SetText(shows.Results[id].Date)
		},
	)
	a.resultsList.results.Refresh()
}

func searchBySongName(song string) (internal.Paginated[internal.ShowMeta], error) {
	u := "http://localhost:8080/shows?song=" + url.QueryEscape(song)
	resp, err := http.Get(u)
	if err != nil {
		return internal.Paginated[internal.ShowMeta]{}, err
	}
	defer resp.Body.Close() //nolint:errcheck

	if resp.StatusCode != http.StatusOK {
		return internal.Paginated[internal.ShowMeta]{}, fmt.Errorf("api returned %d", resp.StatusCode)
	}
	var page internal.Paginated[internal.ShowMeta]
	if err := json.NewDecoder(resp.Body).Decode(&page); err != nil {
		return internal.Paginated[internal.ShowMeta]{}, err
	}
	return page, nil
}

func getShowFromDate(date string) ([]internal.ShowResponse, error) {
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
	return shows, nil
}
