package binstruct

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
)

type marshal struct {
	w Writer
}

func (m *marshal) Marshal(v any) ([]byte, error) {
	return m.marshal(v, nil)
}

func (m *marshal) marshal(v any, parentStructValues []reflect.Value) ([]byte, error) {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Struct {
		return nil, &InvalidUnmarshalError{reflect.TypeOf(v)}
	}

	fieldCount := rv.NumField()
	valueType := rv.Type()

	for i := 0; i < fieldCount; i++ {
		fieldType := valueType.Field(i)
		tags, err := parseTag(fieldType.Tag.Get(tagName))
		if err != nil {
			return nil, fmt.Errorf(`failed parseTag for field "%s": %w`, fieldType.Name, err)
		}

		fieldData, err := parseReadDataFromTags(rv, tags)
		if err != nil {
			return nil, fmt.Errorf(`failed parse ReadData from tags for field "%s": %w`, fieldType.Name, err)
		}
		fieldValue := rv.Field(i)
		err = m.setValueToField(rv, fieldValue, fieldData, parentStructValues)
		if err != nil {
			return nil, fmt.Errorf(`failed set value to field "%s": %w`, fieldType.Name, err)
		}
	}
	return m.w.Bytes(), nil
}

func (m *marshal) setValueToField(structValue, fieldValue reflect.Value, fieldData *fieldReadData, parentStructValues []reflect.Value) error {
	if fieldData == nil {
		fieldData = &fieldReadData{}
	}

	if fieldData.Ignore {
		return nil
	}

	w := m.w
	if fieldData.Order != nil {
		w = w.WithOrder(fieldData.Order)
	}

	var err error
	// err := setOffset(r, fieldData)
	if err != nil {
		return fmt.Errorf("set offset: %w", err)
	}

	if IsInnerFunction(fieldData.FuncName) {
		return InnerFunctionHandler(w, structValue, fieldValue, fieldData, parentStructValues)
	}

	if fieldData.FuncName != "" {
		var okCallFunc bool
		okCallFunc, err = callEncodeFunc(w, fieldData.FuncName, structValue, fieldValue)
		if err != nil {
			return fmt.Errorf("call custom func(%s): %w", structValue.Type().Name(), err)
		}

		if !okCallFunc {
			// Try call function from parent structs
			for i := len(parentStructValues) - 1; i >= 0; i-- {
				sv := parentStructValues[i]
				okCallFunc, err = callEncodeFunc(w, fieldData.FuncName, sv, fieldValue)
				if err != nil {
					return fmt.Errorf("call custom func from parent(%s): %w", sv.Type().Name(), err)
				}

				if okCallFunc {
					return nil
				}
			}

			message := `
		failed call method, expected methods:
			func (*{{Struct}}) {{MethodName}}(r binstruct.Reader) error {}
		or
			func (*{{Struct}}) {{MethodName}}(r binstruct.Reader) ({{FieldType}}, error) {}
		`
			message = strings.NewReplacer(
				`{{Struct}}`, structValue.Type().Name(),
				`{{MethodName}}`, fieldData.FuncName,
				`{{FieldType}}`, fieldValue.Type().String(),
			).Replace(message)
			return errors.New(message)
		}

		return nil
	}

	switch fieldValue.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		var value int64
		var err error

		if fieldData.Length != nil {
			// value, err = r.ReadIntX(int(*fieldData.Length))
			err = w.WriteIntX(fieldValue.Int(), int(*fieldData.Length))
		} else {
			switch fieldValue.Kind() {
			case reflect.Int8:
				e := w.WriteInt8(int8(fieldValue.Int()))
				err = e
			case reflect.Int16:
				e := w.WriteInt16(int16(fieldValue.Int()))
				err = e
			case reflect.Int32:
				e := w.WriteInt32(int32(fieldValue.Int()))
				err = e
			case reflect.Int64:
				e := w.WriteInt64(int64(fieldValue.Int()))
				err = e
			default: // reflect.Int:
				return errors.New("need set tag with len or use int8/int16/int32/int64")
			}
		}

		if err != nil {
			return err
		}

		if fieldValue.CanSet() {
			fieldValue.SetInt(value)
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		// var value uint64
		var err error

		if fieldData.Length != nil {
			// value, err = r.ReadUintX(int(*fieldData.Length))
			err = w.WriteUintX(fieldValue.Uint(), int(*fieldData.Length))
		} else {
			switch fieldValue.Kind() {
			case reflect.Uint8:
				e := w.WriteUint8(uint8(fieldValue.Uint()))
				err = e
			case reflect.Uint16:
				e := w.WriteUint16(uint16(fieldValue.Uint()))
				err = e
			case reflect.Uint32:
				e := w.WriteUint32(uint32(fieldValue.Uint()))
				err = e
			case reflect.Uint64:
				e := w.WriteUint64(uint64(fieldValue.Uint()))
				err = e
			default: // reflect.Uint:
				return errors.New("need set tag with len or use uint8/uint16/uint32/uint64")
			}
		}

		if err != nil {
			return err
		}

		// if fieldValue.CanSet() {
		// 	fieldValue.SetUint(value)
		// }
	case reflect.Float32:
		err := w.WriteFloat32(float32(fieldValue.Float()))
		if err != nil {
			return err
		}
	case reflect.Float64:
		err := w.WriteFloat64(fieldValue.Float())
		if err != nil {
			return err
		}
	case reflect.Bool:
		panic("don't func writeBool")
		// b, err := r.ReadBool()
		// if err != nil {
		// 	return err
		// }
	case reflect.String:
		if fieldData.Length == nil {
			return errors.New("need set tag with len for string")
		}
		_, err := w.Write([]byte(fieldValue.String()))
		if err != nil {
			return err
		}
	case reflect.Slice:
		if fieldData.Length == nil {
			return errors.New("need set tag with len for slice")
		}

		for i := int64(0); i < *fieldData.Length; i++ {
			err = m.setValueToField(structValue, fieldValue.Index(int(i)), fieldData.ElemFieldData, parentStructValues)
			if err != nil {
				return err
			}
		}
	case reflect.Array:
		var arrLen int64

		if fieldData.Length != nil {
			arrLen = *fieldData.Length
		}

		if arrLen == 0 {
			arrLen = int64(fieldValue.Len())
		}

		for i := int64(0); i < arrLen; i++ {
			err = m.setValueToField(structValue, fieldValue.Index(int(i)), fieldData.ElemFieldData, parentStructValues)
			if err != nil {
				return err
			}
		}
	case reflect.Struct:
		_, err := m.marshal(fieldValue.Interface(), append(parentStructValues, structValue))
		if err != nil {
			return fmt.Errorf("unmarshal struct: %w", err)
		}
	default:
		return errors.New(`type "` + fieldValue.Kind().String() + `" not supported`)
	}

	return nil
}

func callEncodeFunc(r Writer, funcName string, structValue, fieldValue reflect.Value) (bool, error) {
	// Call methods
	m := structValue.MethodByName(funcName + "Encode")

	writerType := reflect.TypeOf((*Writer)(nil)).Elem()
	if m.IsValid() && m.Type().NumIn() == 2 && m.Type().In(0) == writerType && m.Type().In(1) == fieldValue.Type() {
		ret := m.Call([]reflect.Value{reflect.ValueOf(r), fieldValue})

		errorType := reflect.TypeOf((*error)(nil)).Elem()

		// Method(w Writer, v value) error
		if len(ret) == 1 && ret[0].Type() == errorType {
			if !ret[0].IsNil() {
				return true, ret[0].Interface().(error)
			}

			return true, nil
		}
	}

	return false, nil
}

func calcLength(v any, parentStructValues []reflect.Value) (int, error) {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Struct {
		return 0, &InvalidUnmarshalError{reflect.TypeOf(v)}
	}

	fieldCount := rv.NumField()
	valueType := rv.Type()

	sumLength := 0
	for i := 0; i < fieldCount; i++ {
		fieldType := valueType.Field(i)
		tags, err := parseTag(fieldType.Tag.Get(tagName))
		if err != nil {
			return 0, fmt.Errorf(`failed parseTag for field "%s": %w`, fieldType.Name, err)
		}

		fieldData, err := parseReadDataFromTags(rv, tags)
		if err != nil {
			return 0, fmt.Errorf(`failed parse ReadData from tags for field "%s": %w`, fieldType.Name, err)
		}
		fieldValue := rv.Field(i)
		length := getValueLength(rv, fieldValue, fieldData, parentStructValues)
		sumLength += length
		if err != nil {
			return 0, fmt.Errorf(`failed set value to field "%s": %w`, fieldType.Name, err)
		}
	}
	return sumLength, nil
}

func getValueLength(structValue, fieldValue reflect.Value, fieldData *fieldReadData, parentStructValues []reflect.Value) int {
	switch fieldValue.Kind() {

	case reflect.Int8, reflect.Uint8:
		if fieldData != nil && fieldData.Length != nil {
			return int(*fieldData.Length)
		}
		return 1
	case reflect.Int16, reflect.Uint16:
		if fieldData != nil && fieldData.Length != nil {
			return int(*fieldData.Length)
		}
		return 2
	case reflect.Int32, reflect.Uint32, reflect.Float32:
		if fieldData != nil && fieldData.Length != nil {
			return int(*fieldData.Length)
		}
		return 4
	case reflect.Int64, reflect.Uint64, reflect.Float64:
		if fieldData != nil && fieldData.Length != nil {
			return int(*fieldData.Length)
		}
		return 8
	case reflect.String:
		return len([]byte(fieldValue.String()))
	case reflect.Slice:
		if fieldData.Length == nil {
			*fieldData.Length = int64(fieldValue.Len())
		}

		elemSize := getValueLength(structValue, fieldValue.Index(0), fieldData.ElemFieldData, parentStructValues)
		return int(*fieldData.Length) * elemSize
	case reflect.Array:
		return fieldValue.Len()
	case reflect.Struct:
		if fieldData != nil && fieldData.Length != nil {
			return int(*fieldData.Length)
		}
		// marshal(fieldValue.Interface(), append(parentStructValues, structValue))
		sum, err := calcLength(fieldValue.Interface(), append(parentStructValues, structValue))
		if err != nil {
			panic("err: " + err.Error())
		}
		return sum
	default: // reflect.Int:
		panic(`type "` + fieldValue.Kind().String() + `" not supported`)
	}
	return 0
}
