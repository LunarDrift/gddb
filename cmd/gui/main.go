package main

import (
	"fmt"
	"strconv"

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
	months := []string{"January", "February", "March", "April", "May", "June", "July", "August", "September", "October", "November", "December"}
	// TODO: Different number of days depending on the month
	for i := 1; i <= 31; i++ {
		days = append(days, strconv.Itoa(i))
	}
	for i := 1965; i <= 1995; i++ {
		years = append(years, strconv.Itoa(i))
	}

	a := &App{
		fyneApp: app.New(),
		search: &searchPanel{
			textSearch: showIDSearch{
				searchID: widget.NewEntry(),
			},
		},
		resultsListPnl: resultsListPanel{
			showResults: &internal.Paginated[internal.ShowMeta]{},
		},
	}
	a.fyneWindow = a.fyneApp.NewWindow("Deadabase")
	a.fyneWindow.Resize(fyne.NewSize(1280, 720))

	a.setupSearchPanel(days, months, years)
	a.setupResultCanvas()

	topMetaData := container.NewVBox(
		a.details.showDetails.showIDLabel,
		a.details.showDetails.dateLabel,
		a.details.showDetails.locationLabel,
		a.details.showDetails.notesLabel,
	)

	scrollableSets := container.NewVScroll(a.details.showDetails.setsLabel)
	scrollableFootnotes := container.NewVScroll(a.details.showDetails.footnotesLabel)
	scrollableFootnotes.SetMinSize(fyne.NewSize(0, 80))
	a.details.container = container.NewBorder(
		topMetaData,
		scrollableFootnotes,
		nil,
		nil,
		scrollableSets,
	)

	a.details.placeholderLayout = container.NewCenter(a.details.showDetails.detailLabel)
	a.resultsListPnl.resultsList = widget.NewList(
		func() int { return len(a.resultsListPnl.showResults.Results) },
		func() fyne.CanvasObject { return widget.NewLabel("template") },
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			obj.(*widget.Label).SetText(a.resultsListPnl.showResults.Results[id].Date)
		},
	)

	a.resultsListPnl.resultsList.OnSelected = func(id widget.ListItemID) {
		date := a.resultsListPnl.showResults.Results[id].Date
		go func() {
			show, err := getShowFromDate(date)
			fyne.Do(func() {
				if err != nil {
					dialog.ShowError(err, a.fyneWindow)
					return
				}
				a.updateMultiShowDetails(show)
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
				a.setResults(&results)
			})
		}()
	}

	// Placeholder buttons
	// TODO: Implement these
	a.resultsListPnl.prevBtn = widget.NewButton("Previous", func() { a.load(*a.resultsListPnl.showResults.Previous) })
	a.resultsListPnl.nextBtn = widget.NewButton("Next", func() { a.load(*a.resultsListPnl.showResults.Next) })
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
		container.NewHBox(a.resultsListPnl.prevBtn, layout.NewSpacer(), a.resultsListPnl.nextBtn),
		nil, nil,
		a.resultsListPnl.resultsList,
	)
	detailsPanel := container.NewBorder(
		nil,
		container.NewHBox(addToPlaylistBtn, layout.NewSpacer(), streamBtn),
		nil, nil,
		container.NewStack(a.details.placeholderLayout, a.details.container),
	)

	dateSearch := container.NewHBox(a.search.dateSearch.yearChoice, a.search.dateSearch.monthChoice, a.search.dateSearch.dayChoice)
	searchPanel := container.NewVBox(
		a.search.textSearch.searchID,
		a.search.textSearch.enterBtn,
		dateSearch, a.search.dateSearch.enterBtn,
		a.search.songSearch.songEntry,
		a.search.songSearch.enterBtn,
	)
	searchAndList := container.NewHSplit(searchPanel, resultsPanel)
	content := container.NewHSplit(searchAndList, detailsPanel)
	content.SetOffset(0.4)
	root := container.NewBorder(nil, musicPlayer, nil, nil, content)

	a.fyneWindow.SetContent(root)
	a.updateButtons()
	a.fyneWindow.ShowAndRun()
}
