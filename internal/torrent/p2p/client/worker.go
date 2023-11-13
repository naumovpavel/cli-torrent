package client

import (
	"bytes"
	"crypto/sha1"
	"time"

	"github.com/naumovpavel/cli-torrent/internal/torrent/p2p/message"
	"github.com/naumovpavel/cli-torrent/internal/torrent/p2p/tracker"
	"github.com/naumovpavel/cli-torrent/internal/torrent/torrentfile"
)

type Job struct {
	index  int
	hash   [20]byte
	length int
}

type Result struct {
	index int
	buf   []byte
}

type DownloadHistoryEntry struct {
	size int
	time time.Time
}

const maxBacklog = 5
const maxBlockSize = 16384

type Worker struct {
	peer      *tracker.Peer
	tf        *torrentfile.Torrentfile
	p2pClient *P2PClient
	state     *TorrentFileState
	peerID    [20]byte
}

func NewWorker(peer *tracker.Peer, state *TorrentFileState, peerID [20]byte, tf *torrentfile.Torrentfile) *Worker {
	return &Worker{
		peer:   peer,
		state:  state,
		peerID: peerID,
		tf:     tf,
	}
}

func (w *Worker) startWorker(jobChan chan *Job, resChan chan *Result, state *TorrentFileState, downloadHistory chan DownloadHistoryEntry) {
	defer state.WorkingPeers.Delete(w.peer.String())
	p2pClient, err := NewP2PClient(w.peer, w.peerID, w.tf)
	if err != nil {
		return
	}
	w.p2pClient = p2pClient
	defer w.p2pClient.conn.Close()

	err = message.NewUnchoke().Send(w.p2pClient.conn)
	if err != nil {
		return
	}
	err = message.NewInterested().Send(w.p2pClient.conn)
	if err != nil {
		return
	}

	w.downloadPieces(jobChan, resChan, downloadHistory, state)
}

func (w *Worker) downloadPieces(jobChan chan *Job, resChan chan *Result, downloadHistory chan DownloadHistoryEntry, state *TorrentFileState) {
	for job := range jobChan {
		if state.GetState() != InProgress {
			jobChan <- job
			if state.GetState() == Paused {
				time.Sleep(100 * time.Millisecond)
				continue
			} else {
				return
			}
		}
		if !w.p2pClient.Bitfield.HasPiece(job.index) {
			jobChan <- job
			continue
		}

		buf, err := w.downloadPiece(job, downloadHistory)
		if err != nil {
			jobChan <- job
			return
		}

		err = checkIntegrity(buf, job.hash[:], job.index)
		if err != nil {
			jobChan <- job
			continue
		}

		message.NewHave(&job.index).Send(w.p2pClient.conn)
		resChan <- &Result{
			index: job.index,
			buf:   buf,
		}
	}
}

func (w *Worker) downloadPiece(job *Job, downloadHistory chan DownloadHistoryEntry) ([]byte, error) {
	w.p2pClient.downloaded = 0
	w.p2pClient.backlog = 0
	buf := make([]byte, job.length)

	w.p2pClient.conn.SetDeadline(time.Now().Add(30 * time.Second))
	defer w.p2pClient.conn.SetDeadline(time.Time{})

	err := w.tryToDownloadPiece(job, buf, downloadHistory)
	if err != nil {
		return buf, err
	}

	return buf, nil
}

func (w *Worker) tryToDownloadPiece(job *Job, buf []byte, downloadHistory chan DownloadHistoryEntry) error {
	requested := 0
	for w.p2pClient.downloaded < job.length {
		if !w.p2pClient.chocked {
			err := w.sendRequests(job, &requested)
			if err != nil {
				return err
			}
		}

		downloadedOld := w.p2pClient.downloaded
		err := w.p2pClient.readMessage(buf, job.index)
		if err != nil {
			return err
		}

		if w.p2pClient.downloaded > downloadedOld {
			downloadHistory <- DownloadHistoryEntry{
				size: w.p2pClient.downloaded - downloadedOld,
				time: time.Now(),
			}
		}
	}
	return nil
}

func (w *Worker) sendRequests(job *Job, requested *int) error {
	for w.p2pClient.backlog < maxBacklog && *requested < job.length {
		blockSize := maxBlockSize

		if job.length-*requested < maxBlockSize {
			blockSize = job.length - *requested
		}

		err := message.NewRequest(job.index, *requested, blockSize).Send(w.p2pClient.conn)
		if err != nil {
			return err
		}

		*requested += blockSize
		w.p2pClient.backlog++
	}
	return nil
}

func checkIntegrity(buf []byte, hash []byte, index int) error {
	bufHash := sha1.Sum(buf)
	if !bytes.Equal(bufHash[:], hash[:]) {
		return message.ErrBadMessage
	}
	return nil
}
