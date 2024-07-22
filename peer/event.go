package peer

import "fmt"

// BGPのRFC内 8.1
// (https://datatracker.ietf.org/doc/html/rfc4271#section-8.1)で
// 定義されているEventを表す定数
type Event int

const (
	MANUAL_START Event = iota
)

func (ev Event) Show() string {
	switch ev {
	case MANUAL_START:
		return "Manual Start"
	default:
		return fmt.Sprintf("%v", ev)
	}
}
