package message

import (
	"encoding/binary"
	"net"
)

type request struct {
	message
	index  int
	begin  int
	length int
}

func NewRequest(index, begin, length int) Message {
	return &request{
		index:  index,
		begin:  begin,
		length: length,
		message: message{
			id: MsgRequest,
		},
	}
}

func (r *request) Serialize() []byte {
	payload := make([]byte, 12)
	binary.BigEndian.PutUint32(payload[0:4], uint32(r.index))
	binary.BigEndian.PutUint32(payload[4:8], uint32(r.begin))
	binary.BigEndian.PutUint32(payload[8:12], uint32(r.length))
	r.message.payload = payload
	return r.message.Serialize()
}

func (r *request) Send(conn net.Conn) error {
	_, err := conn.Write(r.Serialize())
	return err
}
