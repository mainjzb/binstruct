package utils

import (
	"reflect"
	"testing"
)

func Test_CRC16XMODEM(t *testing.T) {
	tests := []struct {
		name string
		args []byte
		want [2]byte
	}{
		{"test1", []byte("123"), [2]byte{0x97, 0x52}},
		{"test2", []byte("123456789"), [2]byte{0x31, 0xC3}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GenCRC16XMODEM(tt.args); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("genCrc16XMODEM() = % X, want % X", got, tt.want)
			}
		})
	}
}

func Test_CRC16MODBUS(t *testing.T) {
	tests := []struct {
		name string
		args []byte
		want [2]byte
	}{
		{"test1", []byte("123"), [2]byte{0x7A, 0x75}},
		{"test2", []byte("123456789"), [2]byte{0x4B, 0x37}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GenCRC16MODBUS(tt.args); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("genCrc16XMODEM() = % X, want % X", got, tt.want)
			}
		})
	}
}
