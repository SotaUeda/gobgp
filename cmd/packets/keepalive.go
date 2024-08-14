package packets

import "fmt"

type KeepaliveMessage struct {
	Header *Header
}

const KEEPALIVE_MESSAGE_LENGTH = HEADER_LENGTH

func NewKeepaliveMessage() *KeepaliveMessage {
	h := NewHeader(KEEPALIVE_MESSAGE_LENGTH, Keepalive)
	return &KeepaliveMessage{
		Header: h,
	}
}

func (m *KeepaliveMessage) Show() string {
	return fmt.Sprintf("Header: %v", m.Header)
}

func (m *KeepaliveMessage) ToMessage(b []byte) error {
	h := &Header{}
	err := h.ToHeader(b)
	if err != nil {
		return err
	}
	if h.Type != Keepalive {
		return fmt.Errorf("TypeがKeepaliveではありません。Type: %d", h.Type)
	}
	m.Header = h
	return nil
}

func (m *KeepaliveMessage) ToBytes() ([]byte, error) {
	return m.Header.ToBytes()
}
