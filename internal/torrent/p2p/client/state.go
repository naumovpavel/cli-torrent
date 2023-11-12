package client

import (
	"sync/atomic"
)

type State uint8

var (
	InProgress State = 0
	Downloaded State = 1
	Failed     State = 2
)

type TorrentFileState struct {
	Name       string
	State      State
	Length     atomic.Int64
	Downloaded atomic.Int64
	Err        error
}

func NewState(length int64, name string) *TorrentFileState {
	t := &TorrentFileState{
		State: InProgress,
		Name:  name,
	}
	t.Length.Store(length)
	t.Downloaded.Store(0)
	return t
}

func (s *TorrentFileState) UpdateDownloadedCount(delta int) {
	new := s.Downloaded.Add(int64(delta))
	if s.Length.Load() == new {
		s.State = Downloaded
	}
}

func (s *TorrentFileState) UpdateState(state State) {
	s.State = state
}
