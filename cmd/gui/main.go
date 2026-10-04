package main

import (
	"encoding/json"
	"net/http"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/LunarDrift/deadabase/internal"
)

func main() {
	a := &App{
		fyneApp:   app.New(),
		searchBox: newSearchPanel(),
	}
	a.fyneWindow = a.fyneApp.NewWindow("Deadabase")
	a.fyneWindow.Resize(fyne.NewSize(1280, 720))

	a.setupResultCanvas()
	a.searchBox.enterBtn = widget.NewButton("Search", a.GetShowFromID)
	a.searchBox.OnEnter = a.GetShowFromID

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

	searchPanel := container.NewVBox(a.searchBox, a.searchBox.enterBtn)
	searchAndList := container.NewHSplit(searchPanel, a.resultsList)
	content := container.NewHSplit(searchAndList, a.detailsContainer)
	content.SetOffset(0.5)

	a.fyneWindow.SetContent(container.NewVBox(content))
	a.fyneWindow.ShowAndRun()
}
