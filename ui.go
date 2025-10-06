package main

import (
	"fmt"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"log/slog"
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

	// Participants
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

	// Stories
	stories := tview.NewTable()
	i := 0
	for _, story := range game.state.Stories {
		cell := tview.NewTableCell(story.Title).SetAlign(tview.AlignLeft)
		switch story.Status {
		case "closed":
			cell.SetTextColor(tcell.ColorGreen)
		}

		stories.SetCell(i, 0, cell)
		i++
	}

	stories.SetSelectable(true, false)

	stories.SetMouseCapture(func(action tview.MouseAction, event *tcell.EventMouse) (tview.MouseAction, *tcell.EventMouse) {
		slog.Debug(fmt.Sprintf("Mouse event. Action: %v; Event: %v", action, event))
		i1, i2, i3, i4 := stories.GetRect()
		slog.Debug(fmt.Sprintf("Stories rect: %d, %d, %d, %d", i1, i2, i3, i4))

		return action, event
	})
	stories.SetBorder(true).SetTitle("Stories")

	flex := tview.NewFlex().
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(gameData, 5, 1, false).
			AddItem(tview.NewFlex().SetDirection(tview.FlexColumn).
				AddItem(stories, 0, 2, false).
				AddItem(tview.NewBox().SetBorder(true).SetTitle("Preview"), 0, 3, false),
				0, 3, false).
			AddItem(tview.NewFlex().SetDirection(tview.FlexColumn).
				AddItem(tview.NewBox().SetBorder(true).SetTitle("Deck"), 0, 2, false).
				AddItem(participants, 50, 3, false),
				0, 3, false),
			0, 1, false)

	inputs := []tview.Primitive{
		gameData,
		stories,
		participants,
	}
	flex.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyTab {
			cycleFocus(app, inputs, false)
		} else if event.Key() == tcell.KeyBacktab {
			cycleFocus(app, inputs, true)
		}
		return event
	})

	return app.EnableMouse(true).SetRoot(flex, true).SetFocus(gameData)
}

func cycleFocus(app *tview.Application, elements []tview.Primitive, reverse bool) {
	for i, el := range elements {
		if !el.HasFocus() {
			continue
		}

		if reverse {
			i = i - 1
			if i < 0 {
				i = len(elements) - 1
			}
		} else {
			i = i + 1
			i = i % len(elements)
		}

		app.SetFocus(elements[i])
		return
	}
}
