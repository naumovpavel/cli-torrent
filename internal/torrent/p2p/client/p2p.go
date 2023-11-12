package client

import (
	"bytes"
	"errors"
	"log"
	"net"
	"time"

	"cli-torrent/internal/torrent/p2p/message"
	"cli-torrent/internal/torrent/p2p/tracker"
	"cli-torrent/internal/torrent/torrentfile"
)

type P2PClient struct {
	conn       net.Conn
	chocked    bool
	peer       *tracker.Peer
	peerId     [20]byte
	Bitfield   message.Bitfield
	tf         *torrentfile.Torrentfile
	downloaded int
	backlog    int
}

func (c *P2PClient) readMessage(buf []byte, pieceIndex int) error {
	msgId, payload, err := message.Read(c.conn)

	if err != nil {
		//log.Println(err, " read err")
		return err
	}

	switch msgId {
	case message.MsgUnchoke:
		c.chocked = false
	case message.MsgChoke:
		c.chocked = true
	case message.MsgHave:
		index := 0
		err = message.NewHave(&index).Deserialize(payload)
		if err != nil {
			return err
		}
		c.Bitfield.SetPiece(index)
	case message.MsgPiece:
		pieceMsg := message.NewPiece(pieceIndex, buf)
		err = pieceMsg.Deserialize(payload)
		if err != nil {
			return err
		}
		c.downloaded += pieceMsg.Length()
		c.backlog--
	}

	return nil
}

func NewP2PClient(peer *tracker.Peer, peerId [20]byte, tf *torrentfile.Torrentfile) (*P2PClient, error) {
	conn, err := net.DialTimeout("tcp", peer.String(), 5*time.Second)

	if err != nil {
		return &P2PClient{}, err
	}

	err = doHandshake(conn, peerId, tf.Info.InfoHash)
	if err != nil {
		conn.Close()
		return &P2PClient{}, err
	}

	bitfield, err := receiveBitfield(conn)

	if err != nil {
		conn.Close()
		return &P2PClient{}, err
	}

	log.Println("connected to peer ", peer.String())

	return &P2PClient{
		conn:     conn,
		chocked:  true,
		peer:     peer,
		Bitfield: bitfield,
		peerId:   peerId,
		tf:       tf,
	}, nil
}

func receiveBitfield(conn net.Conn) (message.Bitfield, error) {
	conn.SetDeadline(time.Now().Add(3 * time.Second))
	defer conn.SetDeadline(time.Time{})
	msgId, payload, err := message.Read(conn)
	if err != nil {
		return make(message.Bitfield, 0), err
	}
	if msgId != message.MsgBitfield {
		return make(message.Bitfield, 0), errors.New("expected bitfield")
	}

	return payload, nil
}

func doHandshake(conn net.Conn, peerId, infoHash [20]byte) error {
	conn.SetDeadline(time.Now().Add(3 * time.Second))
	defer conn.SetDeadline(time.Time{})

	_, err := conn.Write(message.NewHandshake(
		infoHash,
		peerId,
	).Serialize())
	if err != nil {
		return err
	}

	resp, err := message.DeserializeHandshake(conn)
	if err != nil {
		return err
	}
	if !bytes.Equal(resp.InfoHash[:], infoHash[:]) {
		return errors.New("peers hasn't file with needed info hash")
	}
	return nil
}
