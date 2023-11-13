package client

import (
	"errors"
	"log"
	"math/rand"
	"os"
	"time"

	"cli-torrent/internal/torrent/p2p/tracker"
	"cli-torrent/internal/torrent/torrentfile"
	"github.com/gammazero/deque"
)

type TorrentClient interface {
	DownloadFile(torrentFile string, dst string) error
	GetFileStates() []*TorrentFileState
}

var _ TorrentClient = &Client{}

type Client struct {
	peerId     [20]byte
	port       uint16
	buf        []byte
	fileStates []*TorrentFileState
}

func NewClient() *Client {
	return &Client{
		peerId:     [20]byte(genPeerId()),
		port:       10434,
		buf:        make([]byte, 0),
		fileStates: make([]*TorrentFileState, 0),
	}
}

func genPeerId() []byte {
	buf := make([]byte, 20)
	for i := 0; i < 20; i++ {
		buf[i] = byte(rand.Int())
	}
	return buf
}

func (c *Client) DownloadFile(torrentFile, dst string) error {
	tf, err := torrentfile.New(torrentFile)
	if err != nil {
		return err
	}
	pieces := (tf.Info.Length + tf.Info.PieceLength - 1) / tf.Info.PieceLength
	state := NewState(int64(pieces), tf.Info.Name, dst)
	c.fileStates = append(c.fileStates, state)
	go c.startDownload(tf, state, dst)
	return nil
}

func (c *Client) GetFileStates() []*TorrentFileState {
	return c.fileStates
}

func (c *Client) startDownload(tf *torrentfile.Torrentfile, state *TorrentFileState, dst string) {
	cnt := (tf.Info.Length + tf.Info.PieceLength - 1) / tf.Info.PieceLength
	jobChan := make(chan *Job, cnt)
	resChan := make(chan *Result, cnt)
	downloadHistory := make(chan DownloadHistoryEntry, cnt*(tf.Info.PieceLength+maxBlockSize-1)/maxBlockSize)
	c.fillJobChan(jobChan, tf)
	buffer := make([]byte, tf.Info.Length)

	go c.accumulateRes(resChan, buffer, tf, state, jobChan, downloadHistory)
	go c.processSpeedChanging(downloadHistory, state)
	c.downloadFileFromPeers(tf, state, jobChan, resChan, downloadHistory)
	c.saveDownloadedFile(state, dst, buffer)
}

func (c *Client) saveDownloadedFile(state *TorrentFileState, dst string, buffer []byte) {
	if state.GetState() == Downloaded {
		err := c.saveToFile(dst, buffer)
		if err != nil {
			log.Println(err)
			state.UpdateState(Failed)
			state.UpdateErr(err)
		}
	}
}

func (c *Client) downloadFileFromPeers(tf *torrentfile.Torrentfile, state *TorrentFileState, jobChan chan *Job, resChan chan *Result, downloadHistory chan DownloadHistoryEntry) {
	for state.GetState() == InProgress {
		trackerInfo, err := tracker.NewTracker(tf, c.peerId, c.port)
		log.Println(len(trackerInfo.Peers))
		if err != nil {
			state.UpdateState(Failed)
			state.UpdateErr(err)
			return
		}
		for _, peer := range trackerInfo.Peers {
			if _, ok := state.WorkingPeers.Load(peer.String()); ok {
				continue
			}
			if state.GetState() != InProgress {
				break
			}
			worker := NewWorker(peer, state, c.peerId, tf)
			go worker.startWorker(jobChan, resChan, state, downloadHistory)
		}
		time.Sleep(time.Duration(trackerInfo.Interval) * time.Second)
	}
}

var (
	ErrFailedToOpenDstFile = errors.New("failed to open destination file")
	ErrFailedToWrite       = errors.New("error while writing to destination file")
)

func (c *Client) saveToFile(dst string, buffer []byte) error {
	outFile, err := os.Create(dst)
	if err != nil {
		return ErrFailedToOpenDstFile
	}
	defer outFile.Close()
	_, err = outFile.Write(buffer)
	if err != nil {
		return ErrFailedToWrite
	}
	return nil
}

func (c *Client) accumulateRes(resChan chan *Result, buffer []byte, tf *torrentfile.Torrentfile, state *TorrentFileState, jobChan chan *Job, downloadHistory chan DownloadHistoryEntry) {
	donePieces := 0
	for donePieces < len(tf.Info.PieceHashes) {
		res := <-resChan
		begin, end := tf.CalculateBoundsForPiece(res.index)
		copy(buffer[begin:end], res.buf)
		state.UpdateDownloadedCount(1)
		donePieces++
	}
	close(jobChan)
	close(downloadHistory)
	state.UpdateState(Downloaded)
}

func (c *Client) fillJobChan(jobChan chan *Job, tf *torrentfile.Torrentfile) {
	for index, hash := range tf.Info.PieceHashes {
		length := tf.CalculatePieceSize(index)
		jobChan <- &Job{index, hash, length}
	}
}

func (c *Client) processSpeedChanging(downloadHistory chan DownloadHistoryEntry, state *TorrentFileState) {
	history := deque.New[DownloadHistoryEntry]()
	var sumOfBytes = 0
	for entry := range downloadHistory {
		secondAgo := time.Now().Add(-1 * time.Second)
		for history.Len() > 0 && history.Back().time.Before(secondAgo) {
			elem := history.PopBack()
			sumOfBytes -= elem.size
		}
		if entry.time.After(secondAgo) {
			sumOfBytes += entry.size
			history.PushFront(entry)
		}
		state.UpdateSpeed(float64(sumOfBytes) / float64(1024*1024))
	}
}
