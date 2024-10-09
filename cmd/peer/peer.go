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
	// UpdateMessageを処理するため、強引にMessageを埋め込む
	Msg       packets.Message
	TCPConn   *Connection
	Config    *Config
	LocRib    *LocRib
	AdjRibOut *AdjRibOut
	AdjRibIn  *AdjRibIn
}

func NewPeer(conf *Config, locRib *LocRib) *Peer {
	p := &Peer{
		State:      IDLE,
		EventQueue: make(chan Event),
		Config:     conf,
		LocRib:     locRib,
		AdjRibOut:  NewAdjRibOut(locRib.Rib),
		AdjRibIn:   NewAdjRibIn(locRib.Rib),
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
	case *packets.UpdateMessage:
		go func() {
			p.Msg = m
			p.EventQueue <- UPDATE_MSG
		}()
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
	case ESTABLISHED:
		switch ev {
		case ESTABLISHED_STATE_EVENT, LocRibChanged:
			locRib := p.LocRib
			p.AdjRibOut.InstallFromLocRib(locRib, p.Config)
			if p.AdjRibOut.Rib.DoseContainNewRoute() {
				go func() { p.EventQueue <- AdjRibOutChanged }()
				p.AdjRibOut.Rib.UpsateToAllUnchanged()
			}
		case AdjRibOutChanged:
			ums, err := p.AdjRibOut.ToUpdateMessages(
				p.Config.LocalIP,
				p.Config.LocalAS,
			)
			if err != nil {
				return err
			}
			for _, um := range ums {
				if p.TCPConn == nil {
					return fmt.Errorf("TCP Connectionが確立できていません")
				}
				p.TCPConn.Send(um)
			}
		case UPDATE_MSG:
			if p.Msg == nil {
				return fmt.Errorf("UpdateMessageがありません")
			}
			um := p.Msg.(*packets.UpdateMessage)
			p.AdjRibIn.InstallFromUpdate(um, p.Config)
			if p.AdjRibIn.Rib.DoseContainNewRoute() {
				fmt.Println("adj_rib in is updated.")
				go func() { p.EventQueue <- AdjRibInChanged }()
				p.AdjRibIn.Rib.UpsateToAllUnchanged()
			}
		case AdjRibInChanged:
			p.LocRib.Rib.mu.Lock()
			defer p.LocRib.Rib.mu.Unlock()
			p.LocRib.InstallFromAdjRibIn(p.AdjRibIn)
			if p.LocRib.Rib.DoseContainNewRoute() {
				p.LocRib.WriteToKernelRoutingTable()
				go func() { p.EventQueue <- LocRibChanged }()
				p.LocRib.Rib.UpsateToAllUnchanged()
			}
		}
	}
	return nil
}
