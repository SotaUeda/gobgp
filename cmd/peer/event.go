package peer

import "fmt"

// BGPのRFC内 8.1
// (https://datatracker.ietf.org/doc/html/rfc4271#section-8.1)で
// 定義されているEventを表す定数
type Event int

const (
	MANUAL_START Event = iota
	// 正常系しか実装しない本実装では別のEventとして扱う意味がないため、
	// TcpConnectionConfirmedはTcpAckedも兼ねている。
	TCP_CONNECTION_CONFIRMED
	BGP_OPEN
	KEEPALIVE
)

func (ev Event) Show() string {
	switch ev {
	case MANUAL_START:
		return "Manual Start"
	case TCP_CONNECTION_CONFIRMED:
		return "TCP Connection Confirmed"
	case BGP_OPEN:
		return "BGP Open"
	default:
		return fmt.Sprintf("%v", ev)
	}
}
