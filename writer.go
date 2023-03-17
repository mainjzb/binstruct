package binstruct

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"math"
)

var (
	ErrCantWriter = errors.New("write size error")
)

type Writer interface {
	io.Writer

	// Peek returns the next n bytes without advancing the reader.
	// Peek(n int) ([]byte, error)

	WriteByte(c byte) error
	WriteUint8(v uint8) error
	WriteUint16(v uint16) error
	WriteUint32(v uint32) error
	WriteUint64(v uint64) error
	WriteUintX(v uint64, x int) error
	WriteInt8(v int8) error
	WriteInt16(v int16) error
	WriteInt32(v int32) error
	WriteInt64(v int64) error
	WriteIntX(v int64, x int) error

	WriteFloat32(v float32) error
	WriteFloat64(v float64) error

	Bytes() []byte

	// Marshal parses the binary data and stores the result
	// in the value pointed to by v.
	Marshal(v any) ([]byte, error)

	// WithOrder changes the byte order for the new Reader
	WithOrder(order binary.ByteOrder) Writer
}

func NewWriterWithBuffer(buffer *bytes.Buffer, order binary.ByteOrder, debug bool) Writer {
	if order == nil {
		order = binary.BigEndian
	}
	return &writer{
		buffer: buffer,
		order:  order,
		debug:  debug,
	}
}

func NewWriter(order binary.ByteOrder, debug bool) Writer {
	if order == nil {
		order = binary.BigEndian
	}
	return &writer{
		buffer: bytes.NewBuffer(make([]byte, 0, 1024)),
		order:  order,
		debug:  debug,
	}
}

type writer struct {
	buffer *bytes.Buffer
	order  binary.ByteOrder

	debug bool
}

func (w *writer) Write(p []byte) (n int, err error) {
	return w.buffer.Write(p)
}

func (w *writer) WriteByte(c byte) error {
	return w.buffer.WriteByte(c)
}

func (w *writer) WriteUintX(v uint64, x int) error {
	if x > 8 {
		errors.New("cannot write more than 8 bytes for custom length (u)int")
	}

	switch w.order {
	case binary.BigEndian:
		for i := 0; i < x; i++ {
			err := w.WriteByte(byte(v >> (8 * (x - i - 1))))
			if err != nil {
				return err
			}
		}
	case binary.LittleEndian:
		for i := 0; i < x; i++ {
			err := w.WriteByte(byte(v >> (8 * i)))
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (w *writer) WriteUint8(v uint8) error {
	return w.buffer.WriteByte(v)
}

func (w *writer) WriteUint16(v uint16) error {
	b := make([]byte, 2)
	w.order.PutUint16(b, v)
	err := w.buffer.WriteByte(b[0])
	if err != nil {
		return err
	}
	return w.buffer.WriteByte(b[1])
}

func (w *writer) WriteUint32(v uint32) error {
	b := make([]byte, 4)
	w.order.PutUint32(b, v)
	n, err := w.buffer.Write(b)
	if err != nil {
		return err
	}
	if n != 4 {
		return ErrCantWriter
	}
	return nil
}

func (w *writer) WriteUint64(v uint64) error {
	b := make([]byte, 8)
	w.order.PutUint64(b, v)
	n, err := w.buffer.Write(b)
	if err != nil {
		return err
	}
	if n != 4 {
		return ErrCantWriter
	}
	return nil
}

func (w *writer) WriteInt8(v int8) error {
	return w.WriteUint8(uint8(v))
}

func (w *writer) WriteInt16(v int16) error {
	return w.WriteUint16(uint16(v))
}

func (w *writer) WriteInt32(v int32) error {
	return w.WriteUint32(uint32(v))
}

func (w *writer) WriteInt64(v int64) error {
	return w.WriteUint64(uint64(v))
}

func (w *writer) WriteIntX(v int64, x int) error {
	if x > 8 {
		return errors.New("cannot write more than 8 bytes for custom length (u)int")
	}

	v = v << (64 - 8*x)
	u := uint64(v) >> (64 - 8*x)

	return w.WriteUintX(u, x)
}

func (w *writer) WriteFloat32(v float32) error {
	u := math.Float32bits(v)
	return w.WriteUint32(u)
}

func (w *writer) WriteFloat64(v float64) error {
	u := math.Float64bits(v)
	return w.WriteUint64(u)
}

func (w *writer) Bytes() []byte {
	return w.buffer.Bytes()
}

func (w *writer) Marshal(v any) ([]byte, error) {
	m := &marshal{w: w}
	return m.Marshal(v)
}

func (w *writer) WithOrder(order binary.ByteOrder) Writer {
	return NewWriterWithBuffer(w.buffer, order, w.debug)
}
