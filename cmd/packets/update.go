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
	PathAttributes                      []*bgptype.PathAttribute
	pathAttributeLen                    uint16 // bytesにしたときのオクテット数
	NetworkLayerReachabilityInformation []*net.IPNet
	// NLRIのオクテット数はBGP UpdateMessageに含めず、
	// Headerのサイズを計算することにしか使用しないため
	// メンバに含めていない。
}

func NewUpdateMessage(
	pa []*bgptype.PathAttribute,
	nlri []*net.IPNet,
	wr []*net.IPNet) *UpdateMessage {
	//TODO
	return &UpdateMessage{}
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
