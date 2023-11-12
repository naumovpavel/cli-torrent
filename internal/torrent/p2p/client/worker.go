package client

import (
	"bytes"
	"crypto/sha1"
	"sync"
	"time"

	"cli-torrent/internal/torrent/p2p/message"
	"cli-torrent/internal/torrent/p2p/tracker"
	"cli-torrent/internal/torrent/torrentfile"
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

func (w *Worker) startWorker(jobChan chan *Job, resChan chan *Result, wg *sync.WaitGroup) {
	defer wg.Done()
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

	w.downloadPieces(jobChan, resChan)
}

func (w *Worker) downloadPieces(jobChan chan *Job, resChan chan *Result) {
	for job := range jobChan {
		if !w.p2pClient.Bitfield.HasPiece(job.index) {
			jobChan <- job
			continue
		}

		buf, err := w.downloadPiece(job)
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

func (w *Worker) downloadPiece(job *Job) ([]byte, error) {
	w.p2pClient.downloaded = 0
	w.p2pClient.backlog = 0
	buf := make([]byte, job.length)

	w.p2pClient.conn.SetDeadline(time.Now().Add(30 * time.Second))
	defer w.p2pClient.conn.SetDeadline(time.Time{})

	err := w.tryToDownloadPiece(job, buf)
	if err != nil {
		return buf, err
	}

	return buf, nil
}

func (w *Worker) tryToDownloadPiece(job *Job, buf []byte) error {
	requested := 0

	for w.p2pClient.downloaded < job.length {
		if !w.p2pClient.chocked {
			err := w.sendRequests(job, &requested)
			if err != nil {
				return err
			}
		}

		err := w.p2pClient.readMessage(buf, job.index)
		if err != nil {
			return err
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
