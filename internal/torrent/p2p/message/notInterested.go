package message

type notInterested struct {
	message
}

func NewNotInterested() Message {
	return &notInterested{message{
		id: MsgInterested,
	}}
}
