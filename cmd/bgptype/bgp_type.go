package bgptype

import "fmt"

type AutonomousSystemNumber uint16

type HoldTime uint16

func HoldTimeToUint16(ht HoldTime) uint16 {
	return uint16(ht)
}

func Uint16ToHoldTime(u uint16) HoldTime {
	return HoldTime(u)
}

func (ht *HoldTime) default_ht() {
	*ht = HoldTime(0)
}

func NewHoldTime() HoldTime {
	ht := HoldTime(0)
	ht.default_ht()
	return ht
}

type Version uint8

func VersionToUint8(v Version) uint8 {
	return uint8(v)
}

func Uint8ToVersion(u uint8) (Version, error) {
	vlim := uint8(4)
	if u <= vlim {
		v := Version(u)
		return v, nil
	}
	return 0, fmt.Errorf(
		"BGPのVersionは1-%dが期待されていますが、%dが渡されました",
		vlim, u,
	)
}

func (v *Version) default_v() {
	*v = 4
}

func NewVersion() Version {
	v := Version(4)
	v.default_v()
	return v
}
