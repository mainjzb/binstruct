package utils

import (
	"bytes"
	"fmt"
	"reflect"
)

// 帧信息通常需要被转移，为了简化转移带来每次需要重新转移代码的情况
// 自动添加帧头和帧尾和校验值
// 封装转义符如下

type EC struct {
	start        []byte
	end          []byte
	escapeChar   map[byte][2]byte
	unescapeChar map[[2]byte]byte
	checkLength  int                 // 校验值长度,根据checkFunc自动生成
	checkFunc    func([]byte) []byte // 校验函数
}

// NewEC 初始化转义规则
func NewEC(start, end []byte, escape map[byte][2]byte, checksum func([]byte) []byte) *EC {
	unescape := make(map[[2]byte]byte)
	for oldValue, newValue := range escape {
		unescape[newValue] = oldValue
	}

	checkLength := 0
	if checksum != nil {
		crc := checksum([]byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a})
		checkLength = len(crc)
	}

	return &EC{
		start:        start,
		end:          end,
		escapeChar:   escape,
		unescapeChar: unescape,
		checkLength:  checkLength,
		checkFunc:    checksum,
	}
}

// Escape 数据包转义
func (ec EC) Escape(data []byte) []byte {
	// escape content
	content := make([]byte, 0, len(data)*2)
	for _, c := range data {
		v, ok := ec.escapeChar[c]
		if ok {
			content = append(content, v[0], v[1])
		} else {
			content = append(content, c)
		}
	}

	result := make([]byte, 0, len(data)*2)
	// add start
	result = append(result, ec.start...)
	// add content
	result = append(result, content...)
	// add crc
	if ec.checkFunc != nil && ec.checkLength != 0 {
		result = append(result, ec.checkFunc(data)...)
	}
	// add end
	result = append(result, ec.end...)

	return result
}

// Unescape 数据包反转义
func (ec EC) Unescape(data []byte) ([]byte, error) {
	// 校验帧头帧尾
	if !bytes.HasPrefix(data, ec.start) {
		return nil, fmt.Errorf("start Code is not %X", ec.start)
	}
	if !bytes.HasSuffix(data, ec.end) {
		return nil, fmt.Errorf("end Code is not %X", ec.end)
	}

	// 去掉帧头帧尾
	content := data[len(ec.start) : len(data)-len(ec.end)]
	// 反转义
	result := make([]byte, 0, len(content))
	for i := 0; i < len(content); i++ {
		if i+1 == len(content) {
			result = append(result, content[i])
			break
		}

		if c, ok := ec.unescapeChar[[2]byte{content[i], content[i+1]}]; ok {
			result = append(result, c)
			i++
		} else {
			result = append(result, content[i])
		}
	}

	// 校验crc
	if !ec.checkContent(result) {
		return nil, fmt.Errorf("check crc err")
	}

	return result[:len(result)-ec.checkLength], nil
}

func (ec EC) checkContent(content []byte) bool {
	if ec.checkFunc == nil || ec.checkLength == 0 {
		return true
	}
	crc := content[len(content)-ec.checkLength:]
	content = content[:len(content)-ec.checkLength]
	calcCrc := ec.checkFunc(content)
	return reflect.DeepEqual(crc, calcCrc)
}
