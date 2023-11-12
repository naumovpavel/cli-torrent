package message

import (
	"encoding/binary"
	"fmt"
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
		return fmt.Errorf("Payload too short. %d < 8", len(p.payload))
	}
	parsedIndex := int(binary.BigEndian.Uint32(p.payload[0:4]))
	if parsedIndex != p.index {
		return fmt.Errorf("Expected index %d, got %d", p.index, parsedIndex)
	}
	begin := int(binary.BigEndian.Uint32(p.payload[4:8]))
	if begin >= len(p.buf) {
		return fmt.Errorf("Begin offset too high. %d >= %d", begin, len(p.buf))
	}
	data := p.payload[8:]
	if begin+len(data) > len(p.buf) {
		return fmt.Errorf("Data too long [%d] for offset %d with length %d", len(data), begin, len(p.buf))
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
