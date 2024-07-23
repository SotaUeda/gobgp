package peer

import (
	"context"
	"testing"
)

func TestPeerCanTransitionToConnectState(t *testing.T) {
	config, _ := ParseConfig("64512 127.0.0.1 65413 127.0.0.2 active")
	peer := NewPeer(config)
	peer.Start()
	ctx := context.Background()
	peer.Next(ctx)
	want := CONNECT
	if want != peer.State {
		t.Errorf("Want: %d,  Peer State: %d", want, peer.State)
	}
}
