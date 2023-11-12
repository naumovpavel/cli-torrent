package tracker

import (
	"errors"
	"log"
	"net/http"
	"time"

	"cli-torrent/internal/torrent/torrentfile"
	"github.com/jackpal/bencode-go"
)

type Tracker struct {
	Interval int
	Peers    []*Peer
}

type beeTracker struct {
	Interval int    `becode:"interval"`
	Peers    string `bencode:"peers"`
}

var (
	ErrMalformedPeers        = errors.New("peers is malformed, size of each peer must be 6 bytes")
	ErrTrackerDoesntResponse = errors.New("tracker doesn't response")
	ErrTrackerBadResponse    = errors.New("tracker response in incorrect format")
)

func NewTracker(t *torrentfile.Torrentfile, peerID [20]byte, port uint16) (*Tracker, error) {
	bt, err := requestPeers(t, peerID, port)
	if err != nil {
		return nil, err
	}

	const peerSize = 6
	buf := []byte(bt.Peers)
	if len(buf)%peerSize != 0 {
		return nil, ErrMalformedPeers
	}
	cnt := len(buf) / peerSize

	tr := &Tracker{
		Interval: bt.Interval,
		Peers:    make([]*Peer, cnt),
	}

	for i := 0; i < cnt; i++ {
		tr.Peers[i] = &Peer{}
		tr.Peers[i].Deserialize([6]byte(buf[i*peerSize : (i+1)*peerSize]))
	}
	return tr, nil
}

func requestPeers(t *torrentfile.Torrentfile, peerID [20]byte, port uint16) (beeTracker, error) {
	url, err := t.BuildTrackerUrl(peerID, port)

	if err != nil {
		return beeTracker{}, err
	}

	client := http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		log.Println(err)
		return beeTracker{}, ErrTrackerDoesntResponse
	}
	defer resp.Body.Close()

	var bt beeTracker
	err = bencode.Unmarshal(resp.Body, &bt)
	if err != nil {
		return beeTracker{}, ErrTrackerBadResponse
	}

	return bt, nil
}
