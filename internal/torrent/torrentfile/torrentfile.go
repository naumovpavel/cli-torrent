package torrentfile

import (
	"fmt"
	"net/url"
	"strconv"
)

type Torrentfile struct {
	Announce string
	Info     Info
}

type Info struct {
	PieceLength int
	PieceHashes [][20]byte
	Name        string
	Length      int
	InfoHash    [20]byte
}

func New(path string) (*Torrentfile, error) {
	const op = "torrentfile.fromFile"

	btf, err := fromFile(path)
	if err != nil {
		return &Torrentfile{}, fmt.Errorf("%s: %w", op, err)
	}

	pieceHashes, err := btf.Info.pieceHashes()
	if err != nil {
		return &Torrentfile{}, fmt.Errorf("%s: %w", op, err)
	}

	infoHash, err := btf.Info.calcInfoHash()
	if err != nil {
		return &Torrentfile{}, fmt.Errorf("%s: %w", op, err)
	}
	return &Torrentfile{
		Announce: btf.Announce,
		Info: Info{
			PieceLength: btf.Info.PieceLength,
			PieceHashes: pieceHashes,
			Name:        btf.Info.Name,
			Length:      btf.Info.Length,
			InfoHash:    infoHash,
		},
	}, nil
}

func (t *Torrentfile) BuildTrackerUrl(peerId [20]byte, port uint16) (string, error) {
	base, err := url.Parse(t.Announce)
	if err != nil {
		return "", err
	}
	params := url.Values{
		"info_hash":  []string{string(t.Info.InfoHash[:])},
		"peer_id":    []string{string(peerId[:])},
		"port":       []string{strconv.Itoa(int(port))},
		"uploaded":   []string{"0"},
		"downloaded": []string{"0"},
		"compact":    []string{"1"},
		"left":       []string{strconv.Itoa(t.Info.Length)},
	}
	base.RawQuery = params.Encode()
	return base.String(), nil
}

func (t *Torrentfile) CalculateBoundsForPiece(index int) (begin int, end int) {
	begin = index * t.Info.PieceLength
	end = begin + t.Info.PieceLength
	if end > t.Info.Length {
		end = t.Info.Length
	}
	return begin, end
}

func (t *Torrentfile) CalculatePieceSize(index int) int {
	begin, end := t.CalculateBoundsForPiece(index)
	return end - begin
}
