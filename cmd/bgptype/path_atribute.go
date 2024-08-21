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
//
// PathAtributeのBytes表現は関数として用意する
type PathAttribute interface {
	BytesLen() (uint16, error)
	ToBytes() ([]byte, error)
}

// PathAtributeのBytes表現は関数として用意する
func PathAttributeToBytes(pa PathAttribute) ([]byte, error) {
	// TODO
	return nil, nil
}

type Origin int

func (o *Origin) BytesLen() (uint16, error) {
	return 1, nil
}

func (o *Origin) ToBytes() ([]byte, error) {
	return []byte{byte(*o)}, nil
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
	ToBytes() ([]byte, error)
	Add() error
	Get() ([]AutonomousSystemNumber, error)
}

type AsSequence []AutonomousSystemNumber

func (seq *AsSequence) BytesLen() (uint16, error) {
	asBytesLen := 2 * len(*seq)
	// AsSetかAsSequenceかを表すoctet + ASの数を表すoctet + ASのbytesの値
	return 1 + 1 + uint16(asBytesLen), nil
}

func (seq *AsSequence) ToBytes() ([]byte, error) {
	bytes := make([]byte, 0, 2*len(*seq))
	for _, as := range *seq {
		bytes = append(bytes, uint8(as>>8), uint8(as))
	}
	return bytes, nil
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

func (set *AsSet) ToBytes() ([]byte, error) {
	bytes := make([]byte, 0, 2*len(*set))
	for as := range *set {
		bytes = append(bytes, uint8(as>>8), uint8(as))
	}
	return bytes, nil
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

func (n *NextHop) ToBytes() ([]byte, error) {
	return net.IP(*n).To4(), nil
}

type DontKnow []byte // 対応していないPathAtribute用

func (d *DontKnow) BytesLen() (uint16, error) {
	return uint16(len(*d)), nil
}

func (d *DontKnow) ToBytes() ([]byte, error) {
	return *d, nil
}
