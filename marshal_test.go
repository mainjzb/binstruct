package binstruct

import (
	"encoding/binary"
	"fmt"
	"log"
	"testing"
)

func Test_marshal_Marshal(t *testing.T) {
	type A struct {
		I int32 `bin:"len:3"`
		F float64
	}
	w := NewWriter(nil, nil, true)
	m := &marshal{w}

	a := A{-1, 2.2}
	err := m.Marshal(a)
	if err != nil {
		log.Println(err)
	}
	fmt.Println(w.Bytes())
	r := NewReaderFromBytes(w.Bytes(), binary.BigEndian, true)
	a2 := A{0, 2}
	err = r.Unmarshal(&a2)
	if err != nil {
		log.Println(err)
	}
	// json.Marshal()
}
