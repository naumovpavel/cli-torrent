package client

import (
	"sync"
	"sync/atomic"
)

type State int32

var (
	InProgress State = 0
	Downloaded State = 1
	Failed     State = 2
	Stopped    State = 3
	Paused     State = 4
)

func (s State) String() string {
	switch s {
	case InProgress:
		return "downloading"
	case Downloaded:
		return "downloaded"
	case Failed:
		return "failed to download"
	case Paused:
		return "paused"
	case Stopped:
		return "stopped"
	default:
		return "Unknown state"
	}
}

type TorrentFileState struct {
	Name         string
	Dest         string
	State        atomic.Int32
	Pieces       int64
	Downloaded   atomic.Int64
	Err          atomic.Pointer[error]
	WorkingPeers sync.Map
	Speed        atomic.Pointer[float64]
}

func NewState(pieces int64, name, dest string) *TorrentFileState {
	t := &TorrentFileState{
		Name:   name,
		Pieces: pieces,
		Dest:   dest,
	}
	var speed float64 = 0
	t.Downloaded.Store(0)
	t.State.Store(0)
	t.Speed.Store(&speed)
	t.Err.Store(nil)
	return t
}

func (s *TorrentFileState) UpdateDownloadedCount(delta int64) {
	newDownloadedVal := s.Downloaded.Add(delta)
	if s.Pieces == newDownloadedVal {
		s.State.Store(int32(Downloaded))
	}
}

func (s *TorrentFileState) GetDownloadedCount() int64 {
	return s.Downloaded.Load()
}

func (s *TorrentFileState) UpdateState(state State) {
	s.State.Store(int32(state))
	if state == Paused || state == Stopped {
		s.UpdateSpeed(0)
	}
}

func (s *TorrentFileState) GetState() State {
	return State(s.State.Load())
}

func (s *TorrentFileState) UpdateSpeed(speed float64) {
	s.Speed.Store(&speed)
}

func (s *TorrentFileState) GetSpeed() float64 {
	return *s.Speed.Load()
}

func (s *TorrentFileState) UpdateErr(err error) {
	s.Err.Store(&err)
}

func (s *TorrentFileState) GetErr() error {
	return *s.Err.Load()
}
