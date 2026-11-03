package main

func main() {
	a := newApp()
	go a.loadLookups()
	a.win.ShowAndRun()
}
