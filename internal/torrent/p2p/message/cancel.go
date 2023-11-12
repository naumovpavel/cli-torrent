package message

type cancel struct {
	message
}

func NewCancel() Message {
	return &cancel{message{
		id: MsgCancel,
	}}
}
