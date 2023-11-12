package message

type interested struct {
	message
}

func NewInterested() *interested {
	return &interested{message{
		id: MsgInterested,
	}}
}
