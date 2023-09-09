package mdk

import (
	"fmt"
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

	packedData, err := Alloc(value)
	if err != nil {
		panic(err)
	}

	runtime.KeepAlive(value)

	return ArgData(packedData)
}

// TODO: Update comment
func Results(resultsOffset ArgData) []*Result {

	if resultsOffset == 0 {
		return nil
	}

	t, offsetU32, size := utils.UnpackUI64(uint64(resultsOffset))

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
		data[i] = ReadOne(ArgData(pd))
	}

	return data
}

// TODO: Update comment
func ReadOne(packedData ArgData) *Result {

	if packedData == 0 {
		return nil
	}

	valueType, offsetU32, size := utils.UnpackUI64(uint64(packedData))
	offset := uint64(offsetU32)

	var value any

	switch valueType {
	case types.ValueTypeBytes:
		value = unsafe.Slice(ptrToData[byte](offset), size)
	case types.ValueTypeByte:
		value = ptrToData[byte](offset)
	case types.ValueTypeI32:
		value = ptrToData[uint32](offset)
	case types.ValueTypeI64:
		value = ptrToData[uint64](offset)
	case types.ValueTypeF32:
		value = ptrToData[float32](offset)
	case types.ValueTypeF64:
		value = ptrToData[float64](offset)
	case types.ValueTypeString:
		value = string(unsafe.String(ptrToData[byte](offset), size))
	default:
		return nil
	}

	return &Result{
		Size: size,
		Data: value,
	}

}

// Alloc prepares data for interaction with WebAssembly by allocating the necessary memory.
// It accepts a generic input and returns a uint64 value that combines the memory offset and size.
func Alloc(data any) (uint64, error) {

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
		offset = AllocUint32Le(data.(uint32))
	case types.ValueTypeI64:
		offset = AllocUint64Le(data.(uint64))
	case types.ValueTypeF32:
		offset = AllocFloat32Le(data.(float32))
	case types.ValueTypeF64:
		offset = AllocFloat64Le(data.(float64))
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
func AllocUint32Le(data uint32) uint64 {
	return uint32ToLeakedPtr(data)
}
func AllocUint64Le(data uint64) uint64 {
	return uint64ToLeakedPtr(data)
}
func AllocFloat32Le(data float32) uint64 {
	return float32ToLeakedPtr(data)
}
func AllocFloat64Le(data float64) uint64 {
	return float64ToLeakedPtr(data)
}
func AllocString(data string, offsetSize uint32) uint64 {
	return stringToLeakedPtr(data, offsetSize)
}

// FreeMemory frees the memory allocated by the AllocateString or AllocateBytes functions.
// It takes a uint64 that packs a pointer to the allocated memory and its size,
// then sets the memory to zeros and frees it.
func Free(packedData uint64) {
	_, offset, _ := utils.UnpackUI64(packedData)
	free(uint64(offset))
}
