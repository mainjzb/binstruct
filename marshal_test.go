package binstruct

import (
	"encoding/binary"
	"fmt"
	"log"
	"testing"
)

type A struct {
	I int32   `bin:"len:3"`
	F float64 `bin:"Float64Int"`
}

func (a A) Float64IntDecode(r Reader) (float64, error) {
	v, err := r.ReadInt32()
	if err != nil {
		return 0, err
	}
	return float64(v) / 1000000, nil
}

func (a A) Float64IntEncode(w Writer, v float64) error {
	return w.WriteInt32(int32(v * 1000000))
}

func Test_marshal_Marshal(t *testing.T) {

	w := NewWriter(nil, nil, true)
	m := &marshal{w}

	a := A{-1, 2.2}
	b, err := m.Marshal(a)
	if err != nil {
		log.Println(err)
	}
	fmt.Println(b)
	r := NewReaderFromBytes(w.Bytes(), binary.BigEndian, true)
	a2 := A{0, 2}
	err = r.Unmarshal(&a2)
	if err != nil {
		log.Println(err)
	}
	// json.Marshal()
}
