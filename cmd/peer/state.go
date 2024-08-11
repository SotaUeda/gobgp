package peer

type State int

const (
	IDLE State = iota
	CONNECT
	OPEN_SENT
	OPEN_CONFIRM
)
