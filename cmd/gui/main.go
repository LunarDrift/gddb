package main

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/LunarDrift/deadabase/internal"
)

func main() {
	var days, years []string
	for i := 1; i <= 31; i++ {
		days = append(days, strconv.Itoa(i))
	}
	for i := 1965; i <= 1995; i++ {
		years = append(years, strconv.Itoa(i))
	}
	months := []string{"January", "February", "March", "April", "May", "June", "July", "August", "September", "October", "November", "December"}

	a := &App{
		fyneApp: app.New(),
		search:  &searchPanel{textSearch: newTextSearch()},
	}
	a.search.dateSearch = dateChoices{
		dayChoice:   widget.NewSelect(days, a.dayChoiceFn),
		monthChoice: widget.NewSelect(months, a.monthChoiceFn),
		yearChoice:  widget.NewSelect(years, a.yearChoiceFn),
	}
	a.search.dateSearch.enterBtn = widget.NewButton("Search by date", a.searchByDate)
	a.search.dateSearch.dayChoice.PlaceHolder = "Day"
	a.search.dateSearch.monthChoice.PlaceHolder = "Month"
	a.search.dateSearch.yearChoice.PlaceHolder = "Year"
	a.fyneWindow = a.fyneApp.NewWindow("Deadabase")
	a.fyneWindow.Resize(fyne.NewSize(1280, 720))

	a.setupResultCanvas()
	a.search.textSearch.enterBtn = widget.NewButton("Search by ID", a.GetShowFromID)
	a.search.textSearch.OnEnter = a.GetShowFromID

	a.detailsContainer = container.NewVBox(
		a.showResultSingle.showIDLabel,
		a.showResultSingle.dateLabel,
		a.showResultSingle.locationLabel,
		a.showResultSingle.notesLabel,
		a.showResultSingle.setsLabel,
		a.showResultSingle.footnotesLabel,
	)

	var shows internal.Paginated[internal.ShowMeta]

	placeholderLayout := container.NewCenter(a.showResultSingle.detailLabel)
	a.resultsList = widget.NewList(
		func() int { return len(shows.Results) },
		func() fyne.CanvasObject { return widget.NewLabel("template") },
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			obj.(*widget.Label).SetText(shows.Results[id].Date)
		},
	)

	a.resultsList.OnSelected = func(id widget.ListItemID) {
		date := shows.Results[id].Date
		go func() {
			show, err := getShowFromDate(date)
			fyne.Do(func() {
				if err != nil {
					dialog.ShowError(err, a.fyneWindow)
					return
				}
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
				placeholderLayout.Hide()
				a.detailsContainer.Show()
				a.detailsContainer.Refresh()
			})
		}()
	}

	a.search.songSearch = widget.NewEntry()
	a.search.songSearch.SetPlaceHolder("Enter a song name...")
	a.search.songSearch.OnSubmitted = func(song string) {
		go func() {
			results, err := searchBySongName(song)
			fyne.Do(func() {
				if err != nil {
					dialog.ShowError(err, a.fyneWindow)
					return
				}
				shows = results
				a.resultsList.UnselectAll()
				a.resultsList.Refresh() // this is what makes the list redraw

				a.detailsContainer.Hide()
				placeholderLayout.Show()
			})
		}()
	}

	// Placeholder buttons
	// TODO: Implement these
	prevBtn := widget.NewButton("Next", func() {
		fmt.Println("Next...")
	})
	nextBtn := widget.NewButton("Previous", func() {
		fmt.Println("Previous...")
	})
	addToPlaylistBtn := widget.NewButton("Add to Playlist", func() {
		fmt.Println("Add to playlist...")
	})
	streamBtn := widget.NewButton("Stream from archive.org", func() {
		fmt.Println("Streaming...")
	})
	musicPlayer := widget.NewLabel("Music Player")
	resultsPanel := container.NewBorder(
		nil,
		container.NewHBox(prevBtn, layout.NewSpacer(), nextBtn),
		nil, nil,
		a.resultsList,
	)
	detailsPanel := container.NewBorder(
		nil,
		container.NewHBox(addToPlaylistBtn, layout.NewSpacer(), streamBtn),
		nil, nil,
		container.NewVScroll(a.detailsContainer),
	)

	dateSearch := container.NewHBox(a.search.dateSearch.yearChoice, a.search.dateSearch.monthChoice, a.search.dateSearch.dayChoice)
	searchPanel := container.NewVBox(a.search.textSearch, a.search.textSearch.enterBtn, dateSearch, a.search.dateSearch.enterBtn, a.search.songSearch)
	searchAndList := container.NewHSplit(searchPanel, resultsPanel)
	content := container.NewHSplit(searchAndList, detailsPanel)
	content.SetOffset(0.4)
	root := container.NewBorder(nil, musicPlayer, nil, nil, content)

	a.fyneWindow.SetContent(root)
	a.fyneWindow.ShowAndRun()
}
