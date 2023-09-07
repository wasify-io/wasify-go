package mdk

import (
	"fmt"
	"runtime"
	"unsafe"
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

// ValueType is an enumeration of supported data types for function parameters and returns.
type ValueType uint8

// valueTypePack is a reserved ValueType used for packed data.
const valueTypePack ValueType = 255

// These constants represent the possible data types that can be used in function parameters and returns.
const (
	ValueTypeBytes ValueType = iota
	ValueTypeByte
	ValueTypeI32
	ValueTypeI64
	ValueTypeF32
	ValueTypeF64
	ValueTypeString
)

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

// Results unpacks and returns the results of a host function in WebAssembly.
// It takes a ResultOffset, which contains packed memory offsets, and returns a slice of Result structs.
// This utility helps in reading the data returned by WebAssembly functions without dealing with the intricacies of memory offsets.
func Results(resultsOffset ResultOffset) []Result {

	if resultsOffset == 0 {
		return nil
	}

	t, offsetU32, size := UnpackUI64(uint64(resultsOffset))

	if t != valueTypePack {
		panic(fmt.Sprintf("can't unpack data, value type is not a type of valueTypePack. expected %d, got %d", valueTypePack, t))
	}

	// calculate the number of elements in the array
	count := size / 8

	// read the packed pointers and sizes from the array
	packedData := unsafe.Slice(ptrToData[uint64](uint64(offsetU32)), count)

	data := make([]Result, count)

	// Iterate over the packedData, unpack and read data of each element into a Result
	for i, pd := range packedData {
		valueType, offsetU32, size := UnpackUI64(pd)
		offset := uint64(offsetU32)

		var value any

		switch valueType {
		case ValueTypeBytes:
			value = unsafe.Slice(ptrToData[byte](offset), size)
		case ValueTypeByte:
			value = ptrToData[byte](offset)
		case ValueTypeI32:
			value = ptrToData[uint32](offset)
		case ValueTypeI64:
			value = ptrToData[uint64](offset)
		case ValueTypeF32:
			value = ptrToData[float32](offset)
		case ValueTypeF64:
			value = ptrToData[float64](offset)
		case ValueTypeString:
			value = string(unsafe.String(ptrToData[byte](offset), size))
		}

		data[i] = Result{
			Size: size,
			Data: value,
		}

	}

	return data
}

// Alloc prepares data for interaction with WebAssembly by allocating the necessary memory.
// It accepts a generic input and returns a uint64 value that combines the memory offset and size.
func Alloc(data any) (uint64, error) {

	dataType, offsetSize, err := GetOffsetSizeAndDataTypeByConversion(data)
	if err != nil {
		return 0, err
	}

	var offset uint64

	switch dataType {
	case ValueTypeBytes:
		offset = AllocBytes(data.([]byte), offsetSize)
	case ValueTypeByte:
		offset = AllocByte(data.(byte))
	case ValueTypeI32:
		offset = AllocUint32Le(data.(uint32))
	case ValueTypeI64:
		offset = AllocUint64Le(data.(uint64))
	case ValueTypeF32:
		offset = AllocFloat32Le(data.(float32))
	case ValueTypeF64:
		offset = AllocFloat64Le(data.(float64))
	case ValueTypeString:
		offset = AllocString(data.(string), offsetSize)
	default:
		return 0, fmt.Errorf("unsupported data type %d for allocation", dataType)
	}

	return PackUI64(dataType, uint32(offset), offsetSize)
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
	_, offset, _ := UnpackUI64(packedData)
	free(uint64(offset))
}

// PackUI64 takes a data type (in the form of a byte), a pointer (offset in memory),
// and a size (amount of memory/data to consider). It returns a packed uint64 representation.
//
// Structure of the packed uint64:
// - Highest 8 bits: data type
// - Next 32 bits: offset
// - Lowest 24 bits: size
//
// This function will return error if the provided size is larger than what can be represented in 24 bits
// (i.e., larger than 16,777,215).
func PackUI64(dataType ValueType, offset uint32, size uint32) (uint64, error) {
	// Check if the size can be represented in 24 bits
	if size >= (1 << 24) {
		return 0, fmt.Errorf("Size %d exceeds 24 bits precision %d", size, (1 << 24))
	}

	// Shift the dataType into the highest 8 bits
	// Shift the offset into the next 32 bits
	// Use the size as is, but ensure only the lowest 24 bits are used (using bitwise AND)
	return (uint64(dataType) << 56) | (uint64(offset) << 24) | uint64(size&0xFFFFFF), nil
}

// UnpackUI64 reverses the operation done by PackUI64.
// Given a packed uint64, it will extract and return the original dataType, offset (ptr), and size.
func UnpackUI64(packedData uint64) (dataType ValueType, offset uint32, size uint32) {
	// Extract the dataType from the highest 8 bits
	dataType = ValueType(packedData >> 56)

	// Extract the offset (ptr) from the next 32 bits using bitwise AND to mask the other bits
	offset = uint32((packedData >> 24) & 0xFFFFFFFF)

	// Extract the size from the lowest 24 bits
	size = uint32(packedData & 0xFFFFFF)

	return
}
