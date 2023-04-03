package binstruct

import (
	"encoding/binary"
	"fmt"
	"log"
	"reflect"
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

	w := NewWriter(nil, true)
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

// 小端编码
type name struct {
	LinkCode         uint16 `bin:"len:2"` // 链路码
	SenderAdCode     uint32 `bin:"len:3"` // 发送方 行政规划码
	SenderType       uint16 `bin:"len:2"` // 发送方 类型
	SenderNumber     uint16 `bin:"len:2"` // 发送方 编号
	ReceiverAdCode   uint32 `bin:"len:3"` // 接收方 行政规划码
	ReceiveType      uint16 `bin:"len:2"` // 接收方 类型
	ReceiverNumber   uint16 `bin:"len:2"` // 接收方 编号
	TimeStamp        uint32 `bin:"len:4"` // 时间戳
	TimeStampReserve uint16 `bin:"len:2"` // 时间戳预留位置
	TTL              uint8  `bin:"len:1"` // 生存时间
	Version          uint8  `bin:"len:1"` // 协议版本
	Operation        uint8  `bin:"len:1"` // 操作类型
	ObjectName       uint8  `bin:"len:1"` // 对象名称编码
	ObjectType       uint8  `bin:"len:1"` // 对象类型
	Signature        uint8  `bin:"len:1"` // 签名 0:无签名 1:有签名
	Reserve          []byte `bin:"len:3"` // 保留 字段
	// Message
	LightsMessage CrossLight
	Crc           uint16 `bin:"len:2,be"` // CRC-16/MODBUS 大端
}

// CrossLight 灯色状态消息
type CrossLight struct {
	Length       uint16          `bin:"len:2"`                  // 消息长度
	Lon          float64         `bin:"len:4,Int32To10e6Float"` // 经度
	Lat          float64         `bin:"len:4,Int32To10e6Float"` // 纬度
	Height       uint16          `bin:"len:2"`                  // 海拔高度
	CrossInCount uint8           `bin:"len:1"`                  // 路口进口数量
	InLights     []EntranceLight `bin:"len:CrossInCount"`
}

// EntranceLight 进口灯色状态信息
type EntranceLight struct {
	InDir      uint16        // 进口方向
	LightCount uint8         `bin:"len:1"` // 进口灯组数量
	Status     []LightStatus `bin:"len:LightCount"`
}

// LightStatus 灯组灯色信息
type LightStatus struct {
	ID            uint8 // 灯组编号
	Type          uint8 // 灯组类型
	Color         uint8 // 灯组灯色
	RemainingTime uint8 // 剩余时间
}

func (cl CrossLight) Int32To10e6FloatDecode(r Reader) (float64, error) {
	v, err := r.ReadInt32()
	if err != nil {
		return 0, err
	}
	return float64(v) / 1000000, nil
}

func (cl CrossLight) Int32To10e6FloatEncode(w Writer, v float64) error {
	return w.WriteInt32(int32(v * 1000000))
}

func Test_marshal_Marshal2(t *testing.T) {
	data := []byte{0x05, 0x00, 0x00, 0x00, 0x00, 0x08, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x50, 0xff, 0xff, 0xa2, 0x12, 0xef, 0x60, 0x00, 0x00, 0xff, 0x10, 0x87, 0x03, 0x01, 0x00, 0x00, 0x00, 0x00, 0x21, 0x00, 0x00, 0x00, 0x00, 0x80, 0x00, 0x00, 0x00, 0x80, 0x00, 0x00, 0x02, 0x5a, 0x00, 0x02, 0x06, 0x00, 0x25, 0x19, 0x02, 0x00, 0x25, 0x19, 0x0e, 0x01, 0x02, 0x08, 0x00, 0x25, 0x19, 0x04, 0x00, 0x25, 0x19, 0x27, 0x99}

	var actual name
	err := UnmarshalLE(data, &actual) // UnmarshalLE() or Unmarshal()
	if err != nil {
		t.Fatal("UnmarshalLE err:" + err.Error())
	}

	result, err := MarshalLE(actual)
	if err != nil {
		t.Fatal("MarshalLE err: " + err.Error())
	}

	if !reflect.DeepEqual(result, data) {
		t.Fatal("result is not equal")
	}
}
