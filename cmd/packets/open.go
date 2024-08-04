package packets

import (
	"fmt"
	"net"

	"github.com/SotaUeda/gobgp/bgptype"
)

type OpenMessage struct {
	Header        *Header
	Version       bgptype.Version
	MyAS          bgptype.AutonomousSystemNumber
	HoldTime      bgptype.HoldTime // 正常系のみ実装するので一旦実質的に使用しない
	BGPIdentifier net.IP

	// 使用しないが、相手から受信したときに一応保存しておくためにプロパティとして定義
	OptionalParameterLength uint8
	OptionalParameters      []byte
}

const OPEN_MESSAGE_LENGTH = 29 // 自発的にOpenMessageを送信する場合の固定の長さ

func NewOpenMessage(as bgptype.AutonomousSystemNumber, ip net.IP) *OpenMessage {
	h := NewHeader(OPEN_MESSAGE_LENGTH, Open)
	return &OpenMessage{
		Header:                  h,
		Version:                 bgptype.NewVersion(),
		MyAS:                    as,
		HoldTime:                bgptype.NewHoldTime(),
		BGPIdentifier:           ip.To4(),
		OptionalParameterLength: 0,
		OptionalParameters:      []byte{},
	}
}

func (m *OpenMessage) ToMessage(b []byte) error {
	if len(b) < OPEN_MESSAGE_LENGTH {
		return fmt.Errorf(
			"OpenMessageに変換できませんでした。Bytesの長さが最小の長さより短いです。最小: 29, Bytes: %d",
			len(b),
		)
	}
	h := &Header{}
	err := h.ToMessage(b[0:HEADER_LENGTH])
	if err != nil {
		return err
	}
	m.Header = h
	m.Version = bgptype.Version(b[19])
	m.MyAS = bgptype.AutonomousSystemNumber(
		uint16(b[20])<<8 | uint16(b[21]),
	)
	m.HoldTime = bgptype.HoldTime(
		uint16(b[22])<<8 | uint16(b[23]),
	)
	m.BGPIdentifier = net.IPv4(b[24], b[25], b[26], b[27])
	m.OptionalParameterLength = b[28]
	m.OptionalParameters = b[29:]

	return nil
}

func (m *OpenMessage) ToBytes() ([]byte, error) {
	if len(m.BGPIdentifier) != 4 {
		return nil, fmt.Errorf("BGPIdentifierはIPv4アドレスである必要があります。")
	}

	b := make([]byte, 29+len(m.OptionalParameters))
	hb, err := m.Header.ToBytes()
	if err != nil {
		return nil, err
	}
	copy(b[0:HEADER_LENGTH], hb)
	b[19] = byte(m.Version)
	b[20] = byte(m.MyAS >> 8)
	b[21] = byte(m.MyAS & 0xff)
	b[22] = byte(m.HoldTime >> 8)
	b[23] = byte(m.HoldTime & 0xff)
	copy(b[24:28], m.BGPIdentifier.To4())
	b[28] = m.OptionalParameterLength
	copy(b[29:], m.OptionalParameters)

	return b, nil
}
