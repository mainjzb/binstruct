package binstruct

import (
	"errors"
	"reflect"
)

type InnerFunction int
type handler func(w Writer, structValue, fieldValue reflect.Value, fieldData *fieldReadData, parentStructValues []reflect.Value) error

const (
	Length InnerFunction = iota
	LengthWithoutSelf
)

var innerFunctionName = map[InnerFunction]string{
	Length:            "Length",
	LengthWithoutSelf: "LengthWithoutSelf",
}

var innerFunctionHandler = map[InnerFunction]handler{
	Length:            LengthHandler,
	LengthWithoutSelf: LengthWithoutSelfHandler,
}

func (i InnerFunction) String() string {
	return innerFunctionName[i]
}

func IsInnerFunction(funcName string) bool {
	for _, n := range innerFunctionName {
		if n == funcName {
			return true
		}
	}
	return false
}

func InnerFunctionHandler(w Writer, structValue, fieldValue reflect.Value, fieldData *fieldReadData, parentStructValues []reflect.Value) error {
	for innerFunc, n := range innerFunctionName {
		if n == fieldData.FuncName {
			return innerFunctionHandler[innerFunc](w, structValue, fieldValue, fieldData, parentStructValues)
		}
	}
	return errors.New("not fond function")
}

func LengthHandler(w Writer, structValue, fieldValue reflect.Value, fieldData *fieldReadData, parentStructValues []reflect.Value) error {
	sum, err := calcLength(structValue.Interface(), append(parentStructValues))
	if err != nil {
		return err
	}
	if fieldData.Length != nil {
		// value, err = r.ReadIntX(int(*fieldData.Length))
		switch fieldValue.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			err = w.WriteIntX(int64(sum), int(*fieldData.Length))
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			err = w.WriteUintX(uint64(sum), int(*fieldData.Length))
		}
	} else {

		switch fieldValue.Kind() {
		case reflect.Int8:
			return w.WriteInt8(int8(sum))
		case reflect.Int16:
			return w.WriteInt16(int16(sum))
		case reflect.Int32:
			return w.WriteInt32(int32(sum))
		case reflect.Int64:
			return w.WriteInt64(int64(sum))
		case reflect.Uint8:
			return w.WriteUint8(uint8(sum))
		case reflect.Uint16:
			return w.WriteUint16(uint16(sum))
		case reflect.Uint32:
			return w.WriteUint32(uint32(sum))
		case reflect.Uint64:
			return w.WriteUint64(uint64(sum))
		default: // reflect.Int:
			return errors.New("need set tag with len or use int8/int16/int32/int64")
		}
	}
	return nil

}

func LengthWithoutSelfHandler(w Writer, structValue, fieldValue reflect.Value, fieldData *fieldReadData, parentStructValues []reflect.Value) error {
	sum, err := calcLength(structValue.Interface(), append(parentStructValues))
	if err != nil {
		return err
	}
	if fieldData.Length != nil {
		switch fieldValue.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			err = w.WriteIntX(int64(sum-int(*fieldData.Length)), int(*fieldData.Length))
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			err = w.WriteUintX(uint64(sum-int(*fieldData.Length)), int(*fieldData.Length))
		}
	} else {

		switch fieldValue.Kind() {
		case reflect.Int8:
			return w.WriteInt8(int8(sum - 1))
		case reflect.Int16:
			return w.WriteInt16(int16(sum - 2))
		case reflect.Int32:
			return w.WriteInt32(int32(sum - 4))
		case reflect.Int64:
			return w.WriteInt64(int64(sum - 8))
		case reflect.Uint8:
			return w.WriteUint8(uint8(sum - 1))
		case reflect.Uint16:
			return w.WriteUint16(uint16(sum - 2))
		case reflect.Uint32:
			return w.WriteUint32(uint32(sum - 4))
		case reflect.Uint64:
			return w.WriteUint64(uint64(sum - 8))
		default: // reflect.Int:
			return errors.New("need set tag with len or use int8/int16/int32/int64")
		}
	}
	return nil
}
