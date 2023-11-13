package commands

import (
	"strconv"

	torrentClient "github.com/naumovpavel/cli-torrent/internal/torrent/p2p/client"
)

type PauseCommand struct {
	command
	client torrentClient.TorrentClient
}

func NewPauseCommand(torrentClient torrentClient.TorrentClient) *PauseCommand {
	return &PauseCommand{
		command: command{
			name: "pause",
			desc: "index, pauses downloading a file with this index",
		},
		client: torrentClient,
	}
}

func (p *PauseCommand) Execute(args ...string) string {
	if len(args) < 2 {
		return "too few arguments"
	}
	index, err := strconv.Atoi(args[1])
	if err != nil || index < 0 {
		return "argument must be positive number"
	}
	if len(p.client.GetFileStates()) < index {
		return "there are no file with this index"
	}
	p.client.GetFileStates()[index].UpdateState(torrentClient.Paused)
	return "File with index " + strconv.Itoa(index) + " successfully paused!"
}
