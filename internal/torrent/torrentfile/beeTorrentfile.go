package torrentfile

import (
	"bytes"
	"crypto/sha1"
	"errors"
	"fmt"
	"os"

	"github.com/jackpal/bencode-go"
)

type beeTorrentfile struct {
	Announce string  `bencode:"announce"`
	Info     beeInfo `bencode:"info"`
}

type beeInfo struct {
	PieceLength int    `bencode:"piece length"`
	Pieces      string `bencode:"pieces"`
	Name        string `bencode:"name"`
	Length      int    `bencode:"length"`
}

var (
	ErrMalformedPieces = errors.New("pieces is malformed, size of each piece must be 20 bytes")
)

func fromFile(path string) (beeTorrentfile, error) {
	const op = "torrentfile.fromFile"

	file, err := os.Open(path)
	defer file.Close()
	if err != nil {
		return beeTorrentfile{}, fmt.Errorf("%s: %w", op, err)
	}

	btf := beeTorrentfile{}
	err = bencode.Unmarshal(file, &btf)

	return btf, nil
}

func (i *beeInfo) pieceHashes() ([][20]byte, error) {
	const pieceHashLen = 20

	buf := []byte(i.Pieces)
	if len(buf)%pieceHashLen != 0 {
		return make([][20]byte, 0), ErrMalformedPieces
	}

	cnt := len(buf) / pieceHashLen
	pieces := make([][pieceHashLen]byte, cnt)
	for i := 0; i < cnt; i++ {
		copy(pieces[i][:], buf[i*pieceHashLen:(i+1)*pieceHashLen])
	}
	return pieces, nil
}

func (i *beeInfo) calcInfoHash() ([20]byte, error) {
	var buf bytes.Buffer
	err := bencode.Marshal(&buf, *i)
	if err != nil {
		return [20]byte{}, err
	}
	return sha1.Sum(buf.Bytes()), nil
}
