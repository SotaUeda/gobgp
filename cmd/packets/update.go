package packets

import (
	"fmt"
	"net"

	"github.com/SotaUeda/gobgp/bgptype"
)

// TODO: Routeは*net.IPNetを使用せず、自作の型を使用したほうが良いかもしれない
type UpdateMessage struct {
	Header                              Header
	WithdrawnRoutes                     []*net.IPNet
	withdrawnRouteLen                   uint16 // ルート数ではなく、bytesにしたときのオクテット数
	PathAttributes                      []bgptype.PathAttribute
	pathAttributeLen                    uint16 // bytesにしたときのオクテット数
	NetworkLayerReachabilityInformation []*net.IPNet
	// NLRIのオクテット数はBGP UpdateMessageに含めず、
	// Headerのサイズを計算することにしか使用しないため
	// メンバに含めていない。
}

func NewUpdateMessage(
	pas []bgptype.PathAttribute,
	nlri []*net.IPNet,
	wr []*net.IPNet) (*UpdateMessage, error) {
	paLen := uint16(0)
	for _, pa := range pas {
		l := pa.BytesLen()
		paLen += l
	}
	nlriLen := uint16(0)
	for _, n := range nlri {
		l, err := NetByteLen(n)
		if err != nil {
			return nil, err
		}
		nlriLen += l
	}
	wrLen := uint16(0)
	for _, w := range wr {
		l, err := NetByteLen(w)
		if err != nil {
			return nil, err
		}
		wrLen += l
	}
	hMinLen := uint16(19)
	h := NewHeader(
		// +4はpath_attribute_length(u16)と
		// withdrawn_routes_length(u16)のbytes表現分
		hMinLen+paLen+nlriLen+wrLen+4,
		Update,
	)
	return &UpdateMessage{
		Header:                              *h,
		WithdrawnRoutes:                     wr,
		withdrawnRouteLen:                   wrLen,
		PathAttributes:                      pas,
		pathAttributeLen:                    paLen,
		NetworkLayerReachabilityInformation: nlri,
	}, nil
}

func (u *UpdateMessage) Show() string {
	return fmt.Sprintf(
		"Header: %v, WithdrawnRoutes: %v, WithdrawnRoutesLen: %v, PathAttributes: %v, pathAttributeLen: %v, NLRI: %v",
		u.Header,
		u.WithdrawnRoutes,
		u.withdrawnRouteLen,
		u.PathAttributes,
		u.pathAttributeLen,
		u.NetworkLayerReachabilityInformation,
	)
}

// UpdateMassageを[]byteに変換する
func (u *UpdateMessage) ToBytes() ([]byte, error) {
	b := make([]byte, 0)
	// header
	h, err := u.Header.ToBytes()
	if err != nil {
		return nil, err
	}
	b = append(b, h...)
	// withdrawn_routes_length
	wrLen := make([]byte, 2)
	wrLen[0] = byte(u.withdrawnRouteLen >> 8) // 8ビット右シフト
	wrLen[1] = byte(u.withdrawnRouteLen)
	b = append(b, wrLen...)
	// withdrawn_routes
	for _, wr := range u.WithdrawnRoutes {
		wrBytes, err := IPNetToBytes(wr)
		if err != nil {
			return nil, err
		}
		b = append(b, wrBytes...)
	}
	// path_attribute_length
	paLen := make([]byte, 2)
	paLen[0] = byte(u.pathAttributeLen >> 8)
	paLen[1] = byte(u.pathAttributeLen)
	b = append(b, paLen...)
	// path_attributes
	for _, pa := range u.PathAttributes {
		paBytes := pa.ToBytes()
		b = append(b, paBytes...)
	}
	// NLRI
	for _, nlri := range u.NetworkLayerReachabilityInformation {
		nlriBytes, err := IPNetToBytes(nlri)
		if err != nil {
			return nil, err
		}
		b = append(b, nlriBytes...)
	}
	return b, nil
}

func (u *UpdateMessage) ToMessage(b []byte) error {
	// header
	h := Header{}
	if err := h.ToHeader(b[0:HEADER_LENGTH]); err != nil {
		return err
	}
	// WITHDRAWN ROUTES LENGTH
	wrLen := uint16(b[19])<<8 | uint16(b[20])
	// WITHDRAWN ROUTES
	wrEnd := 21 + wrLen
	wrBytes := b[21:wrEnd]
	wrs, err := BytesToIPNets(wrBytes)
	if err != nil {
		return err
	}
	// PATH ATTRIBUTE LENGTH
	paLen := uint16(b[wrEnd])<<8 | uint16(b[wrEnd+1])
	// PATH ATTRIBUTES
	paStart := wrEnd + 2
	paEnd := paStart + paLen
	paBytes := b[paStart:paEnd]
	pas, err := bgptype.BytesToPathAttributes(paBytes)
	if err != nil {
		return err
	}
	// NLRI
	nlriStart := paEnd
	nlriBytes := b[nlriStart:]
	nlris, err := BytesToIPNets(nlriBytes)
	if err != nil {
		return err
	}

	u.Header = h
	u.WithdrawnRoutes = wrs
	u.withdrawnRouteLen = wrLen
	u.PathAttributes = pas
	u.pathAttributeLen = paLen
	u.NetworkLayerReachabilityInformation = nlris
	return nil
}

// NetByteLenはプレフィックスからバイト長を返す
// 0 => 1
// 1~8 => 2
// 9~16 => 3
// 17~24 => 4
// 25~32 => 5
// TODO: routing.goと重複しているので、共通化する
func NetByteLen(n *net.IPNet) (uint16, error) {
	ones, _ := n.Mask.Size()
	switch {
	case ones == 0:
		return 1, nil
	case ones >= 1 && ones <= 8:
		return 2, nil
	case ones >= 9 && ones <= 16:
		return 3, nil
	case ones >= 17 && ones <= 24:
		return 4, nil
	case ones >= 25 && ones <= 32:
		return 5, nil
	default:
		return 0, fmt.Errorf("invalid prefix length")
	}
}

// net.IPNetを[]byteに変換する
// WithdrawnRoutes, NLRIのバイト列表現はPrefix長とネットワークアドレスの組み合わせ
// {Prefix長, ネットワークアドレス}
// 例:
//
//	192.168.0.0/16 => {16, 192, 168}
//
// TODO: テスト
func IPNetToBytes(n *net.IPNet) ([]byte, error) {
	// IPNetのIPはIPv4のみをサポート
	ip := n.IP.To4()
	if ip == nil {
		return nil, fmt.Errorf("invalid ip address")
	}
	// プレフィックス長を取得
	ones, _ := n.Mask.Size()
	// プレフィックス長のバイト表現
	prefixLen := byte(ones)
	// ネットワークアドレスのバイト表現
	// 4オクテットのIPアドレスから、ネットワークアドレス部分のみを取得
	byteNw := make([]byte, 0)
	for i := 0; i < 4; i++ {
		if n.Mask[i] == 0 {
			break
		}
		byteNw = append(byteNw, ip[i]&n.Mask[i])
	}
	return append([]byte{prefixLen}, byteNw...), nil
}

// []byteを[]*net.IPNetに変換する
// 可変長のバイト列から、複数のプレフィックス長とネットワークアドレスを取得する
func BytesToIPNets(b []byte) ([]*net.IPNet, error) {
	nets := make([]*net.IPNet, 0)
	for i := 0; i < len(b); {
		// プレフィックス長
		ones := int(b[i])
		// プレフィックス長のバイト数
		prefixLenBytes := (ones + 7) / 8
		// ネットワークアドレス
		nw := make([]byte, 4)
		for j := 0; j < prefixLenBytes; j++ {
			nw[j] = b[i+j+1]
		}
		// プレフィックス長とネットワークアドレスから、net.IPNetを作成
		_, ipNet, err := net.ParseCIDR(fmt.Sprintf("%d.%d.%d.%d/%d", nw[0], nw[1], nw[2], nw[3], ones))
		if err != nil {
			return nil, err
		}
		nets = append(nets, ipNet)
		i += prefixLenBytes + 1
	}
	return nets, nil
}
