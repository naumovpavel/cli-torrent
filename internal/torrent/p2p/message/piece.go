package message

import (
	"encoding/binary"
)

type piece struct {
	message
	index  int
	buf    []byte
	length int
}

func NewPiece(index int, buf []byte) *piece {
	return &piece{
		index: index,
		buf:   buf,
		message: message{
			id: MsgPiece,
		},
	}
}

func (p *piece) Deserialize(buf []byte) error {
	p.payload = buf
	if len(p.payload) < 8 {
		return ErrBadMessage
	}
	parsedIndex := int(binary.BigEndian.Uint32(p.payload[0:4]))
	if parsedIndex != p.index {
		return ErrBadMessage
	}
	begin := int(binary.BigEndian.Uint32(p.payload[4:8]))
	if begin >= len(p.buf) {
		return ErrBadMessage
	}
	data := p.payload[8:]
	if begin+len(data) > len(p.buf) {
		return ErrBadMessage
	}
	copy(p.buf[begin:], data)
	p.length = len(data)
	return nil
}

func (p *piece) Buf() []byte {
	return p.buf
}

func (p *piece) Length() int {
	return p.length
}
