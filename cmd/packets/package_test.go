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
	ip := net.ParseIP("127.0.0.1").To4()
	if ip == nil {
		t.Errorf("Error: %v", ip)
	}
	openMsg := NewOpenMessage(as, ip)
	b, err := openMsg.ToBytes()
	if err != nil {
		t.Errorf("Error: %v", err)
	}
	newOpenMsg := &OpenMessage{}
	err = newOpenMsg.ToMessage(b)
	if err != nil {
		t.Errorf("Error: %v", err)
	}
	// ヘッダのフィールドを比較
	if openMsg.Header.length != newOpenMsg.Header.length || openMsg.Header.Type != newOpenMsg.Header.Type {
		t.Errorf("Want: %v, Got: %v", openMsg.Header, newOpenMsg.Header)
	}
	// OpenMessageのフィールドを比較
	if openMsg.Version != newOpenMsg.Version ||
		openMsg.MyAS != newOpenMsg.MyAS ||
		openMsg.HoldTime != newOpenMsg.HoldTime ||
		!openMsg.BGPIdentifier.Equal(newOpenMsg.BGPIdentifier) {
		t.Errorf("Want: %v, Got: %v", openMsg, newOpenMsg)
	}
}

// HeaderのToMessageメソッドとToBytesメソッドをテストする
func TestConvertBytesToHeaderAndHeaderToBytes(t *testing.T) {
	header := NewHeader(19, Open)
	b, err := header.ToBytes()
	if err != nil {
		t.Errorf("Error: %v", err)
	}
	newHeader := &Header{}
	err = newHeader.ToMessage(b)
	if err != nil {
		t.Errorf("Error: %v", err)
	}
	// フィールドを比較
	if header.length != newHeader.length || header.Type != newHeader.Type {
		t.Errorf("Want: %v, Got: %v", header, newHeader)
	}
}
