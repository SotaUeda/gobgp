package bgptype

import (
	"fmt"
	"net"
)

// PathAttributeの種類
// Origin
// AsPathAttribute
// NextHop
// DontKnow	対応していないPathAtribute用
type PathAttribute interface {
	BytesLen() (uint16, error)
}

type Origin int

func (o *Origin) BytesLen() (uint16, error) {
	return 1, nil
}

const (
	IGP Origin = iota
	EGP
	INCOMPLETE
)

// AsPathはAsSequenceとAsSetの2つの型を持つ
// AsSequenceはASを順番に並べたもの
// AsSetは重複のないASの集合(集約用途)
type AsPath interface {
	BytesLen() uint16
	Add() error
	Get() ([]AutonomousSystemNumber, error)
}

type AsSequence []AutonomousSystemNumber

func (seq *AsSequence) BytesLen() (uint16, error) {
	asBytesLen := 2 * len(*seq)
	// AsSetかAsSequenceかを表すoctet + ASの数を表すoctet + ASのbytesの値
	return 1 + 1 + uint16(asBytesLen), nil
}

func (seq *AsSequence) Add(as AutonomousSystemNumber) error {
	*seq = append(*seq, as)
	return nil
}

func (seq *AsSequence) Get() []AutonomousSystemNumber {
	return *seq
}

type AsSet map[AutonomousSystemNumber]struct{}

func (set *AsSet) BytesLen() (uint16, error) {
	asBytesLen := 2 * len(*set)
	// AsSetかAsSequenceかを表すoctet + ASの数を表すoctet + ASのbytesの値
	return 1 + 1 + uint16(asBytesLen), nil
}

func (set *AsSet) Add(as AutonomousSystemNumber) error {
	if _, exists := (*set)[as]; exists {
		return fmt.Errorf("AS %d already exists", as)
	}
	(*set)[as] = struct{}{}
	return nil
}

func (set *AsSet) Get() []AutonomousSystemNumber {
	keys := make([]AutonomousSystemNumber, 0, len(*set))
	for key := range *set {
		keys = append(keys, key)
	}
	return keys
}

type NextHop net.IP

func (n *NextHop) BytesLen() (uint16, error) {
	return 4, nil
}

type DontKnow []byte // 対応していないPathAtribute用

func (d *DontKnow) BytesLen() (uint16, error) {
	return uint16(len(*d)), nil
}
