package commands

import (
	"fmt"
	"strings"

	torrentClient "github.com/naumovpavel/cli-torrent/internal/torrent/p2p/client"
)

type HelpCommand struct {
	command
	client torrentClient.TorrentClient
	cm     *CommandManger
}

func NewHelpCommand(torrentClient torrentClient.TorrentClient, cm *CommandManger) *HelpCommand {
	return &HelpCommand{
		command: command{
			name: "help",
			desc: "shows all commands with description",
		},
		client: torrentClient,
		cm:     cm,
	}
}

func (h *HelpCommand) Execute(args ...string) string {
	ans := strings.Builder{}
	for _, command := range h.cm.commands {
		ans.WriteString(fmt.Sprintf("%s", command.Name()))
		ans.WriteString(" ")
		ans.WriteString(command.Desc())
		ans.WriteString("\n")
	}
	return ans.String()
}
