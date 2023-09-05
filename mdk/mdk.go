package mdk

import (
	"runtime"
	"unsafe"
)

type ArgOffset uint64
type ResultOffset uint64

type Result struct {
	Size uint32
	Data any
}

// ValueType represents the type of value used in function parameters and returns.
type ValueType uint8

const valueTypePack uint8 = 0
const (
	ValueTypeBytes ValueType = iota + 1
	ValueTypeI32
	ValueTypeI64
	ValueTypeF32
	ValueTypeF64
)

// Arg is a utility function that converts data of type 'any' into an ArgOffset,
// which is a packed representation of the pointer to the allocated memory and the size of the data.
// Currently, it only supports data of type 'string' or '[]byte'.
// For other data types, it will panic with 'unsupported data type'.
// This function is particularly useful for passing data to host functions,
// as it handles the necessary conversions and memory allocations.
//
// The runtime.KeepAlive call is used to ensure that the 'data' object is not garbage collected
// until the function finishes execution.
//
// ⚠️ Note: The ArgOffset returned by the Arg function does not need to be manually deallocated.
// The memory management is handled on the host side, where the allocated memory is automatically deallocated.
func Arg(data any) ArgOffset {
	var b []byte
	switch v := data.(type) {
	case string:
		b = []byte(v)
	case []byte:
		b = v
	default:
		panic("unsupported data type")
	}

	runtime.KeepAlive(b)

	return ArgOffset(Alloc(b))
}

// Results converts a packed result offset into a slice of Result structs.
//
// Each Result struct in the slice contains the size of the data and a slice of bytes representing the data.
// The resultsOffset is an unsigned integer that packs a pointer to the allocated memory and the size of the byte slice.
// The function unpacks this offset into the actual memory offset and size, then iterates over the packed data,
// unpacking each element into a Result struct and storing it in a slice of Results.
// Finally, the function returns the slice of Results.
func Results(resultsOffset ResultOffset) []Result {
	_, offset, size := UnpackUI64(uint64(resultsOffset))

	// calculate the number of elements in the array
	count := size / 8

	// read the packed pointers and sizes from the array
	packedData := unsafe.Slice((*uint64)(unsafe.Pointer(uintptr(offset))), count)

	data := make([]Result, count)

	// Iterate over the packedData, unpack and read data of each element into a Result
	for i, pd := range packedData {
		t, offset, size := UnpackUI64(pd)

		var value any

		switch ValueType(t) {
		case ValueTypeBytes:
			value = unsafe.Slice((*byte)(unsafe.Pointer(uintptr(offset))), size)
		case ValueTypeI32:
			value = *(*uint32)(unsafe.Pointer(uintptr(offset)))
		case ValueTypeI64:
			value = *(*uint64)(unsafe.Pointer(uintptr(offset)))
		case ValueTypeF32:
			value = *(*float32)(unsafe.Pointer(uintptr(offset)))
		case ValueTypeF64:
			value = *(*float64)(unsafe.Pointer(uintptr(offset)))
		}

		data[i] = Result{
			Size: size,
			Data: value,
		}

	}

	return data
}

// Alloc allocates memory for a byte slice and returns a uint64
// that packs the pointer to the allocated memory and the size of the byte slice.
// This function is useful for passing byte slices to WebAssembly functions.
func Alloc(data []byte) uint64 {
	offset, size := bytesToLeakedPtr(data)
	return PackUI64(valueTypePack, offset, size)
}

// FreeMemory frees the memory allocated by the AllocateString or AllocateBytes functions.
// It takes a uint64 that packs a pointer to the allocated memory and its size,
// then sets the memory to zeros and frees it.
func Free(packedData uint64) {
	_, offset, _ := UnpackUI64(packedData)
	free(offset)
}

// PackUI64 takes a data type (in the form of a byte), a pointer (offset in memory),
// and a size (amount of memory/data to consider). It returns a packed uint64 representation.
//
// Structure of the packed uint64:
// - Highest 8 bits: data type
// - Next 32 bits: offset (ptr)
// - Lowest 24 bits: size
//
// This function will panic if the provided size is larger than what can be represented in 24 bits (i.e., larger than 16,777,215).
func PackUI64(dataType uint8, ptr uint32, size uint32) (packedData uint64) {
	// Check if the size can be represented in 24 bits
	if size >= (1 << 24) {
		panic("Size exceeds 24 bits precision")
	}

	// Shift the dataType into the highest 8 bits
	// Shift the ptr (offset) into the next 32 bits
	// Use the size as is, but ensure only the lowest 24 bits are used (using bitwise AND)
	return (uint64(dataType) << 56) | (uint64(ptr) << 24) | uint64(size&0xFFFFFF)
}

// UnpackUI64 reverses the operation done by PackUI64.
// Given a packed uint64, it will extract and return the original dataType, offset (ptr), and size.
func UnpackUI64(packedData uint64) (dataType uint8, offset uint32, size uint32) {
	// Extract the dataType from the highest 8 bits
	dataType = uint8(packedData >> 56)

	// Extract the offset (ptr) from the next 32 bits using bitwise AND to mask the other bits
	offset = uint32((packedData >> 24) & 0xFFFFFFFF)

	// Extract the size from the lowest 24 bits
	size = uint32(packedData & 0xFFFFFF)

	return
}
