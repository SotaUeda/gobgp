package peer

import (
	"net"
	"sync"

	"github.com/SotaUeda/gobgp/bgptype"
	"github.com/vishvananda/netlink"
)

type LocRib struct {
	Rib        *Rib
	LocalASNum bgptype.AutonomousSystemNumber
}

func NewLocRib(c *Config) (*LocRib, error) {
	igp := bgptype.IGP
	// AS Pathは、ほかのピアから受信したルートと統一的に扱うために、
	// LocRib -> AdjRibOutにルートを送るときに、自分のAS番号を
	// 追加するので、ここでは空にしておく。
	seq := bgptype.AsSequence{}
	nh := bgptype.NextHop(c.LocalIP)
	pas := []bgptype.PathAttribute{
		&igp,
		&seq,
		&nh,
	}

	rib := NewRib()
	locRib := &LocRib{}
	for _, nw := range c.Networks {
		rts, err := locRib.LookupRoutingTable(nw)
		if err != nil {
			return nil, err
		}
		if len(rts) == 0 {
			continue
		}
		for _, rt := range rts {
			rib.Insert(NewRibEntry(rt, pas...))
		}
	}
	locRib.Rib = rib
	locRib.LocalASNum = c.LocalAS
	return locRib, nil
}

// 各種Ribの処理の際、以前に処理したエントリは再処理する必要がない。
// その判別のためのステータス
type RibEntryStatus int

const (
	NEW_RIB_ENT RibEntryStatus = iota
	UN_CHANGED_RIB_ENT
)

// UpdateMessageでは、1つのPathAttributeに複数の
// NLRIを付けて送信する。このように
// PathAttributeは複数のルートに対して同じことがあるが、
// その全てでCloneをすることは避けたいため、
// 参考書ではArc<Vec<PathAttribute>>にしているが、
// ここでは sync.Mutex を使って排他制御を行うことで
// この問題を解決する。 <- これでいいのか？
type RibEntry struct {
	mu             sync.Mutex
	NwAddr         *net.IPNet
	pathAttributes []bgptype.PathAttribute // 排他制御のため、ローカル変数にする
}

func NewRibEntry(nw *net.IPNet, pas ...bgptype.PathAttribute) *RibEntry {
	return &RibEntry{
		NwAddr:         nw,
		pathAttributes: pas,
	}
}

func (re *RibEntry) AddPathAttributes(pas ...bgptype.PathAttribute) {
	re.mu.Lock()
	defer re.mu.Unlock()
	re.pathAttributes = append(re.pathAttributes, pas...)
}

func (re *RibEntry) GetPathAttributes() []bgptype.PathAttribute {
	re.mu.Lock()
	defer re.mu.Unlock()
	return re.pathAttributes
}

func (re *RibEntry) containAS(as bgptype.AutonomousSystemNumber) bool {
	for _, pa := range re.pathAttributes {
		switch t := pa.(type) {
		case *bgptype.AsSequence:
			return t.Contains(as)
		case *bgptype.AsSet:
			return t.Contains(as)
		}
	}
	return false
}

// AdjRibIn / LocRib / AdjRibOut で同じようなデータ構造・処理を持つため、
// 共通の処理はRib構造体に実装し、これらの3つの構造体のメンバにはRib構造体を持たせる。
//
// RibEntryは、3つのRibを渡りながら処理される。
// Rib間で受け渡すときにCloneを避けたいため、参考書ではHashMapのKeyを
// Arc<RibEntry>にしている。
// ここでは、sync.Mutexを使って排他制御を行うことでこの問題を解決する。
type Rib struct {
	mu      sync.Mutex
	entries map[*RibEntry]RibEntryStatus
}

func NewRib() *Rib {
	return &Rib{
		entries: make(map[*RibEntry]RibEntryStatus),
	}
}

// Rib内にentryが存在しなければInsert
func (rib *Rib) Insert(re *RibEntry) {
	rib.mu.Lock()
	defer rib.mu.Unlock()
	if _, ok := rib.entries[re]; !ok {
		rib.entries[re] = NEW_RIB_ENT
	}
}

func (rib *Rib) Routes() []*RibEntry {
	rib.mu.Lock()
	defer rib.mu.Unlock()
	rts := []*RibEntry{}
	for rt := range rib.entries {
		rts = append(rts, rt)
	}
	return rts
}

// AdjRibOut
type AdjRibOut struct {
	Rib *Rib
}

func NewAdjRibOut() *AdjRibOut {
	return &AdjRibOut{Rib: NewRib()}
}

func (aro *AdjRibOut) Insert(re *RibEntry) {
	aro.Rib.Insert(re)
}

// LocRibから必要なルートをインストールする
// この時、Rremote AS番号が含まれているルートはインストールしない。
func (aro *AdjRibOut) InstallFromLocRib(locRib *LocRib, config *Config) {
	rts := locRib.Rib.Routes()
	for _, rt := range rts {
		if rt.containAS(config.RemoteAS) {
			continue
		}
		// ここでAdjRibOutにルートをインストールする
		aro.Insert(rt)
	}
}

func (rib *LocRib) LookupRoutingTable(dst *net.IPNet) ([]*net.IPNet, error) {
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
	return dsts, nil
}
