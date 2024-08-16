package bgptype

import (
	"net"
)

type PathAttribute struct {
	Origin   Origin
	AsPath   AsPath
	NextHop  net.IP
	DontKnow []byte // 対応していないPathAtribute用
}

type Origin int

const (
	IGP Origin = iota
	EGP
	INCOMPLETE
)

// TODO: BTreeの実装等は後回し
type AsPath interface {
	Sequence() AsPath
}

type AsSequence []AutonomousSystemNumber

func (as *AsSequence) Sequence() AsPath {
	return as
}

func NewAsSequence(number ...AutonomousSystemNumber) *AsSequence {
	as := AsSequence(number)
	return &as
}
