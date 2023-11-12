package message

import (
	"encoding/binary"
	"errors"
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

var (
	ErrUnexpectedMessage = errors.New("peer respond with unexpected message")
	ErrBadMessage        = errors.New("peer respond with incorrect message format")
)

type Message interface {
	Serialize() []byte
	Deserialize(payload []byte) error
	Send(conn net.Conn) error
	Payload() []byte
	Id() MessageID
}

type message struct {
	id      MessageID
	payload []byte
}

func (m *message) Payload() []byte {
	return m.payload
}

func (m *message) Id() MessageID {
	return m.id
}

func (m *message) Serialize() []byte {
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
		return -1, make([]byte, 0), ErrBadMessage
	}
	length := binary.BigEndian.Uint32(lengthBuf)
	if length == 0 {
		return KeepAlive, make([]byte, 0), ErrBadMessage
	}
	messageBuf := make([]byte, length)
	_, err = io.ReadFull(r, messageBuf)
	if err != nil {
		return -1, make([]byte, 0), ErrBadMessage
	}

	return MessageID(messageBuf[0]), messageBuf[1:], nil
}

func (m *message) Send(conn net.Conn) error {
	_, err := conn.Write(m.Serialize())
	return err
}
