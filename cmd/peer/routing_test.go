package peer

import (
	"net"
	"testing"
)

// LocRIBのLookupRoutingTableメソッドが正しく動作することを確認するテスト
// LookupRoutingTableメソッドは引数で指定されたネットワークアドレスに対応する
// ローカルのルーティングテーブル上のroute(*net.IPNet)のスライスを返す
func TestLocRIBCanLookupRoutingTable(t *testing.T) {
	// 本テストの値は環境によって異なる
	// 本実装では開発機、テスト実施機に
	// 10.200.100.0/24に属するIPが付与されていることを仮定している
	network := "10.200.100.0/24"
	_, dst, _ := net.ParseCIDR(network)
	rib := NewLocRIB()
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
