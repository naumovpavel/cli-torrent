package client

import (
	"sync/atomic"
)

type State int32

var (
	InProgress State = 0
	Downloaded State = 1
	Failed     State = 2
)

func (s State) String() string {
	switch s {
	case InProgress:
		return "downloading"
	case Downloaded:
		return "downloaded"
	case Failed:
		return "failed to download"
	default:
		return "Unknown state"
	}
}

type TorrentFileState struct {
	Name       string
	Dest       string
	State      atomic.Int32
	Pieces     int64
	Downloaded atomic.Int64
	Err        atomic.Pointer[error]
}

func NewState(pieces int64, name, dest string) *TorrentFileState {
	t := &TorrentFileState{
		Name:   name,
		Pieces: pieces,
		Dest:   dest,
	}
	t.Downloaded.Store(0)
	t.State.Store(0)
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
}

func (s *TorrentFileState) GetState() State {
	return State(s.State.Load())
}

func (s *TorrentFileState) UpdateErr(err error) {
	s.Err.Store(&err)
}

func (s *TorrentFileState) GetErr() error {
	return *s.Err.Load()
}
