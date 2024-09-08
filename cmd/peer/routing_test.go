package peer

import (
	"fmt"
	"net"
	"testing"

	"github.com/SotaUeda/gobgp/bgptype"
)

// LocRibのLookupRoutingTableメソッドが正しく動作することを確認するテスト
// LookupRoutingTableメソッドは引数で指定されたネットワークアドレスに対応する
// ローカルのルーティングテーブル上のroute(*net.IPNet)のスライスを返す
func TestLocRibCanLookupRoutingTable(t *testing.T) {
	// 本テストの値は環境によって異なる
	// 本実装では開発機、テスト実施機に
	// 10.200.100.0/24に属するIPが付与されていることを仮定している
	network := "10.200.100.0/24"
	_, dst, _ := net.ParseCIDR(network)
	rib := &LocRib{}
	routes, err := rib.LookupRoutingTable(dst)
	if err != nil {
		t.Errorf("Route not found")
	}
	if len(routes) == 0 {
		t.Errorf("Route not found")
	}
	want := dst.String()
	for _, route := range routes {
		result := route.String()
		if want == result {
			return
		}
	}
	t.Errorf("Route not found")
}

// AdjRibOutへルートをインストールする機能のテスト
func TestLocRibToAdjRibOut(t *testing.T) {
	// 本テストの値は環境によって異なる
	// 本実装では開発機、テスト実施機に
	// 10.200.100.0/24に属するIPが付与されていることを仮定している
	// docker-composeした環境のhost2で実行することを仮定している
	config, _ := ParseConfig(
		"64513 10.200.100.3 64512 10.200.100.2 passive 10.100.220.0/24",
	)
	lr, err := NewLocRib(config)
	if err != nil {
		t.Errorf("Error: %v", err)
	}
	adjRibOut := NewAdjRibOut()
	adjRibOut.InstallFromLocRib(lr, config)

	expected_adjRibOut := NewAdjRibOut()

	nw := &net.IPNet{
		IP:   net.ParseIP("10.100.220.0").To4(),
		Mask: net.CIDRMask(24, 32),
	}
	originIGP := bgptype.IGP
	nh := bgptype.NextHop(net.ParseIP("10.200.100.3").To4())
	re := NewRibEntry(
		nw,
		&originIGP,
		&bgptype.AsSequence{},
		&nh,
	)
	expected_adjRibOut.Insert(re)

	// AdjRibを比較するための関数を定義
	if !diffAdjRibOut(adjRibOut, expected_adjRibOut) {
		t.Errorf("AdjRibOut is not correct")
	}
}

func diffAdjRibOut(a, e *AdjRibOut) bool {
	if len(a.Rib.entries) != len(e.Rib.entries) {
		return false
	}
	var aps, eps string
	for are, ast := range a.Rib.entries {
		for _, pa := range are.GetPathAttributes() {
			aps += fmt.Sprintf("%v", pa.ToBytes())
		}
		for ere, est := range e.Rib.entries {
			if are.NwAddr.String() != ere.NwAddr.String() {
				return false
			}
			if ast != est {
				return false
			}
			for _, pa := range ere.GetPathAttributes() {
				eps += fmt.Sprintf("%v", pa.ToBytes())
			}
			if aps != eps {
				return false
			}
		}
	}
	return true
}
