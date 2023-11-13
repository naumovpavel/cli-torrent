package commands

import (
	torrentClient "github.com/naumovpavel/cli-torrent/internal/torrent/p2p/client"
)

type DownloadCommand struct {
	command
	client torrentClient.TorrentClient
}

func NewDownloadCommand(torrentClient torrentClient.TorrentClient) *DownloadCommand {
	return &DownloadCommand{
		command: command{
			name: "download",
			desc: "src dest, download torrent file from src to dest",
		},
		client: torrentClient,
	}
}

func (d *DownloadCommand) Execute(args ...string) string {
	if len(args) < 3 {
		return "too few arguments for command, use help to see download arguments"
	}
	err := d.client.DownloadFile(args[1], args[2])
	if err != nil {
		return err.Error()
	}
	return "file successfully start downloading"
}
