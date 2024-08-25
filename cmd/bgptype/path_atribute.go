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

// TODO:
// GetTypeメソッドを追加して処理を共通化した方がいい？
func PathAttributeToBytes(pa PathAttribute) ([]byte, error) {
	bytes := make([]byte, 0)
	// PathAtributeのBytes表現は以下の通り
	// [Attribute Flag (1 octet)]
	// [Attribute Type Code (1 octet)]
	// [Attribute Length (1 or 2 octets)]
	// [Attribute Value (Attribute Lengthのoctet数)]
	//
	// Atribute Flagは以下のBytes表現
	//   - 1bit目: AttributeがOptionalなら1, Well-knownなら0
	//   - 2bit目: Transitive(他ピアに伝える)なら1, そうでないなら0
	//     (補足: ただしWell-knownのものはすべてTransitive)
	//   - 3bit目: Partialなら1, completeなら0
	//     (補足: Well-knownならすべてcomplete)
	//   - 4bit目: Attribute Lengthがone octetなら0, two octetsなら1
	//   - 5-8bit目: 使用しない。ゼロ
	switch pa.(type) {
	case *Origin:
		attF := byte(0b01000000)
		attTC := byte(1)
		attL := []byte{byte(0), byte(1)}
		attV, err := pa.ToBytes()
		if err != nil {
			return nil, err
		}
		bytes = append(bytes, attF, attTC)
		bytes = append(bytes, attL...)
		bytes = append(bytes, attV...)
	case AsPath:
		attF := byte(0b01000000)
		attTC := byte(2)
		len, err := pa.BytesLen()
		if err != nil {
			return nil, err
		}
		var attL []byte
		if len < 256 {
			attL = []byte{byte(0), byte(len)}
		} else {
			attF += 0b00010000 // Attribute Lengthがtwo octetsなので4bit目を1にする
			attL = []byte{byte(len >> 8), byte(len)}
		}
		attV, err := pa.ToBytes()
		if err != nil {
			return nil, err
		}
		bytes = append(bytes, attF, attTC)
		bytes = append(bytes, attL...)
		bytes = append(bytes, attV...)
	case *NextHop:
		attF := byte(0b01000000)
		attTC := byte(3)
		attL := []byte{byte(0), byte(4)}
		attV, err := pa.ToBytes()
		if err != nil {
			return nil, err
		}
		bytes = append(bytes, attF, attTC)
		bytes = append(bytes, attL...)
		bytes = append(bytes, attV...)
	case *DontKnow:
		attV, err := pa.ToBytes()
		if err != nil {
			return nil, err
		}
		bytes = append(bytes, attV...)
	}
	return bytes, nil
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

// AS_PATHは
// Path Segment Type, Path Segment Length, Path Segment Valueの
// 3つから構成される可変長のデータ
// Path Segment Typeは1 octetのデータで、AS Pathを
// 	順番に意味のない集合で扱う場合1に、	-- AsSet
// 	順番に意味のあるシーケンスで扱う場合2に	-- AsSequence
// 設定する。
// Path Segment Lengthは1 octetのデータで、AS Pathの数を表す整数である。
// Path Segment Valueは可変長のデータを保持しており、
// それぞれ1つのAS Pathは2オクテットずつのデータで表される

// AsPathはAsSequenceとAsSetの2つの型を持つ
// AsSequenceはASを順番に並べたもの
// AsSetは重複のないASの集合(集約用途)
type AsPath interface {
	BytesLen() (uint16, error)
	ToBytes() ([]byte, error)
	Add() error
	Get() ([]AutonomousSystemNumber, error)
}

type AsSequence []AutonomousSystemNumber

func (seq *AsSequence) BytesLen() (uint16, error) {
	asBytesLen := 2 * len(*seq)
	// Segment Typeを表すoctet +  Path Segment Lengthを表すoctet + ASのbytesの値
	return 1 + 1 + uint16(asBytesLen), nil
}

// Segment Type, Path Segment Length, Path Segment Value
// 3つから構成される
func (seq *AsSequence) ToBytes() ([]byte, error) {
	bytes := make([]byte, 0)
	// Segment Type
	st := byte(2)
	// Segment Length
	sl := byte(len(*seq))
	bytes = append(bytes, st, sl)
	// Segment Value
	for _, as := range *seq {
		bytes = append(bytes, byte(as>>8), byte(as))
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
	// Segment Typeを表すoctet +  Path Segment Lengthを表すoctet + ASのbytesの値
	return 1 + 1 + uint16(asBytesLen), nil
}

// Segment Type, Path Segment Length, Path Segment Value
// 3つから構成される
func (set *AsSet) ToBytes() ([]byte, error) {
	bytes := make([]byte, 0)
	// Segment Type
	st := byte(1)
	// Segment Length
	sl := byte(len(*set))
	bytes = append(bytes, st, sl)
	// Segment Value
	for as := range *set {
		bytes = append(bytes, byte(as>>8), byte(as))
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
