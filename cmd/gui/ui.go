package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func newApp() *App {
	a := &App{}
	a.shows = &showsTab{app: a, results: &showPage{}}
	a.songs = &songsTab{app: a, page: &songPage{}}
	a.win = app.New().NewWindow("Deadabase")
	a.win.Resize(fyne.NewSize(1200, 875))
	a.win.SetContent(a.buildUI())

	// Set initial state of things that start hidden/disabled
	a.shows.showPlaceholder()
	a.shows.updateButtons()
	a.songs.updateButtons()
	return a
}

func (a *App) buildUI() fyne.CanvasObject {
	tabs := container.NewAppTabs(
		container.NewTabItem("Shows", a.shows.buildShowsTab()),
		container.NewTabItem("Songs", a.songs.buildSongsTab()),
	)

	a.player = widget.NewLabel("Music Player")
	a.player.Alignment = fyne.TextAlignCenter
	return container.NewBorder(nil, a.player, nil, nil, tabs)
}

func (s *showsTab) buildShowsTab() fyne.CanvasObject {
	searchAndList := container.NewHSplit(s.buildShowsSearchPanel(), s.buildShowsResultsPanel())
	content := container.NewHSplit(searchAndList, s.buildShowsDetailsPanel())
	content.SetOffset(0.5)
	return content
}

func (s *songsTab) buildSongsTab() fyne.CanvasObject {
	content := container.NewHSplit(s.buildSongsSearchPanel(), s.buildSongsResultsPanel())
	return content
}

func (s *showsTab) buildShowsSearchPanel() fyne.CanvasObject {
	randShowBtn := widget.NewButton("Random Show", s.randomShow)

	// By ID
	s.idEntry = widget.NewEntry()
	idClearBtn := widget.NewButtonWithIcon("", theme.CancelIcon(), func() {
		s.idEntry.SetText("")
	})
	idClearBtn.Importance = widget.LowImportance
	s.idEntry.ActionItem = idClearBtn
	s.idEntry.SetPlaceHolder("Enter a show ID")
	s.idEntry.OnSubmitted = s.searchByID
	idBtn := widget.NewButtonWithIcon("Search ID", theme.SearchIcon(), func() { s.searchByID(s.idEntry.Text) })

	// By date
	s.yearSelect = widget.NewSelect(numberStrings(1965, 1995), nil)
	s.yearSelect.PlaceHolder = "Year"
	s.monthSelect = widget.NewSelect(monthNames(), nil)
	s.monthSelect.PlaceHolder = "Month"
	s.daySelect = widget.NewSelect(numberStrings(1, 31), nil)
	s.daySelect.PlaceHolder = "Day"
	dateBtn := widget.NewButtonWithIcon("Search date", theme.SearchIcon(), s.searchByDate)
	datePickers := container.NewGridWithColumns(3, s.yearSelect, s.monthSelect, s.daySelect)

	// By year
	yearBtn := widget.NewButtonWithIcon("Search by year", theme.SearchIcon(), func() { s.searchByYear(s.yearSelect.Selected) })

	// By song
	s.songEntry = widget.NewEntry()
	songClearBtn := widget.NewButtonWithIcon("", theme.CancelIcon(), func() {
		s.songEntry.SetText("")
	})
	songClearBtn.Importance = widget.LowImportance
	s.songEntry.ActionItem = songClearBtn
	s.songEntry.SetPlaceHolder("Enter a song name...")
	s.songEntry.OnSubmitted = s.searchBySong
	songBtn := widget.NewButtonWithIcon("Search song", theme.SearchIcon(), func() { s.searchBySong(s.songEntry.Text) })

	// By venue
	venues := getAllVenues()
	s.venueSelect = widget.NewSelectEntry(venues)
	s.venueSelect.PlaceHolder = "Choose or type a venue"
	s.venueSelect.OnSubmitted = s.searchByVenue
	s.venueSelect.OnChanged = func(input string) {
		if input == "" {
			s.venueSelect.SetOptions(venues)
			return
		}
		var filtered []string
		for _, venue := range venues {
			if strings.Contains(strings.ToLower(venue), strings.ToLower(input)) {
				filtered = append(filtered, venue)
			}
		}
		s.venueSelect.SetOptions(filtered)
	}
	venueBtn := widget.NewButtonWithIcon("Search venue", theme.SearchIcon(), func() { s.searchByVenue(s.venueSelect.Text) })

	// By city
	s.cityEntry = widget.NewEntry()
	cityClearBtn := widget.NewButtonWithIcon("", theme.CancelIcon(), func() {
		s.cityEntry.SetText("")
	})
	cityClearBtn.Importance = widget.LowImportance
	s.cityEntry.ActionItem = cityClearBtn
	s.cityEntry.SetPlaceHolder("Enter a city...")
	s.cityEntry.OnSubmitted = s.searchByCity
	cityBtn := widget.NewButtonWithIcon("Search city", theme.SearchIcon(), func() { s.searchByCity(s.cityEntry.Text) })

	// By location
	s.locationEntry = widget.NewEntry()
	locClearBtn := widget.NewButtonWithIcon("", theme.CancelIcon(), func() {
		s.locationEntry.SetText("")
	})
	locClearBtn.Importance = widget.LowImportance
	s.locationEntry.ActionItem = locClearBtn
	s.locationEntry.SetPlaceHolder("(State/Country) e.g. NY / Spain")
	s.locationEntry.OnSubmitted = s.searchByLocation
	locBtn := widget.NewButtonWithIcon("Search location", theme.SearchIcon(), func() { s.searchByLocation(s.locationEntry.Text) })

	// By set name
	s.setNameSelect = widget.NewSelect(setLabels(), nil)
	s.setNameSelect.PlaceHolder = "Sets"
	setNameBtn := widget.NewButtonWithIcon("Search by set", theme.SearchIcon(), s.searchBySetName)

	// By notes
	s.notesSelect = widget.NewSelect(notesLabels(), nil)
	s.notesSelect.PlaceHolder = "Has Notes"
	notesBtn := widget.NewButtonWithIcon("Search notes", theme.SearchIcon(), s.searchByNotes)

	setAndNotes := container.NewGridWithColumns(2, s.setNameSelect, s.notesSelect)
	setAndNotesBtns := container.NewGridWithColumns(2, setNameBtn, notesBtn)

	return container.NewVBox(randShowBtn,
		layout.NewSpacer(),
		s.idEntry, idBtn,
		layout.NewSpacer(),
		datePickers, dateBtn,
		yearBtn,
		layout.NewSpacer(),
		s.songEntry, songBtn,
		layout.NewSpacer(),
		s.venueSelect, venueBtn,
		layout.NewSpacer(),
		s.cityEntry, cityBtn,
		layout.NewSpacer(),
		s.locationEntry, locBtn,
		layout.NewSpacer(),
		setAndNotes, setAndNotesBtns,
	)
}

func (s *showsTab) buildShowsResultsPanel() fyne.CanvasObject {
	s.resultsList = widget.NewList(
		func() int { return len(s.results.Results) },
		func() fyne.CanvasObject { return widget.NewLabel("template") },
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			obj.(*widget.Label).SetText(s.results.Results[id].Date)
		},
	)
	s.resultsList.OnSelected = s.onResultSelected

	s.prevBtn = widget.NewButtonWithIcon("Prev", theme.NavigateBackIcon(), s.loadPrev)
	s.nextBtn = widget.NewButtonWithIcon("Next", theme.NavigateNextIcon(), s.loadNext)

	return container.NewBorder(
		nil,
		container.NewHBox(s.prevBtn, layout.NewSpacer(), s.nextBtn),
		nil, nil,
		s.resultsList,
	)
}

func (s *showsTab) buildShowsDetailsPanel() fyne.CanvasObject {
	s.placeholder = container.NewCenter(widget.NewLabel("Select an item to see details"))

	s.dateLabel = widget.NewLabel("")
	s.locationLabel = widget.NewLabel("")
	s.setsLabel = widget.NewLabel("")
	s.setsLabel.Wrapping = fyne.TextWrapWord
	s.footnotesLabel = widget.NewLabel("")
	s.footnotesLabel.Wrapping = fyne.TextWrapWord

	footnotes := container.NewVScroll(s.footnotesLabel)
	footnotes.SetMinSize(fyne.NewSize(0, 100))
	s.details = container.NewBorder(
		container.NewVBox(s.dateLabel, s.locationLabel),
		footnotes,
		nil, nil,
		container.NewVScroll(s.setsLabel),
	)

	// Placeholder buttons.
	// TODO: implement these
	addBtn := widget.NewButton("Add to playlist", func() { fmt.Println("Add to playlist") })
	streamBtn := widget.NewButton("Stream from archive.org", func() { fmt.Println("Streaming") })

	return container.NewBorder(
		nil,
		container.NewHBox(addBtn, layout.NewSpacer(), streamBtn),
		nil, nil,
		container.NewStack(s.placeholder, s.details),
	)
}

func (s *songsTab) buildSongsSearchPanel() fyne.CanvasObject {
	s.mostPlayed = widget.NewButtonWithIcon("Most Played Songs", theme.SearchIcon(), s.mostPlayedSongs)
	s.setSelect = widget.NewSelect(setLabels(), nil)
	s.setSelect.PlaceHolder = "Sets"
	mostPlayedBySetBtn := widget.NewButtonWithIcon("Most Played by Set", theme.SearchIcon(), s.mostPlayedBySet)

	return container.NewVBox(s.mostPlayed,
		s.setSelect, mostPlayedBySetBtn,
	)
}

func (s *songsTab) buildSongsResultsPanel() fyne.CanvasObject {
	s.list = widget.NewList(
		func() int { return len(s.page.Results) },
		func() fyne.CanvasObject {
			songLabel := widget.NewLabel("")
			songLabel.Truncation = fyne.TextTruncateEllipsis
			// songLabel.Wrapping = fyne.TextWrapWord
			timesLabel := widget.NewLabel("")
			timesLabel.Alignment = fyne.TextAlignTrailing
			return container.NewBorder(nil, nil, nil, timesLabel, songLabel)
			// return widget.NewLabel("template")
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			borderContainer := obj.(*fyne.Container)
			songLabel := borderContainer.Objects[0].(*widget.Label)
			timesLabel := borderContainer.Objects[1].(*widget.Label)

			songLabel.SetText(s.page.Results[id].Song)
			timesLabel.SetText(fmt.Sprintf("Played %d times", s.page.Results[id].TimesPlayed))
			obj.Refresh()
			// obj.(*widget.Label).SetText(fmt.Sprintf("%s\tPlayed %v times", s.page.Results[id].Song, s.page.Results[id].TimesPlayed))
		},
	)
	s.prev = widget.NewButtonWithIcon("Prev", theme.NavigateBackIcon(), s.loadPrev)
	s.next = widget.NewButtonWithIcon("Next", theme.NavigateNextIcon(), s.loadNext)

	s.list.OnHighlighted = func(id widget.ListItemID) {
		fullSongName := s.page.Results[id].Song
		timesPlayed := s.page.Results[id].TimesPlayed
		s.app.player.SetText(fmt.Sprintf("%s | Played %d times", fullSongName, timesPlayed))
	}
	s.list.OnUnselected = func(id widget.ListItemID) {
		s.app.player.SetText("")
	}

	return container.NewBorder(
		nil,
		container.NewHBox(s.app.songs.prev, layout.NewSpacer(), s.app.songs.next),
		nil, nil,
		s.app.songs.list,
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

func notesLabels() []string {
	return []string{"True", "False"}
}
