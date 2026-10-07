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
	// TODO: Different number of days depending on the month
	for i := 1; i <= 31; i++ {
		days = append(days, strconv.Itoa(i))
	}
	for i := 1965; i <= 1995; i++ {
		years = append(years, strconv.Itoa(i))
	}
	months := []string{"January", "February", "March", "April", "May", "June", "July", "August", "September", "October", "November", "December"}

	a := &App{
		fyneApp: app.New(),
		search: &searchPanel{
			textSearch: showIDSearch{
				searchID: widget.NewEntry(),
			},
		},
	}
	a.fyneWindow = a.fyneApp.NewWindow("Deadabase")
	a.fyneWindow.Resize(fyne.NewSize(1280, 720))

	a.search.textSearch.searchID.SetPlaceHolder("Enter a show ID")
	a.search.textSearch.searchID.OnSubmitted = a.getShowFromID
	a.search.dateSearch = showDateSearch{
		dayChoice:   widget.NewSelect(days, a.dayChoiceFn),
		monthChoice: widget.NewSelect(months, a.monthChoiceFn),
		yearChoice:  widget.NewSelect(years, a.yearChoiceFn),
	}
	a.search.dateSearch.enterBtn = widget.NewButton("Search by date", a.searchByDate)
	a.search.dateSearch.dayChoice.PlaceHolder = "Day"
	a.search.dateSearch.monthChoice.PlaceHolder = "Month"
	a.search.dateSearch.yearChoice.PlaceHolder = "Year"

	a.setupResultCanvas()
	a.search.textSearch.enterBtn = widget.NewButton("Search by ID", a.fetchShowFromIDBtn)

	a.details.container = container.NewVBox(
		a.details.showDetails.showIDLabel,
		a.details.showDetails.dateLabel,
		a.details.showDetails.locationLabel,
		a.details.showDetails.notesLabel,
		a.details.showDetails.setsLabel,
		a.details.showDetails.footnotesLabel,
	)

	var shows internal.Paginated[internal.ShowMeta]

	a.details.placeholderLayout = container.NewCenter(a.details.showDetails.detailLabel)
	a.resultsList.results = widget.NewList(
		func() int { return len(shows.Results) },
		func() fyne.CanvasObject { return widget.NewLabel("template") },
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			obj.(*widget.Label).SetText(shows.Results[id].Date)
		},
	)

	a.resultsList.results.OnSelected = func(id widget.ListItemID) {
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

				a.details.showDetails.showIDLabel.Text = fmt.Sprintf("Show ID: %v", show.ShowID)
				a.details.showDetails.dateLabel.Text = fmt.Sprintf("Date: %v", show.Date)
				a.details.showDetails.locationLabel.Text = fmt.Sprintf("Location: %v, %v", show.Venue, show.Location)
				a.details.showDetails.notesLabel.Text = fmt.Sprintf("Notes: %v", show.Notes)
				a.details.showDetails.setsLabel.SetText(setsText.String())
				a.details.showDetails.footnotesLabel.SetText(footnotesText.String())
				a.details.placeholderLayout.Hide()
				a.details.container.Show()
				a.details.container.Refresh()
			})
		}()
	}

	a.search.songSearch.songEntry = widget.NewEntry()
	a.search.songSearch.songEntry.SetPlaceHolder("Enter a song name...")
	a.search.songSearch.songEntry.OnSubmitted = func(song string) {
		go func() {
			results, err := searchBySongName(song)
			fyne.Do(func() {
				if err != nil {
					dialog.ShowError(err, a.fyneWindow)
					return
				}
				shows = results
				a.resultsList.results.UnselectAll()
				a.resultsList.results.Refresh() // this is what makes the list redraw

				a.details.container.Hide()
				a.details.placeholderLayout.Show()
			})
		}()
	}

	// Placeholder buttons
	// TODO: Implement these
	a.resultsList.prevBtn = widget.NewButton("Next", func() {
		fmt.Println("Next...")
	})
	a.resultsList.nextBtn = widget.NewButton("Previous", func() {
		fmt.Println("Previous...")
	})
	addToPlaylistBtn := widget.NewButton("Add to Playlist", func() {
		fmt.Println("Add to playlist...")
	})
	streamBtn := widget.NewButton("Stream from archive.org", func() {
		fmt.Println("Streaming...")
	})
	musicPlayer := widget.NewLabel("Music Player")
	musicPlayer.Alignment = fyne.TextAlignCenter
	resultsPanel := container.NewBorder(
		nil,
		container.NewHBox(a.resultsList.nextBtn, layout.NewSpacer(), a.resultsList.prevBtn),
		nil, nil,
		a.resultsList.results,
	)
	detailsPanel := container.NewBorder(
		nil,
		container.NewHBox(addToPlaylistBtn, layout.NewSpacer(), streamBtn),
		nil, nil,
		container.NewStack(a.details.placeholderLayout, a.details.container),
	)

	dateSearch := container.NewHBox(a.search.dateSearch.yearChoice, a.search.dateSearch.monthChoice, a.search.dateSearch.dayChoice)
	searchPanel := container.NewVBox(a.search.textSearch.searchID, a.search.textSearch.enterBtn, dateSearch, a.search.dateSearch.enterBtn, a.search.songSearch.songEntry)
	searchAndList := container.NewHSplit(searchPanel, resultsPanel)
	content := container.NewHSplit(searchAndList, detailsPanel)
	content.SetOffset(0.4)
	root := container.NewBorder(nil, musicPlayer, nil, nil, content)

	a.fyneWindow.SetContent(root)
	a.fyneWindow.ShowAndRun()
}
