package main

import (
	"encoding/json"
	"fmt"
	"image/color"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/widget"
	"github.com/LunarDrift/deadabase/internal"
)

type searchID struct {
	widget.Entry
	enterBtn *widget.Button
	OnEnter  func()
}

func newSearchPanel() *searchID {
	searchPanel := &searchID{}
	searchPanel.ExtendBaseWidget(searchPanel)
	return searchPanel
}

func (e *searchID) KeyDown(key *fyne.KeyEvent) {
	if key.Name == fyne.KeyReturn || key.Name == fyne.KeyEnter {
		if e.OnEnter != nil {
			e.OnEnter()
		}
	} else {
		e.Entry.KeyDown(key)
	}
}

type singleShowResult struct {
	showIDLabel    *canvas.Text
	dateLabel      *canvas.Text
	locationLabel  *canvas.Text
	notesLabel     *canvas.Text
	setsLabel      *widget.Label
	footnotesLabel *widget.Label
}

type App struct {
	fyneApp          fyne.App
	fyneWindow       fyne.Window
	searchPanel      *searchID
	showResultSingle singleShowResult
}

func (app *App) GetShowFromID() {
	showID := app.searchPanel.Text
	id, err := strconv.Atoi(showID)
	if err != nil {
		fmt.Print(err)
	}

	url := fmt.Sprintf("http://localhost:8080/shows/%d", id)

	res, err := http.Get(url)
	if err != nil {
		fmt.Print(err)
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
	app.showResultSingle.showIDLabel.Text = fmt.Sprintf("Show ID: %v", show.ShowID)
	app.showResultSingle.dateLabel.Text = fmt.Sprintf("Date: %v", show.Date)
	app.showResultSingle.locationLabel.Text = fmt.Sprintf("Location: %v, %v", show.Venue, show.Location)
	app.showResultSingle.notesLabel.Text = fmt.Sprintf("Notes: %v", show.Notes)
	app.showResultSingle.setsLabel.SetText(setsText.String())
	app.showResultSingle.footnotesLabel.SetText(footnotesText.String())

	// Refresh so Fyne redraws with new text
	app.showResultSingle.showIDLabel.Refresh()
	app.showResultSingle.dateLabel.Refresh()
	app.showResultSingle.locationLabel.Refresh()
	app.showResultSingle.notesLabel.Refresh()
	app.showResultSingle.setsLabel.Refresh()
	app.showResultSingle.footnotesLabel.Refresh()

	app.searchPanel.SetText("")
}

func (app *App) setupResultCanvas() {
	app.searchPanel.SetPlaceHolder("Enter a ShowID...")
	app.showResultSingle.showIDLabel = canvas.NewText("", color.White)
	app.showResultSingle.dateLabel = canvas.NewText("", color.White)
	app.showResultSingle.locationLabel = canvas.NewText("", color.White)
	app.showResultSingle.notesLabel = canvas.NewText("", color.White)
	app.showResultSingle.setsLabel = widget.NewLabel("")
	app.showResultSingle.setsLabel.Wrapping = fyne.TextWrapWord
	app.showResultSingle.footnotesLabel = widget.NewLabel("")
	app.showResultSingle.footnotesLabel.Wrapping = fyne.TextWrapWord
}

func newApp() *App {
	a := app.New()
	w := a.NewWindow("Deadabase")
	w.Resize(fyne.NewSize(1280, 720))
	return &App{
		a,
		w,
		newSearchPanel(),
		singleShowResult{},
	}
}
