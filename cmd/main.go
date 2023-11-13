package main

import (
	_ "net/http/pprof"

	"github.com/naumovpavel/cli-torrent/internal/commands"
	"github.com/naumovpavel/cli-torrent/internal/torrent/p2p/client"
)

const greeting = "Hello! It's cli bit-torrent client, use help to see command list"

func main() {
	client := client.NewClient()
	cm := commands.NewCommandManager(client)
	println(greeting)
	cm.Run()
}
