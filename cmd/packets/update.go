package packets

import (
	"fmt"
	"net"

	"github.com/SotaUeda/gobgp/bgptype"
)

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
		l, err := pa.BytesLen()
		if err != nil {
			return nil, err
		}
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

func (u *UpdateMessage) ToBytes() ([]byte, error) {
	//TODO
	return nil, nil
}

func (u *UpdateMessage) ToMessage(b []byte) error {
	//TODO
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
