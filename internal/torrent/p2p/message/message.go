package message

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
)

type MessageID int8

const (
	MsgChoke         MessageID = 0
	MsgUnchoke       MessageID = 1
	MsgInterested    MessageID = 2
	MsgNotInterested MessageID = 3
	MsgHave          MessageID = 4
	MsgBitfield      MessageID = 5
	MsgRequest       MessageID = 6
	MsgPiece         MessageID = 7
	MsgCancel        MessageID = 8
	KeepAlive        MessageID = 9
)

var names = map[MessageID]string{
	MsgChoke:         "Choke",
	MsgUnchoke:       "Unchoke",
	MsgInterested:    "Interested",
	MsgNotInterested: "Not interested",
	MsgHave:          "Have",
	MsgBitfield:      "Bitfield",
	MsgRequest:       "Request",
	MsgPiece:         "Piece",
	MsgCancel:        "Cancel",
}

type Message interface {
	Serialize() []byte
	Deserialize(payload []byte) error
	Send(conn net.Conn) error
	Payload() []byte
	Id() MessageID
	String() string
	name() string
}

type message struct {
	id      MessageID
	payload []byte
}

func NewMessage(id MessageID) Message {
	return &message{id: id}
}

func (m *message) Payload() []byte {
	return m.payload
}

func (m *message) Id() MessageID {
	return m.id
}

func (m *message) Serialize() []byte {
	if m == nil {
		return make([]byte, 4)
	}
	length := uint32(len(m.payload) + 1)
	buf := make([]byte, 4+length)
	binary.BigEndian.PutUint32(buf[0:4], length)
	buf[4] = byte(m.id)
	copy(buf[5:], m.payload)
	return buf
}

func (m *message) Deserialize(payload []byte) error {
	m.payload = payload
	return nil
}

func Read(r io.Reader) (MessageID, []byte, error) {
	lengthBuf := make([]byte, 4)
	_, err := io.ReadFull(r, lengthBuf)
	if err != nil {
		return -1, make([]byte, 0), err
	}
	length := binary.BigEndian.Uint32(lengthBuf)
	if length == 0 {
		return KeepAlive, make([]byte, 0), nil
	}
	messageBuf := make([]byte, length)
	_, err = io.ReadFull(r, messageBuf)
	if err != nil {
		return -1, make([]byte, 0), err
	}

	return MessageID(messageBuf[0]), messageBuf[1:], nil
}

func (m *message) Send(conn net.Conn) error {
	_, err := conn.Write(m.Serialize())
	return err
}

func (m *message) name() string {
	if m == nil {
		return "KeepAlive"
	}

	name, ok := names[m.id]
	if !ok {
		return "Unknown"
	} else {
		return name
	}
}

func (m *message) String() string {
	if m == nil {
		return m.name()
	}
	return fmt.Sprintf("%s [%d]", m.name(), len(m.payload))
}
