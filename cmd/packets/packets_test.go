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
	err = newHeader.ToHeader(b)
	if err != nil {
		t.Errorf("Error: %v", err)
	}
	// フィールドを比較
	if header.length != newHeader.length || header.Type != newHeader.Type {
		t.Errorf("Want: %v, Got: %v", header, newHeader)
	}
}

// UpdateMessageのToMessageメソッドとToBytesメソッドをテストする
func TestConvertBytesToUpdateMessageAndUpdateMessageToBytes(t *testing.T) {
	someAs := bgptype.AutonomousSystemNumber(64513)
	// someIP := net.ParseIP("10.0.100.3").To4()

	localAs := bgptype.AutonomousSystemNumber(64514)
	localIP := net.ParseIP("10.200.100.3").To4()

	updateMsgPas := []*bgptype.PathAttribute{
		{
			Origin:  bgptype.IGP,
			AsPath:  &bgptype.AsPath{someAs, localAs},
			NextHop: localIP,
		},
	}

	rt := &net.IPNet{IP: net.ParseIP("10.100.220.0"), Mask: net.CIDRMask(24, 32)}
	updateMsg := NewUpdateMessage(
		updateMsgPas,
		[]*net.IPNet{rt},
		[]*net.IPNet{},
	)
	updateMsgByte, err := updateMsg.ToBytes()
	if err != nil {
		t.Errorf("Error: %v", err)
	}
	updateMsg2 := &UpdateMessage{}
	updateMsg2.ToMessage(updateMsgByte)
	// TODO: フィールドを比較
}
