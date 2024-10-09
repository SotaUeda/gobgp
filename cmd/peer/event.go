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
	// MSGはMessageの省略形
	KEEPALIVE_MSG
	UPDATE_MSG
	// StateがEstablishedに遷移したことを表す
	// 存在する方が実装が楽なため追加したオリジナルイベント
	ESTABLISHED_STATE_EVENT
	// LocRib / AdjRibOut /AdjRibIn が変わったときのイベント
	// 存在する方が実装が楽なため追加
	LocRibChanged
	AdjRibOutChanged
	AdjRibInChanged
)

func (ev Event) Show() string {
	switch ev {
	case MANUAL_START:
		return "Manual Start"
	case TCP_CONNECTION_CONFIRMED:
		return "TCP Connection Confirmed"
	case BGP_OPEN:
		return "BGP Open"
	case KEEPALIVE_MSG:
		return "Recieved Keepalive Message"
	case ESTABLISHED_STATE_EVENT:
		return "Established"
	case LocRibChanged:
		return "LocRib Changed"
	case AdjRibOutChanged:
		return "AdjRibOut Changed"
	default:
		return fmt.Sprintf("%v", ev)
	}
}
