package peer

import (
	"testing"
)

func TestPeerCanTransitionToConnectState(t *testing.T) {
	config, _ := Str2Config("64512 127.0.0.1 65413 127.0.0.2 active")
	peer := NewPeer(config)
	peer.Start()
	peer.Next()
	want := CONNECT
	if want != peer.State {
		t.Errorf("Want: %d,  Peer State: %d", want, peer.State)
	}
}
