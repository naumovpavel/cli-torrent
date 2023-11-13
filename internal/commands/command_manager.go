package commands

import (
	"bufio"
	"os"
	"strings"

	"github.com/naumovpavel/cli-torrent/internal/torrent/p2p/client"
)

type CommandManger struct {
	commands map[string]Command
	client   client.TorrentClient
	running  bool
}

func NewCommandManager(torrentClient client.TorrentClient) *CommandManger {
	cm := &CommandManger{
		client:   torrentClient,
		commands: make(map[string]Command),
	}
	registerCommands(cm)
	return cm
}

func registerCommands(cm *CommandManger) {
	commands := []Command{
		NewDownloadCommand(cm.client),
		NewShowCommand(cm.client),
		NewExitCommand(cm.client, cm),
		NewPauseCommand(cm.client),
		NewContinueCommand(cm.client),
		NewHelpCommand(cm.client, cm),
	}
	for _, curCommand := range commands {
		cm.commands[curCommand.Name()] = curCommand
	}
}

func (m *CommandManger) Run() {
	m.running = true
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		args := strings.Split(scanner.Text(), " ")
		command, s := m.commands[args[0]]
		if !s {
			println("Unknown command, use help to see commands list")
			continue
		}
		println(command.Execute(args...))
		if !m.running {
			break
		}
	}
}
