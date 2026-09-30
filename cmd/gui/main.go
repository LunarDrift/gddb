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

type ShowMeta struct {
	ShowID   int32  `json:"show_id"`
	Date     string `json:"date"`
	Venue    string `json:"venue"`
	City     string `json:"city"`
	Location string `json:"location"`
	Notes    string `json:"notes"`
}

type ShowResponse struct {
	ShowMeta
	Sets      []SetResponse     `json:"sets"`
	Footnotes map[string]string `json:"footnotes"`
}

type SetResponse struct {
	SetName string   `json:"set_name"`
	Songs   []string `json:"songs"`
}

func main() {
	a := app.New()
	w := a.NewWindow("Deadabase")
	w.Resize(fyne.NewSize(1280, 720))

	inputID := newEnterEntry()
	inputID.SetPlaceHolder("Enter a ShowID...")

	showIDLabel := canvas.NewText("", color.White)
	dateLabel := canvas.NewText("", color.White)
	locationLabel := canvas.NewText("", color.White)
	notesLabel := canvas.NewText("", color.White)
	setsLabel := widget.NewLabel("")
	setsLabel.Wrapping = fyne.TextWrapWord
	footnotesLabel := widget.NewLabel("")
	footnotesLabel.Wrapping = fyne.TextWrapWord

	results := container.NewVBox(
		showIDLabel,
		dateLabel,
		locationLabel,
		notesLabel,
		setsLabel,
		footnotesLabel,
	)

	GetShowFromID := func() {
		showID := inputID.Text
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

		show := ShowResponse{}
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
		showIDLabel.Text = fmt.Sprintf("Show ID: %v", show.ShowID)
		dateLabel.Text = fmt.Sprintf("Date: %v", show.Date)
		locationLabel.Text = fmt.Sprintf("Location: %v, %v", show.Venue, show.Location)
		notesLabel.Text = fmt.Sprintf("Notes: %v", show.Notes)
		setsLabel.SetText(setsText.String())
		footnotesLabel.SetText(footnotesText.String())

		// Refresh so Fyne redraws with new text
		showIDLabel.Refresh()
		dateLabel.Refresh()
		locationLabel.Refresh()
		notesLabel.Refresh()
		setsLabel.Refresh()
		footnotesLabel.Refresh()

		inputID.SetText("")
	}

	enterBtn := widget.NewButton("Enter", GetShowFromID)
	inputID.OnEnter = GetShowFromID

	showSearch := container.NewVBox(inputID, enterBtn)
	content := container.NewHSplit(showSearch, results)
	content.SetOffset(0.25)
	// content := container.NewVBox(input, enterBtn, results)

	w.SetContent(container.NewVBox(content))
	w.ShowAndRun()
}
