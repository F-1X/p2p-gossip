package p2p



type Message struct {
	Payload any
	From string
}

type Handshake struct {
	ListenAddr string
}

type MessagePeerList struct {
	Peers []string
}

func NewMessage(from string, payload any) *Message {
	return &Message{
		From: from,
		Payload: payload,
	}
}