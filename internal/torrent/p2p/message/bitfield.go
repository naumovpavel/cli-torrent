package message

type Bitfield []byte

func (b Bitfield) HasPiece(i int) bool {
	index := i / 8
	if index < 0 || index >= len(b) {
		return false
	}
	return ((b[i/8] >> (7 - i%8)) & 1) > 0
}

func (b Bitfield) SetPiece(i int) {
	b[i/8] |= (1 << (7 - i%8))
}
