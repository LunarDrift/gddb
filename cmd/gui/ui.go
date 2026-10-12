package main

import (
	"fmt"
	"strconv"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func newApp() *App {
	a := &App{
		results: &showPage{}, // never nil, so list callbacks are always safe
	}
	a.win = app.New().NewWindow("Deadabase")
	a.win.Resize(fyne.NewSize(1200, 875))
	a.win.SetContent(a.buildUI())

	// Set initial state of things that start hidden/disabled
	a.showPlaceholder()
	a.updateButtons()
	return a
}

func (a *App) buildUI() fyne.CanvasObject {
	searchAndList := container.NewHSplit(a.buildSearchPanel(), a.buildResultsPanel())
	searchAndList.SetOffset(0.4)
	content := container.NewHSplit(searchAndList, a.buildDetailsPanel())
	content.SetOffset(0.4)

	player := widget.NewLabel("Music Player")
	player.Alignment = fyne.TextAlignCenter
	return container.NewBorder(nil, player, nil, nil, content)
}

func (a *App) buildSearchPanel() fyne.CanvasObject {
	// By ID
	a.idEntry = widget.NewEntry()
	a.idEntry.SetPlaceHolder("Enter a show ID")
	a.idEntry.OnSubmitted = a.searchByID
	idBtn := widget.NewButtonWithIcon("Search ID", theme.SearchIcon(), func() { a.searchByID(a.idEntry.Text) })

	// By date
	a.yearSelect = widget.NewSelect(numberStrings(1965, 1995), nil)
	a.yearSelect.PlaceHolder = "Year"
	a.monthSelect = widget.NewSelect(monthNames(), nil)
	a.monthSelect.PlaceHolder = "Month"
	a.daySelect = widget.NewSelect(numberStrings(1, 31), nil)
	a.daySelect.PlaceHolder = "Day"
	dateBtn := widget.NewButtonWithIcon("Search date", theme.SearchIcon(), a.searchByDate)
	datePickers := container.NewHBox(a.yearSelect, a.monthSelect, a.daySelect)

	// By song
	a.songEntry = widget.NewEntry()
	a.songEntry.SetPlaceHolder("Enter a song name...")
	a.songEntry.OnSubmitted = a.searchBySong
	songBtn := widget.NewButtonWithIcon("Search song", theme.SearchIcon(), func() { a.searchBySong(a.songEntry.Text) })

	// By venue
	a.venueEntry = widget.NewEntry()
	a.venueEntry.SetPlaceHolder("Enter a venue...")
	a.venueEntry.OnSubmitted = a.searchByVenue
	venueBtn := widget.NewButtonWithIcon("Search venue", theme.SearchIcon(), func() { a.searchByVenue(a.venueEntry.Text) })

	// By city
	a.cityEntry = widget.NewEntry()
	a.cityEntry.SetPlaceHolder("Enter a city...")
	a.cityEntry.OnSubmitted = a.searchByCity
	cityBtn := widget.NewButtonWithIcon("Search city", theme.SearchIcon(), func() { a.searchByCity(a.cityEntry.Text) })

	// By location
	a.locationEntry = widget.NewEntry()
	a.locationEntry.SetPlaceHolder("(State/Country) e.g. NY / Spain")
	a.locationEntry.OnSubmitted = a.searchByLocation
	locBtn := widget.NewButtonWithIcon("Search location", theme.SearchIcon(), func() { a.searchByLocation(a.locationEntry.Text) })

	// By set name
	a.setNameSelect = widget.NewSelect(setLabels(), nil)
	a.setNameSelect.PlaceHolder = "Sets"
	setNameBtn := widget.NewButtonWithIcon("Search by set", theme.SearchIcon(), a.searchBySetName)

	return container.NewVBox(a.idEntry, idBtn,
		datePickers, dateBtn,
		a.songEntry, songBtn,
		a.venueEntry, venueBtn,
		a.cityEntry, cityBtn,
		a.locationEntry, locBtn,
		a.setNameSelect, setNameBtn,
	)
}

func (a *App) buildResultsPanel() fyne.CanvasObject {
	a.resultsList = widget.NewList(
		func() int { return len(a.results.Results) },
		func() fyne.CanvasObject { return widget.NewLabel("template") },
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			obj.(*widget.Label).SetText(a.results.Results[id].Date)
		},
	)
	a.resultsList.OnSelected = a.onResultSelected

	a.prevBtn = widget.NewButtonWithIcon("Prev", theme.NavigateBackIcon(), a.loadPrev)
	a.nextBtn = widget.NewButtonWithIcon("Next", theme.NavigateNextIcon(), a.loadNext)

	return container.NewBorder(
		nil,
		container.NewHBox(a.prevBtn, layout.NewSpacer(), a.nextBtn),
		nil, nil,
		a.resultsList,
	)
}

func (a *App) buildDetailsPanel() fyne.CanvasObject {
	a.placeholder = container.NewCenter(widget.NewLabel("Select an item to see details"))

	a.dateLabel = widget.NewLabel("")
	a.locationLabel = widget.NewLabel("")
	a.setsLabel = widget.NewLabel("")
	a.setsLabel.Wrapping = fyne.TextWrapWord
	a.footnotesLabel = widget.NewLabel("")
	a.footnotesLabel.Wrapping = fyne.TextWrapWord

	footnotes := container.NewVScroll(a.footnotesLabel)
	footnotes.SetMinSize(fyne.NewSize(0, 100))
	a.details = container.NewBorder(
		container.NewVBox(a.dateLabel, a.locationLabel),
		footnotes,
		nil, nil,
		container.NewVScroll(a.setsLabel),
	)

	// Placeholder buttons.
	// TODO: implement these
	addBtn := widget.NewButton("Add to playlist", func() { fmt.Println("Add to playlist") })
	streamBtn := widget.NewButton("Stream from archive.org", func() { fmt.Println("Streaming") })

	return container.NewBorder(
		nil,
		container.NewHBox(addBtn, layout.NewSpacer(), streamBtn),
		nil, nil,
		container.NewStack(a.placeholder, a.details),
	)
}

func numberStrings(from, to int) []string {
	out := make([]string, 0, to-from+1)
	for i := from; i <= to; i++ {
		out = append(out, strconv.Itoa(i))
	}
	return out
}

func monthNames() []string {
	out := make([]string, 0, 12)
	for m := time.January; m <= time.December; m++ {
		out = append(out, m.String())
	}
	return out
}

func setLabels() []string {
	out := make([]string, len(validSetNames))
	for i, name := range validSetNames {
		out[i] = formatSetName(name)
	}
	return out
}
