package binstruct

import (
	"encoding/binary"
	"fmt"
	"log"
	"reflect"
	"testing"
	"time"
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
	LightsMessage struct {
		Length       uint16          `bin:"len:2,LengthWithoutSelf"` // 消息长度 Length
		Lon          float64         `bin:"len:4,Int32To10e6Float"`  // 经度
		Lat          float64         `bin:"len:4,Int32To10e6Float"`  // 纬度
		Height       uint16          `bin:"len:2"`                   // 海拔高度
		CrossInCount uint8           `bin:"len:1"`                   // 路口进口数量
		InLights     []EntranceLight `bin:"len:CrossInCount"`
	}
	Crc uint16 `bin:"len:2,be"` // CRC-16/MODBUS 大端
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

func (cl name) Int32To10e6FloatDecode(r Reader) (float64, error) {
	v, err := r.ReadInt32()
	if err != nil {
		return 0, err
	}
	return float64(v) / 1000000, nil
}

func (cl name) Int32To10e6FloatEncode(w Writer, v float64) error {
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

// StatisticsData 统计数据
type StatisticsData struct {
	Len             uint16       `bin:"len:2,Length"`
	Time            time.Time    `bin:"len:7,Time"` // [7]byte
	CollectionCycle uint16       // 采集周期(秒)
	RoadCount       uint8        // M个车道
	Content         []RoadDetail `bin:"len:RoadCount"`
}
type RoadDetail struct {
	LaneID                uint8   // 车道号
	HeadTime              float64 `bin:"len:2,Uint16To10e1Float"` // 车头时距 0.1 s
	BodyTime              float64 `bin:"len:2,Uint16To10e1Float"` // 车身时距 0.1 s
	Speed85p              float64 `bin:"len:2,Uint16To10e1Float"` // 85%速度 0.1 km/h
	TimeOcc               float64 `bin:"len:2,Uint16To10e1Float"` // 时间占有率 0.1
	Car1Flow              uint16  // 车型流量  车型由小到大
	Car1Speed             float64 `bin:"len:2,Uint16To10e1Float"` // 车型速度 0.1 km/h
	Car1Occ               float64 `bin:"len:2,Uint16To10e1Float"` // 车型占有率 0.1
	Car2Flow              uint16  // 车型流量
	Car2Speed             float64 `bin:"len:2,Uint16To10e1Float"` // 车型速度 0.1 km/h
	Car2Occ               float64 `bin:"len:2,Uint16To10e1Float"` // 车型占有率  0.1
	Car3Flow              uint16  // 车型流量
	Car3Speed             float64 `bin:"len:2,Uint16To10e1Float"` // 车型速度 0.1 km/h
	Car3Occ               float64 `bin:"len:2,Uint16To10e1Float"` // 车型占有率  0.1
	MaxVehicleQueueLength uint16  // 最大排队长度 米
	MaxVehicleQueueCount  uint16  // 最大排队数量
	Reserved              [8]byte // 保留字段
}

func (sd StatisticsData) TimeDecode(r Reader) (time.Time, error) {
	_, data, err := r.ReadBytes(7)
	if err != nil {
		return time.Time{}, err
	}

	t := time.Date(int(data[0])+2000, time.Month(data[1]), int(data[2]), int(data[3]), int(data[4]), int(data[5]), 0, time.Local)

	return t, nil
}

func (sd StatisticsData) TimeEncode(w Writer, v time.Time) error {
	data := make([]byte, 7)
	data[0] = byte(v.Year() % 100)
	data[1] = byte(v.Month())
	data[2] = byte(v.Day())
	data[3] = byte(v.Hour())
	data[4] = byte(v.Minute())
	data[5] = byte(v.Second())
	data[6] = byte(v.Weekday())
	_, err := w.Write(data)
	return err
}

func (sd StatisticsData) Uint16To10e1FloatDecode(r Reader) (float64, error) {
	v, err := r.ReadUint16()
	if err != nil {
		return 0, err
	}
	return float64(v) / 10, nil
}

func (sd StatisticsData) Uint16To10e1FloatEncode(w Writer, v float64) error {
	return w.WriteUint16(uint16(v * 10))
}

func Test_StatisticsData(t *testing.T) {
	sd := StatisticsData{
		RoadCount: 2,
		Content:   make([]RoadDetail, 2),
	}
	buf, err := MarshalLE(sd)
	if err != nil {
		t.Error(err)
	}
	if buf[0] != 90 {
		t.Error(fmt.Printf("inner function Length err: want=90, got=%d ", buf[0]))
	}
}

type Packet struct {
	SendID       uint8
	RevID        uint8
	DetectorAddr uint8
	OperateType  uint8
	ContentID    uint8
	Content      []byte
	Crc          uint16
}

func Test_Packet(t *testing.T) {
	pt := Packet{
		RevID:       0x01,
		OperateType: 0x03,
		ContentID:   0x04,
		Content:     []byte{0x01, 0x02},
		Crc:         0x00,
	}
	data, err := MarshalLE(pt)
	if err != nil {
		t.Error(err)
	}
	_ = data
}
