package packets

import (
	"fmt"
)

type Message interface {
	ToMessage([]byte) error
	ToBytes() ([]byte, error)
	Show() string
}

// Goでは、インターフェース型を返す関数で具体的な型のポインタを返すことができる
func BytesToMessage(b []byte) (Message, error) {
	h := &Header{}
	hErr := h.ToHeader(b[0:HEADER_LENGTH])
	if hErr != nil {
		return nil, hErr
	}
	var m Message
	switch h.Type {
	case Open:
		m = &OpenMessage{}
	case Keepalive:
		m = &KeepaliveMessage{}
	case Update:
		m = &UpdateMessage{}
	default:
		return nil, fmt.Errorf(
			"BytesからMessageに変換できませんでした。"+
				"未知のTypeです。Type: %d",
			h.Type,
		)
	}
	mErr := m.ToMessage(b)
	return m, mErr
}

func MessageToBytes(m Message) ([]byte, error) {
	b, err := m.ToBytes()
	if err != nil {
		return nil, fmt.Errorf(
			"MessageからBytesに変換できませんでした。%w",
			err,
		)
	}
	return b, nil
}
