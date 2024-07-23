package peer

import (
	"context"
	"fmt"
)

// BFPのRFCで示されている実装方針
// (https://datatracker.ietf.org/doc/html/rfc4271#section-8)では、
// 1つのPeerを1つのイベント駆動ステートマシンとして実装しています。
// Peer構造体はRFC内で示されている実装方針に従ったイベント駆動ステートマシンです。
type Peer struct {
	State      State
	EventQueue chan Event
	Config     Config
}

func NewPeer(conf *Config) *Peer {
	p := &Peer{
		State:      IDLE,
		EventQueue: make(chan Event),
	}
	return p
}

func (p *Peer) Start() {
	fmt.Print("peer is started.\n")
	// channel は受信した場合でも送信されるまで処理が止まる
	// goroutin で呼び出す必要がある
	go start(p)
}

func start(p *Peer) {
	p.EventQueue <- MANUAL_START
}

func (p *Peer) Next(ctx context.Context) error {
	select {
	case ev := <-p.EventQueue:
		fmt.Printf("event is occured, event=%v.\n", ev.Show())
		p.handleEvent(ev)
		return nil
	case <-ctx.Done():
		fmt.Print("func next is done")
		return nil
	}
}

func (p *Peer) handleEvent(ev Event) {
	switch p.State {
	case IDLE:
		switch ev {
		case MANUAL_START:
			p.State = CONNECT
		}
	}
}
