package binstruct

import (
	"encoding/binary"
	"io"
)

// UnmarshalLE parses the binary data with little-endian byte order and
// stores the result in the value pointed to by v.
func UnmarshalLE(data []byte, v interface{}) error {
	return Unmarshal(data, binary.LittleEndian, v)
}

// UnmarshalBE parses the binary data with big-endian byte order and
// stores the result in the value pointed to by v.
func UnmarshalBE(data []byte, v interface{}) error {
	return Unmarshal(data, binary.BigEndian, v)
}

// Unmarshal parses the binary data with byte order and stores the result
// in the value pointed to by v. If v is nil or not a pointer,
// Unmarshal returns an InvalidUnmarshalError.
func Unmarshal(data []byte, order binary.ByteOrder, v interface{}) error {
	return NewReaderFromBytes(data, order, false).Unmarshal(v)
}

func MarshalLE(v interface{}) ([]byte, error) {
	return NewWriter(binary.LittleEndian, false).Marshal(v)
}

func MarshalBE(v interface{}) ([]byte, error) {
	return NewWriter(binary.BigEndian, false).Marshal(v)
}

func Marshal(order binary.ByteOrder, v interface{}) ([]byte, error) {
	return NewWriter(order, false).Marshal(v)
}

type Encoder struct {
	w     io.Writer
	order binary.ByteOrder
	debug bool
}

// A Decoder reads and decodes binary values from an input stream.
type Decoder struct {
	r     io.ReadSeeker
	order binary.ByteOrder
	debug bool
}

func NewEncoder(w io.Writer, order binary.ByteOrder) *Encoder {
	return &Encoder{
		w:     w,
		order: order,
		debug: false,
	}
}

// NewDecoder returns a new decoder that reads from r with byte order.
func NewDecoder(r io.ReadSeeker, order binary.ByteOrder) *Decoder {
	return &Decoder{
		r:     r,
		order: order,
		debug: false,
	}
}

// SetDebug if set true, all read bytes and offsets will be displayed.
func (dec *Encoder) SetDebug(debug bool) {
	dec.debug = debug
}

// SetDebug if set true, all read bytes and offsets will be displayed.
func (dec *Decoder) SetDebug(debug bool) {
	dec.debug = debug
}

// Decode reads the binary-encoded value from its
// input and stores it in the value pointed to by v.
func (dec *Decoder) Decode(v interface{}) error {
	return NewReader(dec.r, dec.order, dec.debug).Unmarshal(v)
}

// Decode reads the binary-encoded value from its
// input and stores it in the value pointed to by v.
func (dec *Encoder) Encode(v interface{}) ([]byte, error) {
	return NewWriter(dec.order, dec.debug).Marshal(v)
}
