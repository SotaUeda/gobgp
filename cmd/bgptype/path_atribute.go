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
	BytesLen() uint16
	ToBytes() []byte
	ToPA([]byte) error
}

// PathAttributeのフォーマット
//
// Optional Bit (1 bit): Well-known Attribute(0), Optional Attribute(1)
// Transitive Bit (1 bit): Transitive(1), Non-Transitive(0) ※Well-known Attributeの場合は(1)
// Partial Bit (1 bit): Partial(1), Complete(0) ※Well-known Attributeの場合は(0)
// Extended Length Bit (1 bit): PathAttributeのオクテット数が1のとき(0), 2のとき(1)
// Reserved (4 bit): 用途はない。すべて0
// Attr Type Code (8 bit): Origin(1), AS_PATH(2), NEXT_HOP(3), その他(4-255). ここではOrigin, AS_PATH, NEXT_HOPのみ実装
// Attribute Length (8 or 16 bit): Attribute Valueのオクテット数を表す符号なし整数
// Attribute Value (variable): Attr Type Codeによって異なる

type Origin int

const (
	IGP Origin = iota
	EGP
	INCOMPLETE
)

func (o *Origin) BytesLen() uint16 {
	return 4
}

func (o *Origin) ToBytes() []byte {
	attF := byte(0b01000000)
	attTC := byte(1)
	attL := byte(1)
	attV := byte(*o)
	bytes := make([]byte, 0)
	bytes = append(
		bytes, attF, attTC, attL, attV,
	)
	return bytes
}

func (o *Origin) ToPA(b []byte) error {
	if len(b) != 1 {
		return fmt.Errorf("Origin Attribute Length is not 1")
	}
	switch b[0] {
	case 0:
		*o = IGP
	case 1:
		*o = EGP
	case 2:
		*o = INCOMPLETE
	default:
		return fmt.Errorf("Origin Attribute Value is invalid")
	}
	return nil
}

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
	BytesLen() uint16
	ToBytes() []byte
	ToPA([]byte) error
	Add(AutonomousSystemNumber) error
	Get() []AutonomousSystemNumber
}

func NewAsPath(isSeq bool, as ...AutonomousSystemNumber) AsPath {
	var ap AsPath
	if isSeq {
		ap = new(AsSequence)
	} else {
		ap = new(AsSet)
	}
	for _, a := range as {
		ap.Add(a)
	}
	return ap
}

type AsSequence []AutonomousSystemNumber

func (seq *AsSequence) BytesLen() uint16 {
	asBytesLen := 2 * len(*seq)
	// Segment Typeを表すoctet +  Path Segment Lengthを表すoctet + ASのbytesの値
	asBytesLen += 2
	if asBytesLen < 256 {
		return uint16(asBytesLen + 3)
	} else {
		return uint16(asBytesLen + 4)
	}
}

// Segment Type, Path Segment Length, Path Segment Value
// 3つから構成される
func (seq *AsSequence) ToBytes() []byte {
	attF := byte(0b01000000)
	attTC := byte(2)
	bLen := len(*seq) * 2
	bLen += 2 // Segment TypeとSegment Lengthの2オクテット
	var attL []byte
	if bLen < 256 {
		attL = []byte{byte(bLen)}
	} else {
		attF += 0b00010000 // Attribute Lengthがtwo octetsなので4bit目を1にする
		attL = []byte{byte(bLen >> 8), byte(bLen)}
	}
	attV := make([]byte, 0)
	// Segment Type
	st := byte(2)
	// Segment Length
	sl := byte(len(*seq))
	attV = append(attV, st, sl)
	// Segment Value
	for _, as := range *seq {
		attV = append(attV, byte(as>>8), byte(as))
	}
	bytes := make([]byte, 0)
	bytes = append(bytes, attF, attTC)
	bytes = append(bytes, attL...)
	bytes = append(bytes, attV...)
	return bytes
}

func (seq *AsSequence) ToPA(b []byte) error {
	if *seq != nil {
		return fmt.Errorf("AS Path Attribute is already set")
	}
	if len(b) < 2 {
		return fmt.Errorf("AS Path Attribute Length is too short")
	}

	// Segment Type
	st := b[0]
	if st != 2 {
		return fmt.Errorf("AS Path Attribute Segment Type is not 2")
	}
	// Segment Length
	sl := b[1]
	if len(b) < 2+int(sl)*2 {
		return fmt.Errorf("AS Path Attribute Length is too short")
	}

	for i := 0; i < int(sl); i++ {
		as := AutonomousSystemNumber(b[2+i*2])<<8 + AutonomousSystemNumber(b[2+i*2+1])
		err := seq.Add(as)
		if err != nil {
			return err
		}
	}
	return nil
}

func (seq *AsSequence) Add(as AutonomousSystemNumber) error {
	*seq = append(*seq, as)
	return nil
}

func (seq *AsSequence) Get() []AutonomousSystemNumber {
	return *seq
}

type AsSet map[AutonomousSystemNumber]struct{}

func (set *AsSet) BytesLen() uint16 {
	asBytesLen := 2 * len(*set)
	// Segment Typeを表すoctet +  Path Segment Lengthを表すoctet + ASのbytesの値
	asBytesLen += 2
	if asBytesLen < 256 {
		return uint16(asBytesLen + 3)
	} else {
		return uint16(asBytesLen + 4)
	}
}

// Segment Type, Path Segment Length, Path Segment Value
// 3つから構成される
func (set *AsSet) ToBytes() []byte {
	attF := byte(0b01000000)
	attTC := byte(1)
	bLen := 2 * len(*set)
	bLen += 2 // Segment TypeとSegment Lengthの2オクテット
	var attL []byte
	if bLen < 256 {
		attL = []byte{byte(bLen)}
	} else {
		attF += 0b00010000 // Attribute Lengthがtwo octetsなので4bit目を1にする
		attL = []byte{byte(bLen >> 8), byte(bLen)}
	}
	attV := make([]byte, 0)
	// Segment Type
	st := byte(1)
	// Segment Length
	sl := byte(len(*set))
	attV = append(attV, st, sl)
	// Segment Value
	for as := range *set {
		attV = append(attV, byte(as>>8), byte(as))
	}
	bytes := make([]byte, 0)
	bytes = append(bytes, attF, attTC)
	bytes = append(bytes, attL...)
	bytes = append(bytes, attV...)
	return bytes
}

func (set *AsSet) ToPA(b []byte) error {
	if *set != nil {
		return fmt.Errorf("AS Path Attribute is already set")
	}
	if len(b) < 2 {
		return fmt.Errorf("AS Path Attribute Length is too short")
	}

	// Segment Type
	st := b[0]
	if st != 1 {
		return fmt.Errorf("AS Path Attribute Segment Type is not 1")
	}
	// Segment Length
	sl := b[1]
	if len(b) < 2+int(sl)*2 {
		return fmt.Errorf("AS Path Attribute Length is too short")
	}
	for i := 0; i < int(sl); i++ {
		as := AutonomousSystemNumber(b[2+i*2])<<8 + AutonomousSystemNumber(b[2+i*2+1])
		err := set.Add(as)
		if err != nil {
			return err
		}
	}
	return nil
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

func (n *NextHop) BytesLen() uint16 {
	return 7
}

func (n *NextHop) ToBytes() []byte {
	attF := byte(0b01000000)
	attTC := byte(3)
	attL := byte(4)
	attV := net.IP(*n).To4()
	bytes := make([]byte, 0)
	bytes = append(bytes, attF, attTC, attL)
	bytes = append(bytes, attV...)
	return bytes
}

func (n *NextHop) ToPA(b []byte) error {
	if len(b) != 4 {
		return fmt.Errorf("NextHop Attribute Length is not 4")
	}
	*n = NextHop(net.IP(b))
	return nil
}

type DontKnow []byte // 対応していないPathAtribute用

func (d *DontKnow) BytesLen() uint16 {
	return uint16(len(*d))
}

func (d *DontKnow) ToBytes() []byte {
	attV := *d
	bytes := make([]byte, 0)
	bytes = append(bytes, attV...)
	return bytes
}

func (d *DontKnow) ToPA(b []byte) error {
	*d = b
	return nil
}

func BytesToPathAttributes(b []byte) ([]PathAttribute, error) {
	pas := make([]PathAttribute, 0)
	i := 0
	for len(b) > i {
		attF := b[i]
		attLenOct := ((attF & 0b00010000) >> 4) + 1
		attTC := b[i+1]
		var attLen uint16
		if attLenOct == 1 {
			attLen = uint16(b[i+2])
		} else {
			attLen = uint16(b[i+2])<<8 + uint16(b[i+3])
		}

		attStartIdx := i + 1 + int(attLenOct) + 1
		attEndIdx := attStartIdx + int(attLen)
		if len(b) < attEndIdx {
			return nil, fmt.Errorf("attribute Length is too short")
		}
		attV := b[attStartIdx:attEndIdx]
		switch attTC {
		case 1:
			o := new(Origin)
			err := o.ToPA(attV)
			if err != nil {
				return nil, err
			}
			pas = append(pas, o)
		case 2:
			if attV[0] == 1 {
				set := new(AsSet)
				err := set.ToPA(attV)
				if err != nil {
					return nil, err
				}
				pas = append(pas, set)
			} else {
				seq := new(AsSequence)
				err := seq.ToPA(attV)
				if err != nil {
					return nil, err
				}
				pas = append(pas, seq)
			}
		case 3:
			n := new(NextHop)
			err := n.ToPA(attV)
			if err != nil {
				return nil, err
			}
			pas = append(pas, n)
		default:
			d := DontKnow(b[i:attEndIdx])
			pas = append(pas, &d)
		}
		i = attEndIdx
	}
	return pas, nil
}
