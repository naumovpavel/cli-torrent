package message

type unchoke struct {
	message
}

func NewUnchoke() Message {
	return &unchoke{message: message{
		id: MsgUnchoke,
	}}
}
