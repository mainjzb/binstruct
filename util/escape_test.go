package utils

import (
	"reflect"
	"testing"
)

func TestEC(t *testing.T) {
	rule := map[byte][2]byte{
		0x02: {0x1B, 0xE7},
		0x03: {0x1B, 0xE8},
		0x1B: {0x1B, 0x00},
	}

	tests := []struct {
		name string
		rule map[byte][2]byte
		data []byte
		want []byte
	}{
		{
			"test1",
			rule,
			[]byte{0x02, 0x1B, 0x5B, 0x32, 0x4F, 0x31, 0x0D, 0x03},
			[]byte{0x1B, 0xE7, 0x1B, 0x00, 0x5B, 0x32, 0x4F, 0x31, 0x0D, 0x1B, 0xE8},
		},
		{
			"test2",
			map[byte][2]byte{
				0xDB: {0xDB, 0xDD},
				0xC0: {0xDB, 0xDC},
			},
			[]byte{},
			[]byte{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ec := NewEC(nil, nil, rule)
			got := ec.Escape(tt.data)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("escape() = %v, want %v", got, tt.want)
			}
			got2, err := ec.Unescape(got)
			if err != nil || !reflect.DeepEqual(got2, tt.data) {
				t.Errorf("Unescape() = %v, want %v", got2, tt.data)
			}
		})
	}
}
