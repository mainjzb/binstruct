package utils

import (
	"bytes"
	"fmt"
)

// 帧信息通常需要被转移，为了简化转移带来每次需要重新转移代码的情况
// 封装转义符如下

type EC struct {
	start        []byte
	end          []byte
	escapeChar   map[byte][2]byte
	unescapeChar map[[2]byte]byte
}

// NewEC 初始化转义规则
func NewEC(start []byte, end []byte, escape map[byte][2]byte) *EC {
	unescape := make(map[[2]byte]byte)
	for oldValue, newValue := range escape {
		unescape[newValue] = oldValue
	}

	return &EC{
		start:        start,
		end:          end,
		escapeChar:   escape,
		unescapeChar: unescape,
	}
}

// Escape 数据包转义
func (ec EC) Escape(data []byte) []byte {
	result := make([]byte, 0, len(data)*2)
	result = append(result, ec.start...)
	for _, c := range data {
		v, ok := ec.escapeChar[c]
		if ok {
			result = append(result, v[0], v[1])
		} else {
			result = append(result, c)
		}
	}
	result = append(result, ec.end...)
	return result
}

// Unescape 数据包反转义
func (ec EC) Unescape(data []byte) ([]byte, error) {
	result := make([]byte, 0, len(data))
	if !bytes.HasPrefix(data, ec.start) {
		return nil, fmt.Errorf("start Code is not %X", ec.start)
	}
	if !bytes.HasSuffix(data, ec.start) {
		return nil, fmt.Errorf("end Code is not %X", ec.end)
	}

	data = data[len(ec.start) : len(data)-len(ec.end)]
	for i := 0; i < len(data)-1; i++ {
		v := [2]byte{data[i], data[i+1]}
		c, ok := ec.unescapeChar[v]
		if ok {
			result = append(result, c)
			i++
		} else {
			result = append(result, data[i])
			if i == len(data)-2 {
				result = append(result, data[i+1])
			}
		}
	}

	return result, nil
}
