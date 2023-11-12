package message

import (
	"encoding/binary"
	"net"
)

type have struct {
	message
	index *int
}

func NewHave(index *int) Message {
	return &have{
		index: index,
		message: message{
			id: MsgHave,
		},
	}
}

func (r *have) Serialize() []byte {
	r.message.payload = make([]byte, 4)
	binary.BigEndian.PutUint32(r.message.payload, uint32(*r.index))
	return r.message.Serialize()
}

func (r *have) Send(conn net.Conn) error {
	_, err := conn.Write(r.Serialize())
	return err
}

func (m *have) Deserialize(payload []byte) error {
	m.payload = payload
	if len(m.payload) != 4 {
		return ErrBadMessage
	}
	index := int(binary.BigEndian.Uint32(m.payload))
	*(m.index) = index
	return nil
}
