package peer

import (
	"context"
	"testing"
	"time"
)

func TestPeerCanTransitionToConnectState(t *testing.T) {
	config, _ := ParseConfig("64512 127.0.0.1 65413 127.0.0.2 active")
	peer := NewPeer(config)
	peer.Start()
	ctx := context.Background()
	go func() {
		remote_config, _ := ParseConfig("64512 127.0.0.2 65412 127.0.0.1 passive")
		remote_peer := NewPeer(remote_config)
		remote_peer.Start()
		remote_peer.Next(ctx)
	}()
	// remote_peer側の処理が進むことを保証するためのwait
	time.Sleep(1 * time.Second)
	peer.Next(ctx)
	want := CONNECT
	if want != peer.State {
		t.Errorf("Want: %d,  Peer State: %d", want, peer.State)
	}
	ctx.Done()
}
