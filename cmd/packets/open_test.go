package packets

import (
	"net"
	"testing"

	"github.com/SotaUeda/gobgp/bgptype"
)

// OpenMessageのToMessageメソッドとToBytesメソッドをテストする
// OpenMassageインスタンスをNewOpenMessage関数で作成(AS: 64512、IP: 127.0.0.1)し、
// ToBytesメソッドでバイト列に変換し、ToMessageメソッドで元のOpenMessageインスタンスに戻す
// その後、元のOpenMessageインスタンスと戻ったOpenMessageインスタンスが等しいかを確認する
func TestConvertBytesToOpenMessageAndOpenMessageToBytes(t *testing.T) {
	as := bgptype.AutonomousSystemNumber(64512)
	ip := net.ParseIP("127.0.0.1")
	openMsg := NewOpenMessage(as, ip)
	b, err := openMsg.ToBytes()
	if err != nil {
		t.Errorf("Error: %v", err)
	}
	newOpenMsg, err := BytesToMessage(b)
	if err != nil {
		t.Errorf("Error: %v", err)
	}
	if openMsg != newOpenMsg {
		t.Errorf("Want: %v, Got: %v", openMsg, newOpenMsg)
	}
}
