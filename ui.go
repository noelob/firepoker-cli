package main

import (
	"fmt"
	"github.com/rivo/tview"
	"strings"
)

func buildUi(game *Game) *tview.Application {
	app := tview.NewApplication()

	gameData := tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(true).
		SetText(fmt.Sprintf("%s\n%s\n", game.state.Name, game.state.Description)).
		SetChangedFunc(func() {
			app.Draw()
		})

	gameData.SetBorder(true).SetTitle("Firepoker CLI")

	names := make([]string, 0)
	for _, user := range game.state.Participants {
		names = append(names, user.FullName)
	}

	participants := tview.NewTextView().
		SetText(strings.Join(names, "\n")).
		SetChangedFunc(func() {
			app.Draw()
		})

	participants.SetBorder(true).SetTitle("Participants")

	flex := tview.NewFlex().
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(gameData, 5, 1, false).
			AddItem(tview.NewFlex().SetDirection(tview.FlexColumn).
				AddItem(tview.NewBox().SetBorder(true).SetTitle("Stories"), 0, 2, false).
				AddItem(tview.NewBox().SetBorder(true).SetTitle("Preview"), 0, 3, false),
				0, 3, false).
			AddItem(tview.NewFlex().SetDirection(tview.FlexColumn).
				AddItem(tview.NewBox().SetBorder(true).SetTitle("Deck"), 0, 2, false).
				AddItem(participants, 50, 3, false),
				0, 3, false),
			0, 1, false)

	return app.SetRoot(flex, true).SetFocus(flex)
}
