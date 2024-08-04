package packets

import "fmt"

const HEADER_LENGTH = 19

type Header struct {
	length uint16
	Type   MessageType
}

func NewHeader(length uint16, t MessageType) *Header {
	return &Header{
		length: length,
		Type:   t,
	}
}

func (h *Header) ToMessage(b []byte) error {
	if len(b) < HEADER_LENGTH {
		return fmt.Errorf(
			"Headerに変換できませんでした。Bytesの長さが最小の長さより短いです。最小: 19, Bytes: %d",
			len(b),
		)
	}
	// Merkerはすべて1のため無視する
	h.length = uint16(b[16])<<8 | uint16(b[17])
	var err error
	h.Type, err = BytesToMessageType(b[18])
	if err != nil {
		return err
	}

	return nil
}

func (h *Header) ToBytes() ([]byte, error) {
	b := make([]byte, HEADER_LENGTH)
	for i := 0; i < 16; i++ {
		b[i] = 0xff
	}
	b[16] = byte(h.length >> 8)
	b[17] = byte(h.length & 0xff)
	b[18] = byte(h.Type)
	return b, nil
}

type MessageType uint8

const (
	Open MessageType = iota + 1 // 1
)

func BytesToMessageType(b byte) (MessageType, error) {
	switch b {
	case 1:
		return Open, nil
	default:
		return 0, fmt.Errorf("未知のMessageTypeです。Type: %d", b)
	}
}
