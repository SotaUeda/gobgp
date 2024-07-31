package packets

import (
	"fmt"
	"net"
)

type Message interface {
	ToBytes() ([]byte, error)
	FromBytes([]byte) error
}

func BytesToMessage(b []byte) (*Message, error) {
	hl := 19
	if len(b) < hl {
		return nil, fmt.Errorf(
			"BytesからMessageに変換できませんでした。"+
				"Bytesの長さが最小の長さより短いです。最小: %d, Bytes: %d",
			hl, len(b),
		)
	}
	h, err := Header.FromBytes(b[0:hl])
	if err != nil {
		return nil, fmt.Errorf(
			"BytesからMessageに変換できませんでした。"+
				"Headerの変換に失敗しました。%w",
			err,
		)
	}
	switch h.Type {
	case Open:
		m := &OpenMessage{}
		m.Header = h
		err := m.FromBytes(b[hl:])
		return m, err
	default:
		return nil, fmt.Errorf(
			"BytesからMessageに変換できませんでした。"+
				"未知のTypeです。Type: %d",
			h.Type,
		)
	}
}

func MessageToBytes(m *Message) ([]byte, error) {
	b, err := m.ToBytes()
	if err != nil {
		return nil, fmt.Errorf(
			"MessageからBytesに変換できませんでした。%w",
			err,
		)
	}
	return b, nil
}

func (m *OpenMessage) NewOpen(as AoutonomousSystem, ip net.IP) {
	m = NewOpenMessage(as, ip)
}
