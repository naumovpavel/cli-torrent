package commands

import (
	"strconv"

	torrentClient "github.com/naumovpavel/cli-torrent/internal/torrent/p2p/client"
)

type ContinueCommand struct {
	command
	client torrentClient.TorrentClient
}

func NewContinueCommand(torrentClient torrentClient.TorrentClient) *ContinueCommand {
	return &ContinueCommand{
		command: command{
			name: "continue",
			desc: "index, continues downloading a file with this index",
		},
		client: torrentClient,
	}
}

func (c *ContinueCommand) Execute(args ...string) string {
	if len(args) < 2 {
		return "too few arguments"
	}
	index, err := strconv.Atoi(args[1])
	if err != nil || index < 0 {
		return "argument must be positive number"
	}
	if len(c.client.GetFileStates()) < index {
		return "there are no file with this index"
	}
	c.client.GetFileStates()[index].UpdateState(torrentClient.InProgress)
	return "Downloading file with index " + strconv.Itoa(index) + " successfully continued!"
}
