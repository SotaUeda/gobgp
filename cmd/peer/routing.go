package peer

import (
	"fmt"
	"net"

	"github.com/vishvananda/netlink"
)

type LocRIB struct{}

func NewLocRIB() *LocRIB {
	return &LocRIB{}
}

func (rib *LocRIB) LookupRoutingTable(dst *net.IPNet) ([]*net.IPNet, error) {
	// ルーティングテーブルの取得
	routes, err := netlink.RouteList(nil, netlink.FAMILY_V4)
	if err != nil {
		return nil, err
	}
	dsts := []*net.IPNet{}
	for _, route := range routes {
		if route.Dst != nil && route.Dst.Contains(dst.IP) {
			dsts = append(dsts, route.Dst)
		}
	}
	if len(dsts) == 0 {
		return nil, fmt.Errorf("route not found")
	}
	return dsts, nil
}
