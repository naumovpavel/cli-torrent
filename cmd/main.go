package main

import (
	"net/http"
	"os"

	_ "net/http/pprof"

	"cli-torrent/internal/torrent/p2p/client"
)

func main() {
	inPath := os.Args[1]
	outPath := os.Args[2]
	go http.ListenAndServe("0.0.0.0:8085", nil)
	client := client.NewClient()
	client.DownloadFile(inPath, outPath)
}
