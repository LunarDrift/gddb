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
	xtheme "fyne.io/x/fyne/theme"
	xwidget "fyne.io/x/fyne/widget"
)

func newApp() *App {
	a := &App{}
	a.shows = &showsTab{app: a, results: &showPage{}}
	a.songs = &songsTab{app: a, allSongStats: []songStats{}, visibleSongs: []songStats{}}
	a.venues = &venuesTab{app: a, allVenueStats: []venueStats{}, visibleVenues: []venueStats{}}
	fyneApp := app.New()
	fyneApp.Settings().SetTheme(xtheme.AdwaitaTheme())
	a.win = fyneApp.NewWindow("Deadabase")
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
		container.NewTabItem("Venues", a.venues.buildVenuesTab()),
	)

	return container.NewBorder(nil, nil, nil, nil, tabs)
}

func (s *showsTab) buildShowsTab() fyne.CanvasObject {
	searchAndList := container.NewHSplit(s.buildShowsSearchPanel(), s.buildShowsResultsPanel())
	content := container.NewHSplit(searchAndList, s.buildShowsDetailsPanel())
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

func (v *venuesTab) buildVenuesTab() fyne.CanvasObject {
	v.venueSearch = widget.NewEntry()
	v.venueSearch.PlaceHolder = "Search venue..."
	v.venueSearch.OnChanged = v.venuesFilter

	rows, err := getAllVenueStats()
	if err != nil {
		v.app.showError(err)
		return nil
	}
	v.allVenueStats = rows
	v.visibleVenues = v.allVenueStats
	v.table = v.buildVenueTable()
	content := container.NewBorder(
		v.venueSearch,
		nil,
		nil,
		nil,
		v.table,
	)
	return content
}

func (s *showsTab) buildShowsSearchPanel() fyne.CanvasObject {
	randShowBtn := widget.NewButton("Random Show", s.randomShow)
	todayBtn := widget.NewButton("Today in GD History", s.showsFromTodaysDate)
	acousticShowsBtn := widget.NewButton("Acoustic Shows", s.getAcousticShows)
	songDebutsBtn := widget.NewButton("Song Debuts", s.showsWithSongDebuts)

	// By ID
	s.idEntry = widget.NewEntry()
	idSearchBtn := widget.NewButtonWithIcon("", theme.SearchIcon(), func() { s.searchByID(s.idEntry.Text) })
	idSearchBtn.Importance = widget.LowImportance
	s.idEntry.ActionItem = idSearchBtn
	s.idEntry.SetPlaceHolder("Enter a show ID")
	s.idEntry.OnSubmitted = s.searchByID

	// By date
	var suppressDateFetch bool
	dateSelect := func(_ string) {
		if suppressDateFetch {
			return
		}
		s.searchByDate(s.yearSelect.Selected, monthToNumString(s.monthSelect.Selected), s.daySelect.Selected)
	}
	s.yearSelect = widget.NewSelect(numberStrings(1965, 1995), nil)
	s.yearSelect.PlaceHolder = "Year"
	s.monthSelect = widget.NewSelect(monthNames(), nil)
	s.monthSelect.PlaceHolder = "Month"
	s.daySelect = widget.NewSelect(numberStrings(1, 31), nil)
	s.daySelect.PlaceHolder = "Day"
	s.yearSelect.OnChanged = dateSelect
	s.monthSelect.OnChanged = dateSelect
	s.daySelect.OnChanged = dateSelect

	dateClearBtn := widget.NewButtonWithIcon("Clear Date Selections", theme.CancelIcon(), func() {
		// don't send a new API request if we're just clearing buttons
		suppressDateFetch = true
		s.yearSelect.ClearSelected()
		s.monthSelect.ClearSelected()
		s.daySelect.ClearSelected()
		suppressDateFetch = false
	})
	datePickers := container.NewGridWithColumns(3, s.yearSelect, s.monthSelect, s.daySelect)

	// By song
	s.songSearch = xwidget.NewCompletionEntry(s.app.allSongs)
	songSearchBtn := widget.NewButtonWithIcon("", theme.SearchIcon(), func() { s.searchBySong(s.songSearch.Text) })
	songSearchBtn.Importance = widget.LowImportance
	s.songSearch.ActionItem = songSearchBtn
	s.songSearch.SetPlaceHolder("Song...")
	s.songSearch.OnSubmitted = s.searchBySong
	s.songSearch.OnChanged = func(input string) {
		opts := filterOptions(s.app.allSongs, input)
		s.songSearch.SetOptions(opts)
		if len(opts) == 0 {
			s.songSearch.HideCompletion()
			return
		}
		s.songSearch.ShowCompletion()
	}

	// By venue
	s.venueSearch = xwidget.NewCompletionEntry(s.app.allVenues)
	venueSearchBtn := widget.NewButtonWithIcon("", theme.SearchIcon(), func() { s.searchByVenue(s.venueSearch.Text) })
	venueSearchBtn.Importance = widget.LowImportance
	s.venueSearch.ActionItem = venueSearchBtn
	s.venueSearch.PlaceHolder = "Venue..."
	s.venueSearch.OnSubmitted = s.searchByVenue
	s.venueSearch.OnChanged = func(input string) {
		opts := filterOptions(s.app.allVenues, input)
		s.venueSearch.SetOptions(opts)
		if len(opts) == 0 {
			s.venueSearch.HideCompletion()
			return
		}
		s.venueSearch.ShowCompletion()
	}

	// By city
	s.citySearch = xwidget.NewCompletionEntry(s.app.allCities)
	citySearchBtn := widget.NewButtonWithIcon("", theme.SearchIcon(), func() { s.searchByCity(s.citySearch.Text) })
	citySearchBtn.Importance = widget.LowImportance
	s.citySearch.ActionItem = citySearchBtn
	s.citySearch.SetPlaceHolder("City...")
	s.citySearch.OnSubmitted = s.searchByCity
	s.citySearch.OnChanged = func(input string) {
		opts := filterOptions(s.app.allCities, input)
		s.citySearch.SetOptions(opts)
		if len(opts) == 0 {
			s.citySearch.HideCompletion()
			return
		}
		s.citySearch.ShowCompletion()
	}

	// By location
	s.locationSearch = xwidget.NewCompletionEntry(s.app.allLocations)
	locationSearchBtn := widget.NewButtonWithIcon("", theme.SearchIcon(), func() { s.searchByLocation(s.locationSearch.Text) })
	locationSearchBtn.Importance = widget.LowImportance
	s.locationSearch.ActionItem = locationSearchBtn
	s.locationSearch.PlaceHolder = "Locations (State/Country)"
	s.locationSearch.OnSubmitted = s.searchByLocation
	s.locationSearch.OnChanged = func(input string) {
		opts := filterOptions(s.app.allLocations, input)
		s.locationSearch.SetOptions(opts)
		if len(opts) == 0 {
			s.locationSearch.HideCompletion()
			return
		}
		s.locationSearch.ShowCompletion()
	}

	// By set name
	s.setNameSelect = widget.NewSelect(setLabels(), nil)
	s.setNameSelect.PlaceHolder = "Sets"
	s.setNameSelect.OnChanged = func(set string) { s.searchBySetName() }

	// By notes
	s.notesSelect = widget.NewSelect([]string{"True", "False"}, nil)
	s.notesSelect.PlaceHolder = "Has Notes"
	s.notesSelect.OnChanged = func(hasNotes string) { s.searchByNotes() }

	setAndNotes := container.NewGridWithColumns(2, s.setNameSelect, s.notesSelect)

	s.notesSearch = widget.NewEntry()
	notesClearBtn := widget.NewButtonWithIcon("", theme.SearchIcon(), func() { s.searchShowNotes(s.notesSearch.Text) })
	notesClearBtn.Importance = widget.LowImportance
	s.notesSearch.ActionItem = notesClearBtn
	s.notesSearch.PlaceHolder = "Search for term in notes..."
	s.notesSearch.OnSubmitted = s.searchShowNotes

	quickActions := container.NewGridWithColumns(2, randShowBtn, todayBtn, acousticShowsBtn, songDebutsBtn)
	quickActionsAccordion := widget.NewAccordion(widget.NewAccordionItem("Quick Actions", quickActions))

	idDate := container.NewVBox(s.idEntry, datePickers, dateClearBtn)
	idDateAccordion := widget.NewAccordion(widget.NewAccordionItem("ID & Date", idDate))

	songSearch := container.NewVBox(s.songSearch)
	songAccordion := widget.NewAccordion(widget.NewAccordionItem("Shows From Song Name", songSearch))

	venueLocation := container.NewVBox(s.venueSearch, s.locationSearch, s.citySearch)
	venueLocationAccordion := widget.NewAccordion(widget.NewAccordionItem("Venue & Location", venueLocation))

	otherSearches := container.NewVBox(s.notesSearch, setAndNotes)
	otherSearchesAccordion := widget.NewAccordion(widget.NewAccordionItem("Other Searches", otherSearches))

	accordions := []*widget.Accordion{quickActionsAccordion, idDateAccordion, songAccordion, venueLocationAccordion, otherSearchesAccordion}
	for _, a := range accordions {
		a.OpenAll()
	}

	return container.NewVBox(
		quickActionsAccordion,
		layout.NewSpacer(),
		widget.NewSeparator(),
		idDateAccordion,
		layout.NewSpacer(),
		widget.NewSeparator(),
		songAccordion,
		layout.NewSpacer(),
		widget.NewSeparator(),
		venueLocationAccordion,
		layout.NewSpacer(),
		widget.NewSeparator(),
		otherSearchesAccordion,
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

	s.dateLabel = widget.NewRichText()
	s.dateLabel.Wrapping = fyne.TextWrapWord
	s.locationLabel = widget.NewRichText()
	s.locationLabel.Wrapping = fyne.TextWrapWord
	s.setsLabel = widget.NewRichText()
	s.setsLabel.Wrapping = fyne.TextWrapWord
	s.footnotesLabel = widget.NewRichText()
	s.footnotesLabel.Wrapping = fyne.TextWrapWord
	s.setlist = container.NewVScroll(s.setsLabel)

	footnotes := container.NewVScroll(s.footnotesLabel)
	footnotes.SetMinSize(fyne.NewSize(0, 100))
	s.details = container.NewBorder(
		container.NewVBox(s.dateLabel, s.locationLabel),
		footnotes,
		nil, nil,
		s.setlist,
	)

	return container.NewStack(s.placeholder, s.details)
}

func (v *venuesTab) buildVenueTable() *widget.Table {
	headers := []string{"Location", "Venue", "Shows Played"}

	cell := func(r venueStats, col int) string {
		switch col {
		case 0:
			return fmt.Sprintf("%s, %s", r.City, r.Location)
		case 1:
			return r.Venue
		case 2:
			return strconv.FormatInt(r.ShowsPlayed, 10)
		}
		return ""
	}

	table := widget.NewTableWithHeaders(
		func() (int, int) { return len(v.visibleVenues), len(headers) },
		func() fyne.CanvasObject { return widget.NewLabel("") },
		func(id widget.TableCellID, o fyne.CanvasObject) {
			o.(*widget.Label).SetText(cell(v.visibleVenues[id.Row], id.Col))
		},
	)

	table.CreateHeader = func() fyne.CanvasObject { return widget.NewLabel("header") }
	table.UpdateHeader = func(id widget.TableCellID, o fyne.CanvasObject) {
		l := o.(*widget.Label)
		if id.Col >= 0 && id.Row == -1 {
			l.SetText(headers[id.Col])
		} else {
			l.SetText(strconv.Itoa(id.Row + 1))
		}
	}
	table.SetColumnWidth(0, 300)
	table.SetColumnWidth(1, 700)

	return table
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
			l.SetText(strconv.Itoa(id.Row + 1))
		}
	}

	table.SetColumnWidth(0, 370)
	for c := 1; c <= 3; c++ {
		table.SetColumnWidth(c, 100)
	}
	for c := 4; c < len(headers); c++ {
		table.SetColumnWidth(c, 80)
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

func filterOptions(all []string, input string) []string {
	if input == "" {
		return all
	}
	needle := strings.ToLower(strings.TrimSpace(input))
	var filtered []string
	for _, item := range all {
		if strings.Contains(strings.ToLower(item), needle) {
			filtered = append(filtered, item)
		}
	}
	return filtered
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

func (v *venuesTab) venuesFilter(input string) {
	needle := strings.ToLower(strings.TrimSpace(input))
	filtered := make([]venueStats, 0, len(v.allVenueStats))
	for _, venue := range v.allVenueStats {
		if needle == "" || strings.Contains(strings.ToLower(venue.Venue), needle) {
			filtered = append(filtered, venue)
		}
	}
	v.setVisible(filtered)
}
