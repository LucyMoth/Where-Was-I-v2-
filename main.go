package main

import (
	"fmt"
	"strconv"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func mainMenu(app *tview.Application) {
	menu := tview.NewList().
		AddItem("Browse Shows", "", '1', func() {
			browseShows(app)
		}).
		AddItem("Add New Show", "", '2', func() {
			searchShowsView(app)
		}).
		AddItem("Quit", "", 'q', func() {
			app.Stop()
		})

	menu.SetBorder(true).SetTitle("Where Was I?").SetTitleAlign(tview.AlignLeft)

	if err := app.SetRoot(menu, true).Run(); err != nil {
		fmt.Println("Error: ", err)
	}
}

func errorView(app *tview.Application, err error) {
	e := tview.NewTextView().SetText(err.Error() + "\n\nPress Q to go back to the main menu")

	e.SetBorder(true).SetTitle("Error").SetTitleAlign(tview.AlignLeft).SetBorderColor(tcell.ColorRed)

	e.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyRune {
			switch event.Rune() {
			case 'q':
				mainMenu(app)
				return nil
			}
		}
		return event
	})

	if err := app.SetRoot(e, true).Run(); err != nil {
		fmt.Println("Error: ", err)
	}
}

func browseShows(app *tview.Application) {
	submenu := tview.NewList().
		AddItem("Back", "", 'q', func() {
			mainMenu(app)
		})

	submenu.SetBorder(true).SetTitle("Browse Shows").SetTitleAlign(tview.AlignLeft)

	shows := listShows()

	for i, show := range shows {
		index := i + 1
		showCopy := show
		submenu.AddItem(showCopy.Name, "", rune('0'+index), func() {
			s, e := readShow(strconv.Itoa(showCopy.ID))
			if e != nil {
				errorView(app, e)
			}
			browseShowsSubMenu(app, s)
		})
	}

	app.SetRoot(submenu, true)
}

func browseShowsSubMenu(app *tview.Application, show tvshow) {
	showInfo := fmt.Sprintf("\nName: %s\n\nDescription: %s\n\nStart Date: %s\n\nEnd Date: %s\n\nStatus: %s\n\n",
		show.Name, show.Description, show.StartDate, show.EndDate, show.Status)
	tvShowInfo := tview.NewTextView().
		SetText(showInfo).
		SetTextAlign(tview.AlignLeft).
		SetDynamicColors(true)

	episodesTable := tview.NewTable().
		SetSelectable(true, false).
		SetSelectedStyle(tcell.StyleDefault.Background(tcell.ColorDarkMagenta).Foreground(tcell.ColorBlack)).
		SetBorders(true).
		SetBordersColor(tcell.ColorGray)

	episodesTable.SetCell(0, 0, tview.NewTableCell("Season").SetTextColor(tcell.ColorGray))
	episodesTable.SetCell(0, 1, tview.NewTableCell("Episode").SetTextColor(tcell.ColorGray))
	episodesTable.SetCell(0, 2, tview.NewTableCell("Name").SetTextColor(tcell.ColorGray))
	episodesTable.SetCell(0, 3, tview.NewTableCell("Air Date").SetTextColor(tcell.ColorGray))
	episodesTable.SetCell(0, 4, tview.NewTableCell("Seen").SetTextColor(tcell.ColorGray))

	var row int

	colorSeen := func(ep episode) tcell.Color {
		if ep.Seen {
			return tcell.ColorGreen
		} else {
			return tcell.ColorRed
		}
	}

	for i, ep := range show.Episodes {
		row = i + 1

		episodesTable.SetCell(row, 0, tview.NewTableCell(fmt.Sprintf("%d", ep.Season)))
		episodesTable.SetCell(row, 1, tview.NewTableCell(fmt.Sprintf("%d", ep.Episode)))
		episodesTable.SetCell(row, 2, tview.NewTableCell(ep.Name))
		episodesTable.SetCell(row, 3, tview.NewTableCell(ep.AirDate))
		episodesTable.SetCell(row, 4, tview.NewTableCell(strconv.FormatBool(ep.Seen)).SetTextColor(colorSeen(ep)))
	}

	next := func() string {
		for _, x := range show.Episodes {
			if x.Seen {
				continue
			} else {
				return x.Name
			}
		}
		return ""
	}
	tvShowFooter := tview.NewTextView().
		SetText(fmt.Sprintf("Next Unwatched Episode: %s", next())).
		SetTextAlign(tview.AlignLeft).
		SetDynamicColors(true)

	for col := 0; col < episodesTable.GetColumnCount(); col++ {
		cell := episodesTable.GetCell(0, col)
		cell.SetSelectable(false)
	}

	episodesTable.SetFixed(1, 0)

	flexinner := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(episodesTable, 0, 10, true).
		AddItem(nil, 0, 1, false).
		AddItem(tvShowFooter, 0, 2, false)

	flex := tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(flexinner, 0, 2, true).
		AddItem(tvShowInfo, 0, 1, false)

	flex.SetBorder(true).SetTitle("Browse Shows").SetTitleAlign(tview.AlignLeft).SetBorderColor(tcell.ColorPurple)

	flex.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyRune {
			switch event.Rune() {
			case 'q':
				browseShows(app)
				return nil
			}
		}
		return event
	})

	episodesTable.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEnter {
			row, _ := episodesTable.GetSelection()
			if row > 0 && row <= len(show.Episodes) {
				ep := &show.Episodes[row-1]
				ep.Seen = !ep.Seen
				episodesTable.SetCell(row, 4, tview.NewTableCell(strconv.FormatBool(ep.Seen)).SetTextColor(colorSeen(*ep)))
				tvShowFooter.SetText(fmt.Sprintf("Next Unwatched Episode: %s\nPress CTRL+R to remove show", next()))
				if e := writeShow(show, strconv.Itoa(show.ID)); e != nil {
					errorView(app, e)
				}
				return nil
			}
		} else if event.Key() == tcell.KeyCtrlR {
			if e := deleteShow(strconv.Itoa(show.ID)); e != nil {
				errorView(app, e)
			}
			browseShows(app)
		}
		return event
	})

	app.SetRoot(flex, true)
}

func searchShowsView(app *tview.Application) {
	inputField := tview.NewInputField().
		SetLabel("Enter TV Show Name: ").
		SetFieldWidth(30)

	doneFunc := func(key tcell.Key) {
		searchTerm := inputField.GetText()
		results, e := searchShows(searchTerm)

		if e != nil {
			errorView(app, e)
		}

		SearchShowsResult(app, results.Shows)
	}

	inputField.SetDoneFunc(doneFunc)

	flex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(inputField, 0, 1, true)

	app.SetRoot(flex, true)
}

func SearchShowsResult(app *tview.Application, shows []showjson) {
	submenu := tview.NewList().
		AddItem("Back", "", 'q', func() {
			mainMenu(app)
		})

	submenu.SetBorder(true).SetTitle("Add Show").SetTitleAlign(tview.AlignLeft)

	for i, show := range shows {
		index := i + 1
		showCopy := show
		submenu.AddItem(show.Name, "", rune('0'+index), func() {
			if e := downloadShow(showCopy.ID); e != nil {
				errorView(app, e)
			}
			mainMenu(app)
		})
	}

	app.SetRoot(submenu, true)
}

func main() {
	checkDirExsists()
	app := tview.NewApplication()
	mainMenu(app)
}
