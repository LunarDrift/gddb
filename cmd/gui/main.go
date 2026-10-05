package main

import (
	"encoding/json"
	"net/http"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
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

	var data internal.Paginated[internal.ShowMeta]

	resp, err := http.Get("http://localhost:8080/shows?song=althea")
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

	dateSearch := container.NewHBox(a.search.dateSearch.yearChoice, a.search.dateSearch.monthChoice, a.search.dateSearch.dayChoice)
	searchPanel := container.NewVBox(a.search.textSearch, a.search.textSearch.enterBtn, dateSearch, a.search.dateSearch.enterBtn)
	searchAndList := container.NewHSplit(searchPanel, a.resultsList)
	content := container.NewHSplit(searchAndList, a.detailsContainer)
	content.SetOffset(0.4)

	a.fyneWindow.SetContent(container.NewVBox(content))
	a.fyneWindow.ShowAndRun()
}
