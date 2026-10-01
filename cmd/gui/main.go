package main

import (
	"encoding/json"
	"net/http"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/LunarDrift/deadabase/internal"
)

func main() {
	app := newApp()
	app.setupResultCanvas()

	showResult := container.NewVBox(
		app.showResultSingle.showIDLabel,
		app.showResultSingle.dateLabel,
		app.showResultSingle.locationLabel,
		app.showResultSingle.notesLabel,
		app.showResultSingle.setsLabel,
		app.showResultSingle.footnotesLabel,
	)

	app.searchPanel.enterBtn = widget.NewButton("Search", app.GetShowFromID)
	app.searchPanel.OnEnter = app.GetShowFromID

	var data internal.Paginated[internal.ShowMeta]

	resp, err := http.Get("http://localhost:8080/shows?song=althea")
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close() //nolint:errcheck
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		panic(err)
	}

	list := widget.NewList(
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

	searchPanel := container.NewVBox(app.searchPanel, app.searchPanel.enterBtn)
	thing := container.NewHSplit(searchPanel, list)
	content := container.NewHSplit(thing, showResult)
	content.SetOffset(0.25)

	app.fyneWindow.SetContent(container.NewVBox(content))
	app.fyneWindow.ShowAndRun()
}
