package tracker

import (
	"encoding/binary"
	"net"
	"strconv"
)

type Peer struct {
	IP   net.IP
	Port uint16
}

func (p *Peer) Deserialize(bytes [6]byte) {
	p.IP = net.IP(bytes[0:4])
	p.Port = binary.BigEndian.Uint16(bytes[4:6])
}

func (p *Peer) String() string {
	return net.JoinHostPort(p.IP.String(), strconv.Itoa(int(p.Port)))
}
