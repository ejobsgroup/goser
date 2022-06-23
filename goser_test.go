package goser

import (
	"fmt"
	"log"
	"reflect"
	"testing"
	"time"
	"unsafe"
)

func TestRegister(t *testing.T) {
	type TestSubContainer struct {
		floatiness   float32
		floatitude   float64
		complexion   complex64
		complexitude complex128
	}
	type TestContainer struct {
		subPtr          *TestSubContainer
		subStruct       TestSubContainer
		truthiness      bool
		falsiness       bool
		numberness      int
		numberone       int8
		numbertwo       int16
		numberfour      int32
		numbereight     int64
		notanumberness  uint
		notanumberone   uint8
		notanumbertwo   uint16
		notanumberfour  uint32
		notanumbereight uint64
		whatisthisthing uintptr
		textiness       string
		sliceiness      []byte
		arrayiness      [5]int
		mapination      map[string]int
	}
	Register(TestSubContainer{})
	Register(TestContainer{})
	instance := &TestContainer{
		subPtr: &TestSubContainer{
			floatiness:   1.11,
			complexitude: 1 + 8i,
		},
		subStruct: TestSubContainer{
			floatiness:   3.14,
			complexitude: 10i,
		},
		truthiness: true,
		mapination: map[string]int{"a": 4},
	}
	bytes, err := Marshal(instance)
	if err != nil {
		t.Error(err)
	}
	anyInstance, err := Unmarshal(bytes)
	if err != nil {
		t.Error(err)
	}
	frankenInstance, ok := anyInstance.(*TestContainer)
	if !ok {
		t.Error(fmt.Errorf("can't cast to original struct"))
	}
	if instance.numberness != frankenInstance.numberness {
		t.Error(fmt.Errorf("before and after for int is not the same"))
	}
	if instance.subPtr.complexitude != frankenInstance.subPtr.complexitude {
		t.Error(fmt.Errorf("before and after for complex128 is not the same"))
	}
}

func TestNullPtr(t *testing.T) {
	bytes, err := Marshal(nil)
	if err != nil {
		t.Error(err)
	}
	stillNil, err := Unmarshal(bytes)
	if err != nil {
		t.Error(err)
	}
	if stillNil != nil {
		t.Error(fmt.Errorf("before and after for nil is not the same"))
	}
}

func TestPtrChan(t *testing.T) {
	testChan := make(chan int, 0)
	ptrChan := &testChan
	_, err := Marshal(ptrChan)
	if err == nil {
		t.Error(fmt.Errorf("no error raised"))
	}
}

func TestSliceChan(t *testing.T) {
	_, err := Marshal([]chan int{make(chan int)})
	if err == nil {
		t.Error(fmt.Errorf("no error raised"))
	}
}

func TestAnySliceChan(t *testing.T) {
	_, err := Marshal([]any{make(chan int)})
	if err == nil {
		t.Error(fmt.Errorf("no error raised"))
	}
}

func TestArrayChan(t *testing.T) {
	_, err := Marshal([1]chan int{make(chan int)})
	if err == nil {
		t.Error(fmt.Errorf("no error raised"))
	}
}

func TestAnyArrayChan(t *testing.T) {
	_, err := Marshal([1]any{make(chan int)})
	if err == nil {
		t.Error(fmt.Errorf("no error raised"))
	}
}

func TestStructChan(t *testing.T) {
	type TheStruct struct {
		Ch chan int
	}
	Register(TheStruct{})
	_, err := Marshal(&TheStruct{Ch: make(chan int)})
	if err == nil {
		t.Error(fmt.Errorf("no error raised"))
	}
}

func TestMapChanKey(t *testing.T) {
	_, err := Marshal(map[chan int]string{make(chan int): ""})
	if err == nil {
		t.Error(fmt.Errorf("no error raised"))
	}
}

func TestMapChanValue(t *testing.T) {
	_, err := Marshal(map[string]chan int{"": make(chan int)})
	if err == nil {
		t.Error(fmt.Errorf("no error raised"))
	}
}

func TestAnyMapChanKey(t *testing.T) {
	_, err := Marshal(map[any]any{make(chan int): ""})
	if err == nil {
		t.Error(fmt.Errorf("no error raised"))
	}
}

func TestAnyMapChanValue(t *testing.T) {
	_, err := Marshal(map[any]any{"": make(chan int)})
	if err == nil {
		t.Error(fmt.Errorf("no error raised"))
	}
}

func TestSliceInt(t *testing.T) {
	ints := make([]int, 0)
	sampleInts := []int{0, 1, 2, 3, 4}
	ints = append(ints, sampleInts...)
	bytes, err := Marshal(ints)
	if err != nil {
		t.Error(err)
	}
	anyStillInts, err := Unmarshal(bytes)
	if err != nil {
		t.Error(err)
	}
	log.Printf("%#v", anyStillInts)
	stillInts, ok := anyStillInts.([]int)
	if !ok {
		t.Error(fmt.Errorf("after object is not []int"))
	}
	for _, i := range sampleInts {
		if stillInts[i] != i {
			t.Error(fmt.Errorf("before and after for []int is not the same"))
		}
	}
}

func TestUnsafePointer(t *testing.T) {
	a := [16]int{3: 3, 9: 9, 11: 11}
	p9 := &a[9]
	up9 := unsafe.Pointer(p9)
	_, err := Marshal(up9)
	if err == nil {
		t.Error(fmt.Errorf("no error raised"))
	}
}

func TestUnregisteredType(t *testing.T) {
	type IAmNotRegistered struct{}
	_, err := Marshal(IAmNotRegistered{})
	if err == nil {
		t.Error(fmt.Errorf("no error raised"))
	}
}

func TestExtraBytes(t *testing.T) {
	bytes, err := Marshal(0)
	if err != nil {
		t.Error(err)
	}
	bytes = append(bytes, 0)
	_, err = Unmarshal(bytes)
	if err == nil {
		t.Error(fmt.Errorf("no error raised"))
	}
}

func TestTime(t *testing.T) {
	timeNow := time.Now()
	bytes, err := Marshal(timeNow)
	if err != nil {
		t.Error(err)
	}
	timeNowStr := timeNow.Format(time.RFC3339Nano)
	timeNow, err = time.Parse(time.RFC3339Nano, timeNowStr)
	if err != nil {
		t.Error(err)
	}
	anyTimeNow, err := Unmarshal(bytes)
	if err != nil {
		t.Error(err)
	}
	backToTimeNow := anyTimeNow.(time.Time)
	if backToTimeNow != timeNow {
		t.Error(fmt.Errorf("before and after for time.Time is not the same"))
	}
}

func TestRegisterUnnamed(t *testing.T) {
	Register(struct{}{})
	if idToType[0x5c18d754] == nil {
		t.Errorf("empty struct not registered")
	}
	Register(&struct{}{})
	if idToType[0xecee41fc] == nil {
		t.Errorf("empty struct pointer not registered")
	}
	Register("")
	if idToType[0x17c16538] == nil {
		t.Errorf("empty string not registered")
	}
}

func TestUnknown(t *testing.T) {
	bytes := []byte{27}
	_, err := Unmarshal(bytes)
	if err == nil {
		t.Error(fmt.Errorf("no error raised"))
	}
}

func TestNothing(t *testing.T) {
	bytes := make([]byte, 0)
	_, err := Unmarshal(bytes)
	if err == nil {
		t.Error(fmt.Errorf("no error raised"))
	}
}

func TestShortTypes(t *testing.T) {
	for typ := reflect.Bool; typ < reflect.UnsafePointer; typ++ {
		bytes := []byte{byte(typ)}
		_, err := Unmarshal(bytes)
		if err == nil {
			t.Error(fmt.Errorf("no error raised for type: %d", typ))
		}
	}
}

func TestShortString(t *testing.T) {
	bytes := []byte{byte(reflect.String), 2, 0, 0, 0, 0, 0, 0, 0, 65}
	_, err := Unmarshal(bytes)
	if err == nil {
		t.Error(fmt.Errorf("no error raised for short string"))
	}
}

func TestPointerToChan(t *testing.T) {
	bytes := []byte{byte(reflect.Pointer), 1, byte(reflect.Chan)}
	_, err := Unmarshal(bytes)
	if err == nil {
		t.Error(fmt.Errorf("no error raised for pointer to chan"))
	}
}

func TestArrayOfChan(t *testing.T) {
	bytes := []byte{byte(reflect.Array), 0, 0, 0, 0, 0, 0, 0, 0, byte(reflect.Chan)}
	_, err := Unmarshal(bytes)
	if err == nil {
		t.Error(fmt.Errorf("no error raised for array of chan"))
	}
}

func TestArrayOfPointerToChan(t *testing.T) {
	bytes := []byte{byte(reflect.Array), 1, 0, 0, 0, 0, 0, 0, 0, byte(reflect.Pointer), 1, byte(reflect.Bool), 0, byte(reflect.Pointer), 1, byte(reflect.Chan)}
	_, err := Unmarshal(bytes)
	if err == nil {
		t.Error(fmt.Errorf("no error raised for array of pointer to chan"))
	}
}

func TestSliceOfChan(t *testing.T) {
	bytes := []byte{byte(reflect.Slice), 0, 0, 0, 0, 0, 0, 0, 0, byte(reflect.Chan)}
	_, err := Unmarshal(bytes)
	if err == nil {
		t.Error(fmt.Errorf("no error raised for slice of chan"))
	}
}

func TestSliceOfPointerToChan(t *testing.T) {
	bytes := []byte{byte(reflect.Slice), 1, 0, 0, 0, 0, 0, 0, 0, byte(reflect.Pointer), 1, byte(reflect.Bool), 0, byte(reflect.Pointer), 1, byte(reflect.Chan)}
	_, err := Unmarshal(bytes)
	if err == nil {
		t.Error(fmt.Errorf("no error raised for slice of pointer to chan"))
	}
}

func TestMapOfChanKey(t *testing.T) {
	bytes := []byte{byte(reflect.Map), 0, 0, 0, 0, 0, 0, 0, 0, byte(reflect.Chan)}
	_, err := Unmarshal(bytes)
	if err == nil {
		t.Error(fmt.Errorf("no error raised for map key of chan"))
	}
}

func TestMapOfChanValue(t *testing.T) {
	bytes := []byte{byte(reflect.Map), 0, 0, 0, 0, 0, 0, 0, 0, byte(reflect.Bool), 0, byte(reflect.Chan)}
	_, err := Unmarshal(bytes)
	if err == nil {
		t.Error(fmt.Errorf("no error raised for map value of chan"))
	}
}

func TestMapOfChanItemKey(t *testing.T) {
	bytes := []byte{byte(reflect.Map), 1, 0, 0, 0, 0, 0, 0, 0, byte(reflect.Bool), 0, byte(reflect.Bool), 0, byte(reflect.Chan)}
	_, err := Unmarshal(bytes)
	if err == nil {
		t.Error(fmt.Errorf("no error raised for map item value of chan"))
	}
}

func TestMapOfChanItemValue(t *testing.T) {
	bytes := []byte{byte(reflect.Map), 1, 0, 0, 0, 0, 0, 0, 0, byte(reflect.Bool), 0, byte(reflect.Bool), 0, byte(reflect.Bool), 0, byte(reflect.Chan)}
	_, err := Unmarshal(bytes)
	if err == nil {
		t.Error(fmt.Errorf("no error raised for map item value of chan"))
	}
}

func TestInvalidTime(t *testing.T) {
	bytes := []byte{byte(reflect.Struct), 't', 'i', 'm', 'e', byte(reflect.Chan)}
	_, err := Unmarshal(bytes)
	if err == nil {
		t.Error(fmt.Errorf("no error raised for invalid time"))
	}
}

func TestInvalidTimeString(t *testing.T) {
	bytes := []byte{byte(reflect.Struct), 't', 'i', 'm', 'e', byte(reflect.String), 1, 0, 0, 0, 0, 0, 0, 0, 'x'}
	_, err := Unmarshal(bytes)
	if err == nil {
		t.Error(fmt.Errorf("no error raised for invalid time"))
	}
}

func TestInvalidTimeType(t *testing.T) {
	bytes := []byte{byte(reflect.Struct), 't', 'i', 'm', 'e', byte(reflect.Bool), 1}
	_, err := Unmarshal(bytes)
	if err == nil {
		t.Error(fmt.Errorf("no error raised for invalid time"))
	}
}

func TestStructWithInvalidType(t *testing.T) {
	bytes := []byte{byte(reflect.Struct), 0xe0, 0x1d, 0xdf, 0x2c, byte(reflect.Chan)}
	_, err := Unmarshal(bytes)
	if err == nil {
		t.Error(fmt.Errorf("no error raised for struct with invalid type"))
	}
}

func TestStructWithInvalidField(t *testing.T) {
	Register(struct{ a chan int }{})
	bytes := []byte{byte(reflect.Struct), 0xe0, 0x1d, 0xdf, 0x2c, byte(reflect.Chan)}
	_, err := Unmarshal(bytes)
	if err == nil {
		t.Error(fmt.Errorf("no error raised for struct with invalid field"))
	}
}
