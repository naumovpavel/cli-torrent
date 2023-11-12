package client

import (
	"log"
	"math/rand"
	"os"
	"runtime"
	"sync"

	"cli-torrent/internal/torrent/p2p/tracker"
	"cli-torrent/internal/torrent/torrentfile"
)

type TorrentClient interface {
	DownloadFile(torrentFile string, dst string) *TorrentFileState
}

const maxConcurrentDownloadingFiles = 16
const maxConcurrentWorkers = 1

var _ TorrentClient = &Client{}

type Client struct {
	peerId [20]byte
	port   uint16
	buf    []byte
}

func NewClient() TorrentClient {
	return &Client{
		peerId: [20]byte(genPeerId()),
		port:   10434,
		buf:    make([]byte, 0),
	}
}

func genPeerId() []byte {
	buf := make([]byte, 20)
	for i := 0; i < 20; i++ {
		buf[i] = byte(rand.Int())
	}
	return buf
}

func (c *Client) DownloadFile(torrentFile string, dst string) *TorrentFileState {
	tf, err := torrentfile.New(torrentFile)
	state := NewState(int64(tf.Info.Length), tf.Info.Name)
	if err != nil {
		state.State = Failed
		state.Err = err
		return state
	}

	c.startDownload(tf, state, dst)
	return state
}

func (c *Client) startDownload(tf *torrentfile.Torrentfile, state *TorrentFileState, dst string) {
	cnt := (tf.Info.Length + tf.Info.PieceLength - 1) / tf.Info.PieceLength
	jobChan := make(chan *Job, cnt)
	c.fillJobChan(jobChan, tf)
	resChan := make(chan *Result, cnt)
	semaphore := make(chan struct{}, maxConcurrentWorkers)
	//defer close(semaphore)
	bufer := make([]byte, tf.Info.Length)
	go c.accumulateRes(resChan, bufer, tf, state, jobChan)

	var wg *sync.WaitGroup = &sync.WaitGroup{}
	//for state.State == InProgress {
	//	trackerInfo, err := tracker.NewTracker(tf, c.peerId, c.port)
	//	//t0 := time.Now()
	//	if err != nil {
	//		state.State = Failed
	//		state.Err = err
	//		break
	//	}
	//
	//	break
	//	//break
	//	//time.Sleep(time.Now().Sub(t0.Add(time.Duration(trackerInfo.Interval))))
	//}
	trackerInfo, _ := tracker.NewTracker(tf, c.peerId, c.port)
	for _, peer := range trackerInfo.Peers {
		peer := peer
		if state.State != InProgress {
			break
		}
		//semaphore <- struct{}{}
		//p2pClient, err := NewP2PClient(peer, c.peerId, tf)
		//if err != nil {
		//	<-semaphore
		//	continue
		//}
		wg.Add(1)
		worker := NewWorker(peer, state, semaphore, c.peerId, tf)
		go worker.startWorker(jobChan, resChan, wg)
		//log.Println("worker stopped")
	}

	wg.Wait()
	log.Println("download done")

	if state.State == InProgress {
		err := c.saveToFile(dst)
		if err != nil {
			state.State = Failed
			state.Err = err
		}
	}
}

func (c *Client) saveToFile(dst string) error {
	outFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer outFile.Close()
	_, err = outFile.Write(c.buf)
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) accumulateRes(resChan chan *Result, buffer []byte, tf *torrentfile.Torrentfile, state *TorrentFileState, jobChan chan *Job) {
	donePieces := 0
	cnt := (tf.Info.Length + tf.Info.PieceLength - 1) / tf.Info.PieceLength
	for donePieces < len(tf.Info.PieceHashes) {
		res := <-resChan
		log.Println("goroutines ", runtime.NumGoroutine())
		log.Println((float64(donePieces)*100)/float64(cnt), "%, pieces ", donePieces, " downloaded")
		begin, end := tf.CalculateBoundsForPiece(res.index)
		copy(buffer[begin:end], res.buf)
		state.UpdateDownloadedCount(tf.CalculatePieceSize(res.index))
		donePieces++
	}
	close(jobChan)
}

func (c *Client) fillJobChan(jobChan chan *Job, tf *torrentfile.Torrentfile) {
	for index, hash := range tf.Info.PieceHashes {
		length := tf.CalculatePieceSize(index)
		jobChan <- &Job{index, hash, length}
	}
}
