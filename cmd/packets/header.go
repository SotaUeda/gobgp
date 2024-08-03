package packets

import "fmt"

type MessageType uint8

const (
	Open MessageType = iota
)

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
	h.Type = MessageType(b[18])

	return nil
}

// TODO: ToBytesメソッドを実装する
