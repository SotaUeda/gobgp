package peer

import (
	"testing"
)

func TestPeerCanTransitionToConnectState(t *testing.T) {
	config := "64512 127.0.0.1 65413 127.0.0.2 active" // 参考書では".parse().unwrap()を行っているが、一旦文字列のまま
	peer := NewPeer(config)
	peer.start()
	peer.next().await
	want := CONNECT
	if want != peer.state {
		t.Errorf("Want: %q,  Peer State: %q", &want, peer.state)
	}
}
