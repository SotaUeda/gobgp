package peer

import (
	"context"
	"fmt"

	"github.com/SotaUeda/gobgp/packets"
)

// BFPのRFCで示されている実装方針
// (https://datatracker.ietf.org/doc/html/rfc4271#section-8)では、
// 1つのPeerを1つのイベント駆動ステートマシンとして実装しています。
// Peer構造体はRFC内で示されている実装方針に従ったイベント駆動ステートマシンです。
type Peer struct {
	State      State
	EventQueue chan Event
	TCPConn    *Connection
	Config     *Config
}

func NewPeer(conf *Config) *Peer {
	p := &Peer{
		State:      IDLE,
		EventQueue: make(chan Event),
		Config:     conf,
	}
	return p
}

func (p *Peer) Start() {
	fmt.Print("peer is started.\n")
	// channel は受信した場合でも送信されるまで処理が止まる
	// goroutin で呼び出す必要がある
	go func() { p.EventQueue <- MANUAL_START }()
}

func (p *Peer) Next(ctx context.Context) error {
	for {
		select {
		case ev := <-p.EventQueue:
			fmt.Printf("event is occured, event=%v.\n", ev.Show())
			p.handleEvent(ev)
			return nil
		case <-ctx.Done():
			fmt.Print("func next is done.\n")
			if p.TCPConn != nil {
				p.TCPConn.conn.Close()
				fmt.Print("close connection\n")
			}
			return nil
		default:
			if p.TCPConn != nil && p.State != CONNECT {
				m, err := p.TCPConn.Recv()
				if err != nil {
					return err
				}
				fmt.Printf("message is received, message=%v.\n", m.Show())
				p.handleMessage(m)
				return nil
			}
		}
	}
}

func (p *Peer) handleMessage(m packets.Message) {
	switch m.(type) {
	case *packets.OpenMessage:
		go func() { p.EventQueue <- BGP_OPEN }()
	case *packets.KeepaliveMessage:
		go func() { p.EventQueue <- KEEPALIVE_MSG }()
	}
}

func (p *Peer) handleEvent(ev Event) error {
	switch p.State {
	case IDLE:
		switch ev {
		case MANUAL_START:
			// 参考記事 https://qiita.com/tutuz/items/e875d8ea3c31450195a7
			conn, err := NewConnection(p.Config)
			if err != nil {
				return err
			}
			p.TCPConn = conn
			if p.TCPConn == nil {
				return fmt.Errorf("TCP Connectionが確立できませんでした")
			}
			p.State = CONNECT
			go func() { p.EventQueue <- TCP_CONNECTION_CONFIRMED }()
		}
	case CONNECT:
		switch ev {
		case TCP_CONNECTION_CONFIRMED:
			if p.TCPConn == nil {
				return fmt.Errorf("TCP Connectionが確立できていません")
			}
			err := p.TCPConn.Send(packets.NewOpenMessage(
				p.Config.LocalAS,
				p.Config.LocalIP,
			))
			if err != nil {
				return err
			}
			p.State = OPEN_SENT
		}
	case OPEN_SENT:
		switch ev {
		case BGP_OPEN:
			if p.TCPConn == nil {
				return fmt.Errorf("TCP Connectionが確立できていません")
			}
			err := p.TCPConn.Send(packets.NewKeepaliveMessage())
			if err != nil {
				return err
			}
			p.State = OPEN_CONFIRM
		}
	case OPEN_CONFIRM:
		switch ev {
		case KEEPALIVE_MSG:
			p.State = ESTABLISHED
			go func() { p.EventQueue <- ESTABLISHED_STATE_EVENT }()
		}
	}
	return nil
}
