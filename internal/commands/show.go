package commands

import (
	"context"
	"strconv"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/naumovpavel/cli-torrent/internal/torrent/p2p/client"
	"github.com/rivo/tview"
)

type ShowCommand struct {
	command
	client client.TorrentClient
}

func NewShowCommand(torrentClient client.TorrentClient) *ShowCommand {
	return &ShowCommand{
		command: command{
			name: "show",
			desc: "shows all torrent files that you attempt to download, press escape to exit from show command",
		},
		client: torrentClient,
	}
}

func (s *ShowCommand) Execute(args ...string) string {
	app := tview.NewApplication()
	table := tview.NewTable().SetBorders(true)
	ctx, cancel := context.WithCancel(context.Background())
	table.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEscape {
			app.Stop()
			cancel()
		}
	})
	go s.draw(ctx, table, app)
	println("press escape to close show")
	if err := app.SetRoot(table, true).SetFocus(table).Run(); err != nil {
		return "oops.. something has gone wrong, please try again!"
	}
	return ""
}

func (s *ShowCommand) draw(ctx context.Context, table *tview.Table, app *tview.Application) {
	columns := []string{
		"id",
		"name",
		"destination",
		"status",
		"downloaded",
		"speed (MBite/s)",
	}
	const frames = 30
	for {
		select {
		case <-ctx.Done():
			return
		default:

		}
		app.QueueUpdateDraw(func() {
			s.refreshTable(columns, table)
		})

		time.Sleep(time.Second / frames)
	}

}

func (s *ShowCommand) refreshTable(columns []string, table *tview.Table) {
	table.Clear()
	for i, column := range columns {
		table.SetCell(0, i, tview.NewTableCell(column))
	}
	for i, state := range s.client.GetFileStates() {
		table.SetCell(i+1, 0, tview.NewTableCell(strconv.Itoa(i)).SetAlign(tview.AlignCenter))
		table.SetCell(i+1, 1, tview.NewTableCell(state.Name).SetAlign(tview.AlignCenter))
		table.SetCell(i+1, 2, tview.NewTableCell(state.Dest).SetAlign(tview.AlignCenter))
		table.SetCell(i+1, 3, tview.NewTableCell(state.GetState().String()).SetAlign(tview.AlignCenter))
		downPercent := float64(state.GetDownloadedCount()) * 100 / float64(state.Pieces)
		table.SetCell(i+1, 4, tview.NewTableCell(strconv.FormatFloat(downPercent, 'g', 3, 64)+"%").SetAlign(tview.AlignCenter))
		table.SetCell(i+1, 5, tview.NewTableCell(strconv.FormatFloat(state.GetSpeed(), 'g', 3, 64)).SetAlign(tview.AlignCenter))

	}
}
