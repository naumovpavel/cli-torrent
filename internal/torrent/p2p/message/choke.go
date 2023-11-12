package message

type choke struct {
	message
}

func NewChoke() Message {
	return &choke{message: message{
		id: MsgChoke,
	}}
}
