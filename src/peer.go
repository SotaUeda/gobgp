package peer

import "fmt"

// BFPのRFCで示されている実装方針
// (https://datatracker.ietf.org/doc/html/rfc4271#section-8)では、
// 1つのPeerを1つのイベント駆動ステートマシンとして実装しています。
// Peer構造体はRFC内で示されている実装方針に従ったイベント駆動ステートマシンです。
type Peer struct {
	State      State
	EventQueue chan Event
	Config     Config
}

func NewPeer(conf Config) *Peer {
	p := new(Peer)
	p.State = IDLE
	p.EventQueue = make(chan Event)
	return p
}

func (p *Peer) Start() {
	fmt.Print("peer is started.")
	p.EventQueue <- MANUAL_START
}
