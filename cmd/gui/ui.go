package main

import (
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

var (
	venues    = getAllVenues()
	locations = getAllLocations()
	cities    = getAllCities()
)

func newApp() *App {
	a := &App{}
	a.shows = &showsTab{app: a, results: &showPage{}}
	a.songs = &songsTab{app: a, allSongStats: []songStats{}, visibleSongs: []songStats{}}
	a.win = app.New().NewWindow("Deadabase")
	a.win.Resize(fyne.NewSize(1200, 875))
	a.win.SetContent(a.buildUI())

	// Set initial state of things that start hidden/disabled
	a.shows.showPlaceholder()
	a.shows.updateButtons()
	return a
}

func (a *App) buildUI() fyne.CanvasObject {
	tabs := container.NewAppTabs(
		container.NewTabItem("Shows", a.shows.buildShowsTab()),
		container.NewTabItem("Songs", a.songs.buildSongsTab()),
	)

	return container.NewBorder(nil, nil, nil, nil, tabs)
}

func (s *showsTab) buildShowsTab() fyne.CanvasObject {
	searchAndList := container.NewHSplit(s.buildShowsSearchPanel(), s.buildShowsResultsPanel())
	content := container.NewHSplit(searchAndList, s.buildShowsDetailsPanel())
	content.SetOffset(0.5)
	return content
}

func (s *songsTab) buildSongsTab() fyne.CanvasObject {
	s.songSearch = widget.NewEntry()
	s.songSearch.PlaceHolder = "Search song..."
	s.songSearch.OnChanged = s.songsFilter

	rows, err := getAllSongStats()
	if err != nil {
		s.app.showError(err)
		return nil
	}
	s.allSongStats = rows
	s.visibleSongs = s.allSongStats
	s.table = s.buildSongTable()
	content := container.NewBorder(
		s.songSearch,
		nil,
		nil,
		nil,
		s.table,
	)
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
	s.venueSearch = widget.NewSelectEntry(venues)
	s.venueSearch.PlaceHolder = "Start typing to filter venues"
	s.venueSearch.OnSubmitted = s.searchByVenue
	s.venueSearch.OnChanged = s.venueFilter
	venueBtn := widget.NewButtonWithIcon("Search venue", theme.SearchIcon(), func() { s.searchByVenue(s.venueSearch.Text) })

	// By city
	s.citySearch = widget.NewSelectEntry(cities)
	s.citySearch.SetPlaceHolder("Start typing to filter cities")
	s.citySearch.OnSubmitted = s.searchByCity
	s.citySearch.OnChanged = s.cityFilter
	cityBtn := widget.NewButtonWithIcon("Search city", theme.SearchIcon(), func() { s.searchByCity(s.citySearch.Text) })

	// By location
	s.locationSearch = widget.NewSelect(locations, nil)
	s.locationSearch.PlaceHolder = "Locations (State/Country)"
	locBtn := widget.NewButtonWithIcon("Search location", theme.SearchIcon(), func() { s.searchByLocation(s.locationSearch.Selected) })

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

	return container.NewVBox(
		layout.NewSpacer(),
		randShowBtn,
		layout.NewSpacer(),
		s.idEntry, idBtn,
		layout.NewSpacer(),
		datePickers, dateBtn,
		yearBtn,
		layout.NewSpacer(),
		s.songEntry, songBtn,
		layout.NewSpacer(),
		s.venueSearch, venueBtn,
		layout.NewSpacer(),
		s.citySearch, cityBtn,
		layout.NewSpacer(),
		s.locationSearch, locBtn,
		layout.NewSpacer(),
		setAndNotes, setAndNotesBtns,
		layout.NewSpacer(),
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
	s.dateLabel.Selectable = true
	s.locationLabel = widget.NewLabel("")
	s.locationLabel.Selectable = true
	s.setsLabel = widget.NewLabel("")
	s.setsLabel.Wrapping = fyne.TextWrapWord
	s.setsLabel.Selectable = true
	s.footnotesLabel = widget.NewLabel("")
	s.footnotesLabel.Wrapping = fyne.TextWrapWord
	s.footnotesLabel.Selectable = true

	footnotes := container.NewVScroll(s.footnotesLabel)
	footnotes.SetMinSize(fyne.NewSize(0, 100))
	s.details = container.NewBorder(
		container.NewVBox(s.dateLabel, s.locationLabel),
		footnotes,
		nil, nil,
		container.NewVScroll(s.setsLabel),
	)

	return container.NewStack(s.placeholder, s.details)
}

func (s *songsTab) buildSongTable() *widget.Table {
	headers := []string{"Title", "Times Played", "First Played", "Last Played", "Set 1", "Set 2/3", "Encore", "Acoustic", "Electric"}

	cell := func(r songStats, col int) string {
		switch col {
		case 0:
			return r.Title
		case 1:
			return strconv.FormatInt(r.TimesPlayed, 10)
		case 2:
			return r.FirstPlayed.Format(time.DateOnly)
		case 3:
			return r.LastPlayed.Format(time.DateOnly)
		case 4:
			return strconv.FormatInt(r.Set1, 10)
		case 5:
			return strconv.FormatInt(r.Set23, 10)
		case 6:
			return strconv.FormatInt(r.Encore, 10)
		case 7:
			return strconv.FormatInt(r.Acoustic, 10)
		case 8:
			return strconv.FormatInt(r.Electric, 10)
		}
		return ""
	}

	table := widget.NewTableWithHeaders(
		func() (int, int) { return len(s.visibleSongs), len(headers) },
		func() fyne.CanvasObject { return widget.NewLabel("") },
		func(id widget.TableCellID, o fyne.CanvasObject) {
			o.(*widget.Label).SetText(cell(s.visibleSongs[id.Row], id.Col))
		},
	)

	table.CreateHeader = func() fyne.CanvasObject { return widget.NewLabel("header") }
	table.UpdateHeader = func(id widget.TableCellID, o fyne.CanvasObject) {
		l := o.(*widget.Label)
		if id.Col >= 0 && id.Row == -1 {
			l.SetText(headers[id.Col])
		} else {
			l.SetText(strconv.Itoa(id.Row + 1)) // row numbers
		}
	}

	table.SetColumnWidth(0, 370)
	for c := 1; c < len(headers); c++ {
		table.SetColumnWidth(c, 100)
	}
	return table
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

func (s *showsTab) venueFilter(input string) {
	if input == "" {
		s.venueSearch.SetOptions(venues)
		return
	}
	var filtered []string
	for _, venue := range venues {
		if strings.Contains(strings.ToLower(venue), strings.ToLower(input)) {
			filtered = append(filtered, venue)
		}
	}
	s.venueSearch.SetOptions(filtered)
}

func (s *showsTab) cityFilter(input string) {
	if input == "" {
		s.citySearch.SetOptions(cities)
		return
	}
	var filtered []string
	for _, city := range cities {
		if strings.Contains(strings.ToLower(city), strings.ToLower(input)) {
			filtered = append(filtered, city)
		}
	}
	s.citySearch.SetOptions(filtered)
}

func (s *songsTab) songsFilter(input string) {
	needle := strings.ToLower(strings.TrimSpace(input))
	filtered := make([]songStats, 0, len(s.allSongStats))
	for _, song := range s.allSongStats {
		if needle == "" || strings.Contains(strings.ToLower(song.Title), needle) {
			filtered = append(filtered, song)
		}
	}
	s.setVisible(filtered)
}
