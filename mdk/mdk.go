package mdk

import (
	"fmt"
	"reflect"
	"runtime"
	"unsafe"

	"github.com/wasify-io/wasify-go/internal/types"
	"github.com/wasify-io/wasify-go/internal/utils"
)

// ArgData represents an offset into WebAssembly memory that refers to an argument's location.
// This packed representation consists of a memory offset and the size of the argument data.
type ArgData uint64

// ResultOffset represents an offset into WebAssembly memory for function results.
type ResultOffset uint64

// Result is a structure that contains a size and a generic data representation for function results.
type Result struct {
	Size uint32
	Data any
}

// Arg prepares data for passing as an argument to a host function in WebAssembly.
// It accepts a generic data input and returns an ArgData which packs the memory offset and size of the data.
// This function abstracts away the complexity of memory management and conversion for the user.
//
// The runtime.KeepAlive call is used to ensure that the 'value' object is not garbage collected
// until the function finishes execution.
//
// ⚠️ Note: The ArgData returned by the Arg function does not need to be manually deallocated.
// The memory management is handled on the host side, where the allocated memory is automatically deallocated.
func Arg(value any) ArgData {

	packedData, err := AllocPack(value)
	if err != nil {
		panic(err)
	}

	runtime.KeepAlive(value)

	return ArgData(packedData)
}

// TODO: Update comment
func ReadResults(packedDatas ArgData) []*Result {

	if packedDatas == 0 {
		return nil
	}

	t, offsetU32, size := utils.UnpackUI64(uint64(packedDatas))

	if t != types.ValueTypePack {
		panic(fmt.Sprintf("can't unpack data, value type is not a type of valueTypePack. expected %d, got %d", types.ValueTypePack, t))
	}

	// calculate the number of elements in the array
	count := size / 8

	// read the packed pointers and sizes from the array
	packedData := unsafe.Slice(ptrToData[uint64](uint64(offsetU32)), count)

	data := make([]*Result, count)

	// Iterate over the packedData, unpack and read data of each element into a Result
	for i, pd := range packedData {

		v, s := ReadAny(ArgData(pd))

		data[i] = &Result{
			Size: s,
			Data: v,
		}
	}

	return data
}

// TODO: Update comment
func ReadAny(packedData ArgData) (any, uint32) {

	valueType, _, size := utils.UnpackUI64(uint64(packedData))
	var value any

	switch valueType {
	case types.ValueTypeBytes:
		value, _ = ReadBytes(packedData)
	case types.ValueTypeByte:
		value, _ = ReadBytes(packedData)
	case types.ValueTypeI32:
		value, _ = ReadBytes(packedData)
	case types.ValueTypeI64:
		value, _ = ReadBytes(packedData)
	case types.ValueTypeF32:
		value, _ = ReadBytes(packedData)
	case types.ValueTypeF64:
		value, _ = ReadBytes(packedData)
	case types.ValueTypeString:
		value, _ = ReadBytes(packedData)
	default:
		return nil, 0
	}

	return value, size
}

func ReadBytes(packedData ArgData) ([]byte, uint32) {
	valueType, offsetU32, size := utils.UnpackUI64(uint64(packedData))
	data := unsafe.Slice(ptrToData[byte](uint64(offsetU32)), size)
	if valueType != types.ValueTypeBytes {
		panic(fmt.Sprintf("data is not bytes: %s", reflect.ValueOf(data)))
	}
	return data, size
}

func ReadByte(packedData ArgData) byte {
	valueType, offsetU32, _ := utils.UnpackUI64(uint64(packedData))
	data := *ptrToData[byte](uint64(offsetU32))
	if valueType != types.ValueTypeBytes {
		panic(fmt.Sprintf("data is not byte: %s", reflect.ValueOf(data)))
	}
	return data
}

func ReadI32(packedData ArgData) uint32 {
	valueType, offsetU32, _ := utils.UnpackUI64(uint64(packedData))
	data := *ptrToData[uint32](uint64(offsetU32))
	if valueType != types.ValueTypeString {
		panic(fmt.Sprintf("data is not uint32: %s", reflect.ValueOf(data)))
	}
	return data
}
func ReadI64(packedData ArgData) uint64 {
	valueType, offsetU32, _ := utils.UnpackUI64(uint64(packedData))
	data := *ptrToData[uint64](uint64(offsetU32))
	if valueType != types.ValueTypeString {
		panic(fmt.Sprintf("data is not uint32: %s", reflect.ValueOf(data)))
	}
	return data
}
func ReadF32(packedData ArgData) float32 {
	valueType, offsetU32, _ := utils.UnpackUI64(uint64(packedData))
	data := *ptrToData[float32](uint64(offsetU32))
	if valueType != types.ValueTypeString {
		panic(fmt.Sprintf("data is not float32: %s", reflect.ValueOf(data)))
	}
	return data
}
func ReadF64(packedData ArgData) float64 {
	valueType, offsetU32, _ := utils.UnpackUI64(uint64(packedData))
	data := *ptrToData[float64](uint64(offsetU32))
	if valueType != types.ValueTypeString {
		panic(fmt.Sprintf("data is not float64: %s", reflect.ValueOf(data)))
	}
	return data
}
func ReadString(packedData ArgData) (string, uint32) {
	valueType, offsetU32, size := utils.UnpackUI64(uint64(packedData))
	data := unsafe.String(ptrToData[byte](uint64(offsetU32)), size)
	if valueType != types.ValueTypeString {
		panic(fmt.Sprintf("data is not a string: %s", reflect.ValueOf(data)))
	}
	return data, size
}

// AllocPack prepares data for interaction with WebAssembly by allocating the necessary memory.
// It accepts a generic input and returns a uint64 value that combines the value type, memory offset, size.
func AllocPack(data any) (uint64, error) {

	dataType, offsetSize, err := types.GetOffsetSizeAndDataTypeByConversion(data)
	if err != nil {
		return 0, err
	}

	var offset uint64

	switch dataType {
	case types.ValueTypeBytes:
		offset = AllocBytes(data.([]byte), offsetSize)
	case types.ValueTypeByte:
		offset = AllocByte(data.(byte))
	case types.ValueTypeI32:
		offset = AllocUint32(data.(uint32))
	case types.ValueTypeI64:
		offset = AllocUint64(data.(uint64))
	case types.ValueTypeF32:
		offset = AllocFloat32(data.(float32))
	case types.ValueTypeF64:
		offset = AllocFloat64(data.(float64))
	case types.ValueTypeString:
		offset = AllocString(data.(string), offsetSize)
	default:
		return 0, fmt.Errorf("unsupported data type %d for allocation", dataType)
	}

	return utils.PackUI64(dataType, uint32(offset), offsetSize)
}

func AllocBytes(data []byte, offsetSize uint32) uint64 {
	return bytesToLeakedPtr(data, offsetSize)
}
func AllocByte(data byte) uint64 {
	return byteToLeakedPtr(data)
}
func AllocUint32(data uint32) uint64 {
	return uint32ToLeakedPtr(data)
}
func AllocUint64(data uint64) uint64 {
	return uint64ToLeakedPtr(data)
}
func AllocFloat32(data float32) uint64 {
	return float32ToLeakedPtr(data)
}
func AllocFloat64(data float64) uint64 {
	return float64ToLeakedPtr(data)
}
func AllocString(data string, offsetSize uint32) uint64 {
	return stringToLeakedPtr(data, offsetSize)
}

// Free frees the memory.
func Free(packedDatas ...ArgData) {
	for _, p := range packedDatas {
		_, offset, _ := utils.UnpackUI64(uint64(p))
		free(uint64(offset))
	}
}
