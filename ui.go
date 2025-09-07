package main

import "github.com/rivo/tview"

func buildUi(game *Game) *tview.Application {
	app := tview.NewApplication()

	flex := tview.NewFlex().
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(tview.NewBox().SetBorder(true).SetTitle("Firepoker CLI"), 0, 1, false).
			AddItem(tview.NewFlex().SetDirection(tview.FlexColumn).
				AddItem(tview.NewBox().SetBorder(true).SetTitle("Stories"), 0, 2, false).
				AddItem(tview.NewBox().SetBorder(true).SetTitle("Preview"), 0, 3, false),
				0, 3, false).
			AddItem(tview.NewFlex().SetDirection(tview.FlexColumn).
				AddItem(tview.NewBox().SetBorder(true).SetTitle("Deck"), 0, 2, false).
				AddItem(tview.NewBox().SetBorder(true).SetTitle("Participants"), 50, 3, false),
				0, 3, false),
			0, 1, false)

	return app.SetRoot(flex, true).SetFocus(flex)
}
