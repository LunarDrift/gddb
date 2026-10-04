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
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/widget"
	"github.com/LunarDrift/deadabase/internal"
)

type searchPanel struct {
	widget.Entry
	enterBtn *widget.Button
	OnEnter  func()
}

func newSearchPanel() *searchPanel {
	s := &searchPanel{}
	s.ExtendBaseWidget(s)

	return s
}

func (s *searchPanel) KeyDown(key *fyne.KeyEvent) {
	if key.Name == fyne.KeyReturn || key.Name == fyne.KeyEnter {
		if s.OnEnter != nil {
			s.OnEnter()
		}
	} else {
		s.Entry.KeyDown(key)
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
	searchBox        *searchPanel
	resultsList      *widget.List
	detailsContainer *fyne.Container
	showResultSingle singleShowResult
}

func (a *App) setupResultCanvas() {
	a.showResultSingle.showIDLabel = canvas.NewText("", color.White)
	a.showResultSingle.dateLabel = canvas.NewText("", color.White)
	a.showResultSingle.locationLabel = canvas.NewText("", color.White)
	a.showResultSingle.notesLabel = canvas.NewText("", color.White)
	a.showResultSingle.setsLabel = widget.NewLabel("")
	a.showResultSingle.setsLabel.Wrapping = fyne.TextWrapWord
	a.showResultSingle.footnotesLabel = widget.NewLabel("")
	a.showResultSingle.footnotesLabel.Wrapping = fyne.TextWrapWord
}

func (a *App) SearchBySongName() {
	// TODO: Figure out how to get this to work without segfaulting
	// I think it's got something to do with the length function when there are no results yet? I'm getting this error:
	// panic: runtime error: invalid memory address or nil pointer dereference
	var data internal.Paginated[internal.ShowMeta]

	songName := a.searchBox.Text
	resp, err := http.Get(fmt.Sprintf("http://localhost:8080/shows?song=%s", songName))
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close() //nolint:errcheck
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		panic(err)
	}

	a.resultsList = widget.NewList(
		func() int {
			return len(data.Results)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("template")
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			obj.(*widget.Label).SetText(data.Results[id].Date)
		},
	)
}

func (a *App) GetShowFromID() {
	showID := a.searchBox.Text
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
	a.showResultSingle.showIDLabel.Text = fmt.Sprintf("Show ID: %v", show.ShowID)
	a.showResultSingle.dateLabel.Text = fmt.Sprintf("Date: %v", show.Date)
	a.showResultSingle.locationLabel.Text = fmt.Sprintf("Location: %v, %v", show.Venue, show.Location)
	a.showResultSingle.notesLabel.Text = fmt.Sprintf("Notes: %v", show.Notes)
	a.showResultSingle.setsLabel.SetText(setsText.String())
	a.showResultSingle.footnotesLabel.SetText(footnotesText.String())

	// Refresh so Fyne redraws with new text
	a.showResultSingle.showIDLabel.Refresh()
	a.showResultSingle.dateLabel.Refresh()
	a.showResultSingle.locationLabel.Refresh()
	a.showResultSingle.notesLabel.Refresh()
	a.showResultSingle.setsLabel.Refresh()
	a.showResultSingle.footnotesLabel.Refresh()

	a.searchBox.SetText("")
}
