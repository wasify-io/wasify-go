package mdk

import (
	"runtime"
	"unsafe"
)

type ArgOffset uint64
type ResultOffset uint64

type Result struct {
	Size uint32
	Data []byte
}

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

	runtime.KeepAlive(data)

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
	offset, size := UnpackUI64(uint64(resultsOffset))

	// calculate the number of elements in the array
	count := size / 8

	// read the packed pointers and sizes from the array
	packedData := unsafe.Slice((*uint64)(unsafe.Pointer(uintptr(offset))), count)

	data := make([]Result, count)

	// Iterate over the packedData, unpack and read data of each element into a Result
	for i, pd := range packedData {
		offset, size := UnpackUI64(pd)

		data[i] = Result{
			Size: size,
			Data: unsafe.Slice((*byte)(unsafe.Pointer(uintptr(offset))), size),
		}
	}

	return data
}

// Alloc allocates memory for a byte slice and returns a uint64
// that packs the pointer to the allocated memory and the size of the byte slice.
// This function is useful for passing byte slices to WebAssembly functions.
func Alloc(data []byte) uint64 {
	ptr, size := bytesToLeakedPtr(data)
	return PackUI64(ptr, size)
}

// FreeMemory frees the memory allocated by the AllocateString or AllocateBytes functions.
// It takes a uint64 that packs a pointer to the allocated memory and its size,
// then sets the memory to zeros and frees it.
func Free(packedData uint64) {
	offset, _ := UnpackUI64(packedData)
	free(offset)
}

// packUI64 packs an offset and a size into a single uint64.
// The offset is stored in the higher 32 bits and the size in the lower 32 bits.
func PackUI64(ptr uint32, size uint32) (packedData uint64) {
	return (uint64(ptr) << 32) | uint64(size)
}

// unpackUI64 unpacks an offset and a size from a single uint64.
// The offset is extracted from the higher 32 bits and the size from the lower 32 bits.
func UnpackUI64(packedData uint64) (offset uint32, size uint32) {
	offset = uint32(packedData >> 32)
	size = uint32(packedData)

	return
}
