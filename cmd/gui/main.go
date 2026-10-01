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
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/LunarDrift/deadabase/internal"
)

type enterEntry struct {
	widget.Entry
	OnEnter func()
}

func newEnterEntry() *enterEntry {
	entry := &enterEntry{}
	entry.ExtendBaseWidget(entry)
	return entry
}

func (e *enterEntry) KeyDown(key *fyne.KeyEvent) {
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
	fyneApp    fyne.App
	fyneWindow fyne.Window
	search     *enterEntry
	showResult singleShowResult
}

func (app *App) GetShowFromID() {
	showID := app.search.Text
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
	app.showResult.showIDLabel.Text = fmt.Sprintf("Show ID: %v", show.ShowID)
	app.showResult.dateLabel.Text = fmt.Sprintf("Date: %v", show.Date)
	app.showResult.locationLabel.Text = fmt.Sprintf("Location: %v, %v", show.Venue, show.Location)
	app.showResult.notesLabel.Text = fmt.Sprintf("Notes: %v", show.Notes)
	app.showResult.setsLabel.SetText(setsText.String())
	app.showResult.footnotesLabel.SetText(footnotesText.String())

	// Refresh so Fyne redraws with new text
	app.showResult.showIDLabel.Refresh()
	app.showResult.dateLabel.Refresh()
	app.showResult.locationLabel.Refresh()
	app.showResult.notesLabel.Refresh()
	app.showResult.setsLabel.Refresh()
	app.showResult.footnotesLabel.Refresh()

	app.search.SetText("")
}

func newApp() *App {
	a := app.New()
	w := a.NewWindow("Deadabase")
	w.Resize(fyne.NewSize(1280, 720))
	return &App{
		a,
		w,
		newEnterEntry(),
		singleShowResult{},
	}
}

func main() {
	// a := app.New()
	// w := a.NewWindow("Deadabase")
	// w.Resize(fyne.NewSize(1280, 720))
	app := newApp()

	// inputID := newEnterEntry()
	app.search.SetPlaceHolder("Enter a ShowID...")

	app.showResult.showIDLabel = canvas.NewText("", color.White)
	app.showResult.dateLabel = canvas.NewText("", color.White)
	app.showResult.locationLabel = canvas.NewText("", color.White)
	app.showResult.notesLabel = canvas.NewText("", color.White)
	app.showResult.setsLabel = widget.NewLabel("")
	app.showResult.setsLabel.Wrapping = fyne.TextWrapWord
	app.showResult.footnotesLabel = widget.NewLabel("")
	app.showResult.footnotesLabel.Wrapping = fyne.TextWrapWord

	results := container.NewVBox(
		app.showResult.showIDLabel,
		app.showResult.dateLabel,
		app.showResult.locationLabel,
		app.showResult.notesLabel,
		app.showResult.setsLabel,
		app.showResult.footnotesLabel,
	)

	enterBtn := widget.NewButton("Search", app.GetShowFromID)
	app.search.OnEnter = app.GetShowFromID

	showSearch := container.NewVBox(app.search, enterBtn)
	content := container.NewHSplit(showSearch, results)
	content.SetOffset(0.25)

	app.fyneWindow.SetContent(container.NewVBox(content))
	app.fyneWindow.ShowAndRun()
}
