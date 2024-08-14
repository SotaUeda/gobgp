package peer

import (
	"context"
	"testing"
	"time"
)

func TestPeerCanTransitionToConnectState(t *testing.T) {
	config, _ := ParseConfig("64512 127.0.0.1 64513 127.0.0.2 active")
	peer := NewPeer(config)
	peer.Start()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		remote_config, _ := ParseConfig("64513 127.0.0.2 64512 127.0.0.1 passive")
		remote_peer := NewPeer(remote_config)
		remote_peer.Start()
		remote_peer.Next(ctx)
	}()
	// remote_peer側の処理が進むことを保証するためのwait
	time.Sleep(1 * time.Second)
	peer.Next(ctx)
	peer.TCPConn.conn.Close()
	want := CONNECT
	if want != peer.State {
		t.Errorf("Want: %d,  Peer State: %d", want, peer.State)
	}
}

func TestPeerCanTransitionToOpenSentState(t *testing.T) {
	config, _ := ParseConfig("64512 127.0.0.3 64513 127.0.0.4 active")
	peer := NewPeer(config)
	peer.Start()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		remote_config, _ := ParseConfig("64513 127.0.0.4 64512 127.0.0.3 passive")
		remote_peer := NewPeer(remote_config)
		remote_peer.Start()
		remote_peer.Next(ctx)
		remote_peer.Next(ctx) // イベントをenqueueできていない？
	}()
	// remote_peer側の処理が進むことを保証するためのwait
	time.Sleep(1 * time.Second)
	peer.Next(ctx)
	peer.Next(ctx)
	peer.TCPConn.conn.Close()
	want := OPEN_SENT
	if want != peer.State {
		t.Errorf("Want: %d,  Peer State: %d", want, peer.State)
	}
}

func TestPeerCanTransitionToOpenConfirmState(t *testing.T) {
	config, _ := ParseConfig("64512 127.0.0.5 64513 127.0.0.6 active")
	peer := NewPeer(config)
	peer.Start()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		remote_config, _ := ParseConfig("64513 127.0.0.6 64512 127.0.0.5 passive")
		remote_peer := NewPeer(remote_config)
		remote_peer.Start()
		maxStep := 50
		for i := 0; i < maxStep; i++ {
			remote_peer.Next(ctx)
			if remote_peer.State == OPEN_CONFIRM {
				break
			}
			time.Sleep(1 * time.Millisecond)
		}
	}()
	// remote_peer側の処理が進むことを保証するためのwait
	time.Sleep(1 * time.Second)
	maxStep := 50
	for i := 0; i < maxStep; i++ {
		peer.Next(ctx)
		if peer.State == OPEN_CONFIRM {
			break
		}
		time.Sleep(1 * time.Millisecond)
	}
	peer.TCPConn.conn.Close()
	want := OPEN_CONFIRM
	if want != peer.State {
		t.Errorf("Want: %d,  Peer State: %d", want, peer.State)
	}
}

func TestPeerCanTransitionToEstablishedState(t *testing.T) {
	config, _ := ParseConfig("64512 127.0.0.7 64513 127.0.0.8 active")
	peer := NewPeer(config)
	peer.Start()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		remote_config, _ := ParseConfig("64513 127.0.0.8 64512 127.0.0.9 passive")
		remote_peer := NewPeer(remote_config)
		remote_peer.Start()
		maxStep := 50
		for i := 0; i < maxStep; i++ {
			remote_peer.Next(ctx)
			if remote_peer.State == ESTABLISHED {
				break
			}
			time.Sleep(1 * time.Millisecond)
		}
	}()
	// remote_peer側の処理が進むことを保証するためのwait
	time.Sleep(1 * time.Second)
	maxStep := 50
	for i := 0; i < maxStep; i++ {
		peer.Next(ctx)
		if peer.State == ESTABLISHED {
			break
		}
		time.Sleep(1 * time.Millisecond)
	}
	peer.TCPConn.conn.Close()
	want := ESTABLISHED
	if want != peer.State {
		t.Errorf("Want: %d,  Peer State: %d", want, peer.State)
	}
}
