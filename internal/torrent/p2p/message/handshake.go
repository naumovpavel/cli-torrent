package message

import (
	"errors"
	"io"
)

type Handshake struct {
	Pstr     string
	InfoHash [20]byte
	PeerID   [20]byte
}

const handshakeSize = 68

func NewHandshake(infoHash, peerID [20]byte) *Handshake {
	return &Handshake{
		Pstr:     "BitTorrent protocol",
		InfoHash: infoHash,
		PeerID:   peerID,
	}
}

func (h *Handshake) Serialize() []byte {
	buf := make([]byte, handshakeSize)
	buf[0] = 19
	copy(buf[1:20], []byte(h.Pstr))
	copy(buf[20:28], make([]byte, 8))
	copy(buf[28:48], h.InfoHash[:])
	copy(buf[48:68], h.PeerID[:])
	return buf
}

func DeserializeHandshake(r io.Reader) (*Handshake, error) {
	buf := make([]byte, handshakeSize)
	_, err := io.ReadFull(r, buf)
	if err != nil {
		return nil, err
	}

	if int(buf[0]) != 19 {
		err := errors.New("pstrlen must be equal 19")
		return nil, err
	}

	var infoHash, peerID [20]byte

	copy(infoHash[:], buf[28:48])
	copy(peerID[:], buf[48:])

	return &Handshake{
		Pstr:     string(buf[0:19]),
		InfoHash: infoHash,
		PeerID:   peerID,
	}, nil
}
