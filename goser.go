package goser

import (
	"encoding/binary"
	"fmt"
	"hash/fnv"
	"math"
	"reflect"
	"time"
	"unsafe"
)

var idToType map[uint32]reflect.Type
var typeToId map[reflect.Type]uint32

func init() {
	idToType = make(map[uint32]reflect.Type)
	typeToId = make(map[reflect.Type]uint32)
}

func Register(valueOfType any) {
	thetype := reflect.TypeOf(valueOfType)

	typeName := thetype.String()
	star := ""
	if thetype.Name() == "" {
		if thetype.Kind() == reflect.Pointer {
			star = "*"
			thetype = thetype.Elem()
		}
	}
	if thetype.Name() != "" {
		if thetype.PkgPath() == "" {
			typeName = star + thetype.Name()
		} else {
			typeName = star + thetype.PkgPath() + "." + thetype.Name()
		}
	}

	hash := fnv.New32a()
	hash.Write([]byte(typeName))
	typeId := hash.Sum32()

	_, typeKnown := idToType[typeId]
	if !typeKnown {
		idToType[typeId] = thetype
		typeToId[thetype] = typeId
	}
}

func Marshal(obj any) ([]byte, error) {
	thetype := reflect.TypeOf(obj)
	var kind reflect.Kind
	if thetype == nil {
		kind = reflect.Pointer
	} else {
		kind = thetype.Kind()
	}
	value := reflect.ValueOf(obj)
	serialized := make([]byte, 0)
	serialized = append(serialized, byte(kind))
	switch kind {
	case reflect.Bool:
		if value.Bool() {
			serialized = append(serialized, byte(1))
		} else {
			serialized = append(serialized, byte(0))
		}
	case reflect.Int:
		encodedInt := make([]byte, 8)
		binary.LittleEndian.PutUint64(encodedInt, uint64(value.Int()))
		serialized = append(serialized, encodedInt...)
	case reflect.Int8:
		serialized = append(serialized, byte(value.Int()))
	case reflect.Int16:
		encodedInt16 := make([]byte, 2)
		binary.LittleEndian.PutUint16(encodedInt16, uint16(value.Int()))
		serialized = append(serialized, encodedInt16...)
	case reflect.Int32:
		encodedInt32 := make([]byte, 4)
		binary.LittleEndian.PutUint32(encodedInt32, uint32(value.Int()))
		serialized = append(serialized, encodedInt32...)
	case reflect.Int64:
		encodedInt64 := make([]byte, 8)
		binary.LittleEndian.PutUint64(encodedInt64, uint64(value.Int()))
		serialized = append(serialized, encodedInt64...)
	case reflect.Uint:
		encodedUint := make([]byte, 8)
		binary.LittleEndian.PutUint64(encodedUint, uint64(value.Uint()))
		serialized = append(serialized, encodedUint...)
	case reflect.Uint8:
		serialized = append(serialized, byte(value.Uint()))
	case reflect.Uint16:
		encodedUint16 := make([]byte, 2)
		binary.LittleEndian.PutUint16(encodedUint16, uint16(value.Uint()))
		serialized = append(serialized, encodedUint16...)
	case reflect.Uint32:
		encodedUint32 := make([]byte, 4)
		binary.LittleEndian.PutUint32(encodedUint32, uint32(value.Uint()))
		serialized = append(serialized, encodedUint32...)
	case reflect.Uint64:
		encodedUint64 := make([]byte, 8)
		binary.LittleEndian.PutUint64(encodedUint64, value.Uint())
		serialized = append(serialized, encodedUint64...)
	case reflect.Uintptr:
		encodedUintptr := make([]byte, 8)
		binary.LittleEndian.PutUint64(encodedUintptr, value.Uint())
		serialized = append(serialized, encodedUintptr...)
	case reflect.Float32:
		encodedFloat32 := make([]byte, 4)
		binary.LittleEndian.PutUint32(encodedFloat32, math.Float32bits(float32(value.Float())))
		serialized = append(serialized, encodedFloat32...)
	case reflect.Float64:
		encodedFloat64 := make([]byte, 8)
		binary.LittleEndian.PutUint64(encodedFloat64, math.Float64bits(value.Float()))
		serialized = append(serialized, encodedFloat64...)
	case reflect.Complex64:
		encodedComplex64 := make([]byte, 8)
		objComplex64 := value.Complex()
		objReal32 := real(objComplex64)
		objImag32 := imag(objComplex64)
		binary.LittleEndian.PutUint32(encodedComplex64, math.Float32bits(float32(objReal32)))
		binary.LittleEndian.PutUint32(encodedComplex64[4:], math.Float32bits(float32(objImag32)))
		serialized = append(serialized, encodedComplex64...)
	case reflect.Complex128:
		encodedComplex128 := make([]byte, 16)
		objComplex128 := value.Complex()
		objReal64 := real(objComplex128)
		objImag64 := imag(objComplex128)
		binary.LittleEndian.PutUint64(encodedComplex128, math.Float64bits(objReal64))
		binary.LittleEndian.PutUint64(encodedComplex128[8:], math.Float64bits(objImag64))
		serialized = append(serialized, encodedComplex128...)
	case reflect.String:
		encodedLength := make([]byte, 8)
		length := value.Len()
		binary.LittleEndian.PutUint64(encodedLength, uint64(length))
		serialized = append(serialized, encodedLength...)
		serialized = append(serialized, []byte(value.Interface().(string))...)
	case reflect.Pointer:
		if obj != nil && !value.IsNil() {
			serialized = append(serialized, 1)
			encodedPointerContents, err := Marshal(value.Elem().Interface())
			if err != nil {
				return nil, fmt.Errorf("couldn't serialize pointer contents: %w", err)
			}
			serialized = append(serialized, encodedPointerContents...)
		} else {
			serialized = append(serialized, 0)
		}
	case reflect.Array:
		encodedLength := make([]byte, 8)
		length := value.Len()
		binary.LittleEndian.PutUint64(encodedLength, uint64(length))
		serialized = append(serialized, encodedLength...)
		typeMarker := reflect.Zero(thetype.Elem())
		encodedTypeMarker, err := Marshal(typeMarker.Interface())
		if err != nil {
			return nil, fmt.Errorf("couldn't serialize array type marker: %w", err)
		}
		serialized = append(serialized, encodedTypeMarker...)
		for i := 0; i < length; i++ {
			item := value.Index(i)
			encodedItem, err := Marshal(item.Interface())
			if err != nil {
				return nil, fmt.Errorf("couldn't serialize array item: %w", err)
			}
			serialized = append(serialized, encodedItem...)
		}
	case reflect.Slice:
		encodedLength := make([]byte, 8)
		length := value.Len()
		binary.LittleEndian.PutUint64(encodedLength, uint64(length))
		serialized = append(serialized, encodedLength...)
		typeMarker := reflect.Zero(thetype.Elem())
		encodedTypeMarker, err := Marshal(typeMarker.Interface())
		if err != nil {
			return nil, fmt.Errorf("couldn't serialize slice type marker: %w", err)
		}
		serialized = append(serialized, encodedTypeMarker...)
		for i := 0; i < length; i++ {
			item := value.Index(i)
			encodedItem, err := Marshal(item.Interface())
			if err != nil {
				return nil, fmt.Errorf("couldn't serialize slice item: %w", err)
			}
			serialized = append(serialized, encodedItem...)
		}
	case reflect.Map:
		encodedLength := make([]byte, 8)
		length := value.Len()
		binary.LittleEndian.PutUint64(encodedLength, uint64(length))
		serialized = append(serialized, encodedLength...)
		keyTypeMarker := reflect.Zero(thetype.Key())
		encodedKeyTypeMarker, err := Marshal(keyTypeMarker.Interface())
		if err != nil {
			return nil, fmt.Errorf("couldn't serialize map key type marker: %w", err)
		}
		serialized = append(serialized, encodedKeyTypeMarker...)
		valueTypeMarker := reflect.Zero(thetype.Elem())
		encodedValueTypeMarker, err := Marshal(valueTypeMarker.Interface())
		if err != nil {
			return nil, fmt.Errorf("couldn't serialize map value type marker: %w", err)
		}
		serialized = append(serialized, encodedValueTypeMarker...)
		mapRange := value.MapRange()
		for mapRange.Next() {
			key := mapRange.Key()
			encodedKey, err := Marshal(key.Interface())
			if err != nil {
				return nil, fmt.Errorf("couldn't serialize map key: %w", err)
			}
			serialized = append(serialized, encodedKey...)
			value := mapRange.Value()
			encodedValue, err := Marshal(value.Interface())
			if err != nil {
				return nil, fmt.Errorf("couldn't serialize map value: %w", err)
			}
			serialized = append(serialized, encodedValue...)
		}
	case reflect.Struct:
		if timeObj, ok := obj.(time.Time); ok {
			encodedTypeId := []byte("time")
			serialized = append(serialized, encodedTypeId...)
			encodedField, err := Marshal(timeObj.Format(time.RFC3339Nano))
			if err != nil {
				return nil, fmt.Errorf("couldn't serialize time.Time field: %w", err)
			}
			serialized = append(serialized, encodedField...)
		} else {
			typeId, typeKnown := typeToId[thetype]
			if !typeKnown {
				return nil, fmt.Errorf("can't serialize type %v (not registered)", thetype)
			}
			encodedTypeId := make([]byte, 4)
			binary.LittleEndian.PutUint32(encodedTypeId, typeId)
			serialized = append(serialized, encodedTypeId...)
			for i := 0; i < value.NumField(); i++ {
				structCopy := reflect.New(thetype).Elem()
				structCopy.Set(value)
				unsafeField := structCopy.Field(i)
				unsafeField = reflect.NewAt(unsafeField.Type(), unsafe.Pointer(unsafeField.UnsafeAddr())).Elem()
				encodedField, err := Marshal(unsafeField.Interface())
				if err != nil {
					return nil, fmt.Errorf("couldn't serialize struct field: %w", err)
				}
				serialized = append(serialized, encodedField...)
			}
		}
	case reflect.Chan:
		return nil, fmt.Errorf("can't serialize channel (%v)", kind)
	default:
		return nil, fmt.Errorf("can't serialize kind %v", kind)
	}
	return serialized, nil
}

func Unmarshal(serialized []byte) (any, error) {
	value, serialized, err := unmarshalRecursive(serialized)
	if len(serialized) > 0 {
		return nil, fmt.Errorf("couldn't consume all the provided bytes")
	}
	return value, err
}

func unmarshalRecursive(serialized []byte) (any, []byte, error) {
	if len(serialized) < 1 {
		return nil, nil, fmt.Errorf("can't read kind %v", serialized)
	}
	kind := reflect.Kind(serialized[0])
	serialized = serialized[1:]

	switch kind {
	case reflect.Bool:
		if len(serialized) < 1 {
			return nil, nil, fmt.Errorf("can't read bool %v", serialized)
		}
		if serialized[0] == 1 {
			return true, serialized[1:], nil
		}
		return false, serialized[1:], nil
	case reflect.Int:
		if len(serialized) < 8 {
			return nil, nil, fmt.Errorf("can't read int %v", serialized)
		}
		return binary.LittleEndian.Uint64(serialized[:8]), serialized[8:], nil
	case reflect.Int8:
		if len(serialized) < 1 {
			return nil, nil, fmt.Errorf("can't read int8 %v", serialized)
		}
		return serialized[0], serialized[1:], nil
	case reflect.Int16:
		if len(serialized) < 2 {
			return nil, nil, fmt.Errorf("can't read int16 %v", serialized)
		}
		return binary.LittleEndian.Uint16(serialized[:2]), serialized[2:], nil
	case reflect.Int32:
		if len(serialized) < 4 {
			return nil, nil, fmt.Errorf("can't read int32 %v", serialized)
		}
		return binary.LittleEndian.Uint32(serialized[:4]), serialized[4:], nil
	case reflect.Int64:
		if len(serialized) < 8 {
			return nil, nil, fmt.Errorf("can't read int64 %v", serialized)
		}
		return binary.LittleEndian.Uint64(serialized[:8]), serialized[8:], nil
	case reflect.Uint:
		if len(serialized) < 8 {
			return nil, nil, fmt.Errorf("can't read uint %v", serialized)
		}
		return binary.LittleEndian.Uint64(serialized[:8]), serialized[8:], nil
	case reflect.Uint8:
		if len(serialized) < 1 {
			return nil, nil, fmt.Errorf("can't read uint8 %v", serialized)
		}
		return uint8(serialized[0]), serialized[1:], nil
	case reflect.Uint16:
		if len(serialized) < 2 {
			return nil, nil, fmt.Errorf("can't read uint16 %v", serialized)
		}
		return uint16(binary.LittleEndian.Uint16(serialized[:2])), serialized[2:], nil
	case reflect.Uint32:
		if len(serialized) < 4 {
			return nil, nil, fmt.Errorf("can't read uint32 %v", serialized)
		}
		return uint32(binary.LittleEndian.Uint32(serialized[:4])), serialized[4:], nil
	case reflect.Uint64:
		if len(serialized) < 8 {
			return nil, nil, fmt.Errorf("can't read uint64 %v", serialized)
		}
		return binary.LittleEndian.Uint64(serialized[:8]), serialized[8:], nil
	case reflect.Uintptr:
		if len(serialized) < 8 {
			return nil, nil, fmt.Errorf("can't read uintptr %v", serialized)
		}
		return uintptr(binary.LittleEndian.Uint64(serialized[:8])), serialized[8:], nil
	case reflect.Float32:
		if len(serialized) < 4 {
			return nil, nil, fmt.Errorf("can't read float32 %v", serialized)
		}
		return math.Float32frombits(binary.LittleEndian.Uint32(serialized[:4])), serialized[4:], nil
	case reflect.Float64:
		if len(serialized) < 8 {
			return nil, nil, fmt.Errorf("can't read float64 %v", serialized)
		}
		return math.Float64frombits(binary.LittleEndian.Uint64(serialized[:8])), serialized[8:], nil
	case reflect.Complex64:
		if len(serialized) < 8 {
			return nil, nil, fmt.Errorf("can't read complex64 %v", serialized)
		}
		return complex(math.Float32frombits(binary.LittleEndian.Uint32(serialized[:4])), math.Float32frombits(binary.LittleEndian.Uint32(serialized[4:8]))), serialized[8:], nil
	case reflect.Complex128:
		if len(serialized) < 16 {
			return nil, nil, fmt.Errorf("can't read complex128 %v", serialized)
		}
		return complex(math.Float64frombits(binary.LittleEndian.Uint64(serialized[:8])), math.Float64frombits(binary.LittleEndian.Uint64(serialized[8:16]))), serialized[16:], nil
	case reflect.String:
		if len(serialized) < 8 {
			return nil, nil, fmt.Errorf("can't read string length %v", serialized)
		}
		length := binary.LittleEndian.Uint64(serialized[:8])
		serialized = serialized[8:]
		if len(serialized) < int(length) {
			return nil, nil, fmt.Errorf("can't read string as not enough data is present %v", serialized)
		}
		value := string(serialized[:length])
		serialized = serialized[length:]
		return value, serialized, nil
	case reflect.Pointer:
		if len(serialized) < 1 {
			return nil, nil, fmt.Errorf("can't read pointer nil status %v", serialized)
		}
		nonNil := uint8(serialized[0])
		serialized = serialized[1:]
		if nonNil == 1 {
			obj, serialized, err := unmarshalRecursive(serialized)
			if err != nil {
				return nil, nil, fmt.Errorf("couldn't deserialize pointer contents: %w", err)
			}
			pointer := reflect.New(reflect.TypeOf(obj))
			pointer.Elem().Set(reflect.ValueOf(obj))
			return pointer.Interface(), serialized, nil
		} else {
			return nil, serialized, nil
		}
	case reflect.Array:
		if len(serialized) < 8 {
			return nil, nil, fmt.Errorf("can't read array length %v", serialized)
		}
		length := binary.LittleEndian.Uint64(serialized[:8])
		serialized = serialized[8:]
		typeMarker, serialized, err := unmarshalRecursive(serialized)
		if err != nil {
			return nil, nil, fmt.Errorf("couldn't deserialize array type marker: %w", err)
		}
		arrayPtr := reflect.New(reflect.ArrayOf(int(length), reflect.TypeOf(typeMarker)))
		array := arrayPtr.Elem()
		for i := 0; i < int(length); i++ {
			var item any
			var err error
			item, serialized, err = unmarshalRecursive(serialized)
			if err != nil {
				return nil, nil, fmt.Errorf("couldn't deserialize array item: %w", err)
			}
			array.Index(i).Set(reflect.ValueOf(item))
		}
		return array.Interface(), serialized, nil
	case reflect.Slice:
		if len(serialized) < 8 {
			return nil, nil, fmt.Errorf("can't read slice length %v", serialized)
		}
		length := binary.LittleEndian.Uint64(serialized[:8])
		serialized = serialized[8:]
		typeMarker, serialized, err := unmarshalRecursive(serialized)
		if err != nil {
			return nil, nil, fmt.Errorf("couldn't deserialize slice type marker: %w", err)
		}
		slice := reflect.MakeSlice(reflect.SliceOf(reflect.TypeOf(typeMarker)), int(length), int(length))
		for i := 0; i < int(length); i++ {
			var item any
			var err error
			item, serialized, err = unmarshalRecursive(serialized)
			if err != nil {
				return nil, nil, fmt.Errorf("couldn't deserialize slice item: %w", err)
			}
			slice.Index(i).Set(reflect.ValueOf(item))
		}
		return slice.Interface(), serialized, nil
	case reflect.Map:
		if len(serialized) < 8 {
			return nil, nil, fmt.Errorf("can't read map length %v", serialized)
		}
		length := binary.LittleEndian.Uint64(serialized[:8])
		serialized = serialized[8:]
		keyTypeMarker, serialized, err := unmarshalRecursive(serialized)
		if err != nil {
			return nil, nil, fmt.Errorf("couldn't deserialize map key type marker: %w", err)
		}
		valueTypeMarker, serialized, err := unmarshalRecursive(serialized)
		if err != nil {
			return nil, nil, fmt.Errorf("couldn't deserialize map value type marker: %w", err)
		}
		themap := reflect.MakeMap(reflect.MapOf(reflect.TypeOf(keyTypeMarker), reflect.TypeOf(valueTypeMarker)))
		for i := 0; i < int(length); i++ {
			var key, itemValue any
			var err error
			key, serialized, err = unmarshalRecursive(serialized)
			if err != nil {
				return nil, nil, fmt.Errorf("couldn't deserialize map key: %w", err)
			}
			itemValue, serialized, err = unmarshalRecursive(serialized)
			if err != nil {
				return nil, nil, fmt.Errorf("couldn't deserialize map value: %w", err)
			}
			themap.SetMapIndex(reflect.ValueOf(key), reflect.ValueOf(itemValue))
		}
		return themap.Interface(), serialized, nil
	case reflect.Struct:
		if len(serialized) < 4 {
			return nil, nil, fmt.Errorf("can't read struct type id %v", serialized)
		}
		if string(serialized[:4]) == "time" {
			serialized = serialized[4:]
			value, serialized, err := unmarshalRecursive(serialized)
			if err != nil {
				return nil, nil, fmt.Errorf("couldn't deserialize time as string: %w", err)
			}
			if timeAsStr, ok := value.(string); ok {
				parsedTime, err := time.Parse(time.RFC3339Nano, timeAsStr)
				if err != nil {
					return nil, nil, fmt.Errorf("couldn't parse time as string: %w", err)
				}
				return parsedTime, serialized, nil
			} else {
				return nil, nil, fmt.Errorf("expected time as string but got %T", value)
			}
		} else {
			typeId := binary.LittleEndian.Uint32(serialized[:4])
			theType, typeKnown := idToType[typeId]
			if !typeKnown {
				return nil, nil, fmt.Errorf("can't deserialize type id %v (not registered)", typeId)
			}
			serialized = serialized[4:]
			structCopy := reflect.New(theType).Elem()
			for i := 0; i < structCopy.NumField(); i++ {
				field := structCopy.Field(i)
				var fieldValue any
				var err error
				fieldValue, serialized, err = unmarshalRecursive(serialized)
				if err != nil {
					return nil, nil, fmt.Errorf("couldn't deserialize struct field: %w", err)
				}
				if fieldValue != nil {
					unsafeField := reflect.NewAt(field.Type(), unsafe.Pointer(field.UnsafeAddr())).Elem()
					unsafeField.Set(reflect.ValueOf(fieldValue).Convert(field.Type()))
				}
			}
			return structCopy.Interface(), serialized, nil
		}
	case reflect.Chan:
		return nil, nil, fmt.Errorf("can't deserialize channel (%v)", kind)
	default:
		return nil, nil, fmt.Errorf("can't deserialize kind %v", kind)
	}
}
