package mdk

// #include <stdlib.h>
// #include <string.h>
import "C"
import (
	"encoding/binary"
	"fmt"
	"unsafe"

	"github.com/wasify-io/wasify-go/internal/types"
)

func readBytes(offset64 uint64, size int) []byte {
	return unsafe.Slice(ptrToData[byte](offset64), size)
}

func readByte(offset64 uint64) byte {
	return *ptrToData[byte](offset64)
}

func readI32(offset64 uint64) uint32 {
	return *ptrToData[uint32](offset64)
}

func readI64(offset64 uint64) uint64 {
	return *ptrToData[uint64](offset64)
}

func readF32(offset64 uint64) float32 {
	return *ptrToData[float32](offset64)
}

func readF64(offset64 uint64) float64 {
	return *ptrToData[float64](uint64(offset64))
}

func readString(offset64 uint64, size int) string {
	return string(unsafe.Slice(ptrToData[byte](offset64), size))
}

// bytesToLeakedPtr converts a byte slice to an offset and size pair.
// It allocates memory of size 'len(data)' and copies the data into this memory.
// It returns the offset to the allocated memory and the size of the data.
func bytesToLeakedPtr(data []byte, offsetSize uint32) (offset uint64) {
	ptr := unsafe.Pointer(C.malloc(C.ulong(offsetSize)))
	copy(unsafe.Slice((*byte)(ptr), C.ulong(offsetSize)), data)
	return uint64(uintptr(ptr))
}

// byteToLeakedPtr allocates memory for a byte (uint8) and stores the value in that memory.
// It returns the offset to the allocated memory.
func byteToLeakedPtr(data byte) (offset uint64) {
	ptr := unsafe.Pointer(C.malloc(1))
	*(*byte)(ptr) = data

	return uint64(uintptr(ptr))
}

// uint64ToLeakedPtr allocates memory for a uint32 and stores the value in that memory.
// It returns the offset to the allocated memory.
func uint32ToLeakedPtr(data uint32) (offset uint64) {
	ptr := unsafe.Pointer(C.malloc(4))
	*(*uint32)(ptr) = data

	return uint64(uintptr(ptr))
}

// uint64ToLeakedPtr allocates memory for a uint64 and stores the value in that memory.
// It returns the offset to the allocated memory.
func uint64ToLeakedPtr(data uint64) (offset uint64) {
	ptr := unsafe.Pointer(C.malloc(4))
	*(*uint64)(ptr) = data

	return uint64(uintptr(ptr))
}

// float32ToLeakedPtr allocates memory for a float32 and stores the value in that memory.
// It returns the offset to the allocated memory.
func float32ToLeakedPtr(data float32) (offset uint64) {
	ptr := unsafe.Pointer(C.malloc(4))
	*(*float32)(ptr) = data

	return uint64(uintptr(ptr))
}

// float64ToLeakedPtr allocates memory for a float64 and stores the value in that memory.
// It returns the offset to the allocated memory.
func float64ToLeakedPtr(data float64) (offset uint64) {
	ptr := unsafe.Pointer(C.malloc(4))
	*(*float64)(ptr) = data

	return uint64(uintptr(ptr))
}

// stringToLeakedPtr allocates memory for a string and stores the value in that memory.
// It returns the offset to the allocated memory.
func stringToLeakedPtr(data string, offsetSize uint32) (offset uint64) {
	byteSlice := unsafe.Slice(unsafe.StringData(data), len(data))
	return bytesToLeakedPtr(byteSlice, offsetSize)
}

// ptrToData converts a given memory address (ptr) into a pointer of type T.
// This function uses unsafe operations to cast the provided uint64 pointer
// to a pointer of the desired type, allowing for direct memory access to the
// underlying data.
//
// return a pointer of type T pointing to the data at the specified memory address.
func ptrToData[T any](ptr uint64) *T {
	return (*T)(unsafe.Pointer(uintptr(ptr)))
}

// free deallocates the memory previously allocated by a call to malloc (C.malloc).
// The offset parameter is a uint64 representing the starting address of the block
// of linear memory to be deallocated.
func free(offset uint64) {
	C.free(unsafe.Pointer(uintptr(offset)))
}

func unpackDataAndCheckType(packedData PackedData, expectedType types.ValueType) (types.ValueType, uint32, uint32) {
	valueType, offsetU32, size := unpackUI64(uint64(packedData))
	if valueType != expectedType {
		LogError("Unexpected data type. Expected %s, but got %s", expectedType, valueType)
		return 0, 0, 0
	}
	return valueType, offsetU32, size
}

// multiPackedDataToBytes converts a slice of uint64 integers to a slice of bytes.
// This function is typically used to convert a slice of packed data into bytes,
// which can then be stored in linear memory.
func multiPackedDataToBytes(data []PackedData) []byte {
	// Calculate the total number of bytes required to represent all the uint64
	// integers in the data slice. Since each uint64 integer is 8 bytes long,
	// we multiply the number of uint64 integers by 8 to get the total number of bytes.
	size := len(data) * 8

	result := make([]byte, size)

	for i, d := range data {
		// Convert d to its little-endian byte representation and store it in the
		// result slice. The binary.LittleEndian.PutUint64 function takes a slice
		// of bytes and a uint64 integer, and writes the uint64 integer into the slice
		// of bytes in little-endian order.
		// The result[i<<3:] slice expression ensures that each uint64 integer is
		// written to the correct position in the result slice.
		// i<<3 is equivalent to i*8, but using bit shifting (<<3) is slightly more
		// efficient than multiplication.
		binary.LittleEndian.PutUint64(result[i<<3:], uint64(d))
	}

	// Return the result slice of bytes.
	return result
}

// packUI64 takes a data type (in the form of a byte), a pointer (offset in memory),
// and a size (amount of memory/data to consider). It returns a packed uint64 representation.
//
// Structure of the packed uint64:
// - Highest 8 bits: data type
// - Next 32 bits: offset
// - Lowest 24 bits: size
//
// This function will return error if the provided size is larger than what can be represented in 24 bits
// (i.e., larger than 16,777,215).
func packUI64(dataType types.ValueType, offset uint32, size uint32) uint64 {
	// Check if the size can be represented in 24 bits
	if size >= (1 << 24) {
		err := fmt.Errorf("Size %d exceeds 24 bits precision %d", size, (1 << 24))
		LogError(err.Error())
	}

	// Shift the dataType into the highest 8 bits
	// Shift the offset into the next 32 bits
	// Use the size as is, but ensure only the lowest 24 bits are used (using bitwise AND)
	return (uint64(dataType) << 56) | (uint64(offset) << 24) | uint64(size&0xFFFFFF)
}

// unpackUI64 reverses the operation done by packUI64.
// Given a packed uint64, it will extract and return the original dataType, offset (ptr), and size.
func unpackUI64(packedData uint64) (dataType types.ValueType, offset uint32, size uint32) {
	// Extract the dataType from the highest 8 bits
	dataType = types.ValueType(packedData >> 56)

	// Extract the offset (ptr) from the next 32 bits using bitwise AND to mask the other bits
	offset = uint32((packedData >> 24) & 0xFFFFFFFF)

	// Extract the size from the lowest 24 bits
	size = uint32(packedData & 0xFFFFFF)

	return
}

func unpackMultiPackedData(packedData MultiPackedData) (dataType types.ValueType, offset uint32, size uint32) {
	return unpackUI64(uint64(packedData))
}

func packBytes(offset uint32, size uint32) uint64 {
	return packUI64(types.ValueTypeBytes, offset, size)
}
func packByte(offset uint32) uint64 {
	return packUI64(types.ValueTypeByte, offset, 1)
}
func packI32(offset uint32) uint64 {
	return packUI64(types.ValueTypeI32, offset, 4)
}
func packI64(offset uint32) uint64 {
	return packUI64(types.ValueTypeI64, offset, 8)
}
func packF32(offset uint32) uint64 {
	return packUI64(types.ValueTypeF32, offset, 4)
}
func packF64(offset uint32) uint64 {
	return packUI64(types.ValueTypeF64, offset, 8)
}
func packString(offset uint32, size uint32) uint64 {
	return packUI64(types.ValueTypeString, offset, size)
}
