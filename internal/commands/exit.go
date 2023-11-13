package commands

import (
	torrentClient "github.com/naumovpavel/cli-torrent/internal/torrent/p2p/client"
)

type ExitCommand struct {
	command
	client torrentClient.TorrentClient
	cm     *CommandManger
}

func NewExitCommand(torrentClient torrentClient.TorrentClient, cm *CommandManger) *ExitCommand {
	return &ExitCommand{
		command: command{
			name: "exit",
			desc: "stops all downloads and exits the program",
		},
		client: torrentClient,
		cm:     cm,
	}
}

func (e *ExitCommand) Execute(args ...string) string {
	e.client.StopAll()
	e.cm.running = false
	return "Goodbye!"
}
