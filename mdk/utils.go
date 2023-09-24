package mdk

// #include <stdlib.h>
// #include <string.h>
import "C"
import (
	"unsafe"

	"github.com/wasify-io/wasify-go/internal/types"
	"github.com/wasify-io/wasify-go/internal/utils"
)

func readBytes(packedData PackedData) ([]byte, uint32) {
	_, offsetU32, size := unpackDataAndCheckType(packedData, types.ValueTypeBytes)
	return unsafe.Slice(ptrToData[byte](uint64(offsetU32)), int(size)), size
}

func readByte(packedData PackedData) byte {
	_, offsetU32, _ := unpackDataAndCheckType(packedData, types.ValueTypeByte)
	return *ptrToData[byte](uint64(offsetU32))
}

func readI32(packedData PackedData) uint32 {
	_, offsetU32, _ := unpackDataAndCheckType(packedData, types.ValueTypeI32)
	return *ptrToData[uint32](uint64(offsetU32))
}

func readI64(packedData PackedData) uint64 {
	_, offsetU32, _ := unpackDataAndCheckType(packedData, types.ValueTypeI64)
	return *ptrToData[uint64](uint64(offsetU32))
}

func readF32(packedData PackedData) float32 {
	_, offsetU32, _ := unpackDataAndCheckType(packedData, types.ValueTypeF32)
	return *ptrToData[float32](uint64(offsetU32))
}

func readF64(packedData PackedData) float64 {
	_, offsetU32, _ := unpackDataAndCheckType(packedData, types.ValueTypeF64)
	return *ptrToData[float64](uint64(offsetU32))
}

func readString(packedData PackedData) (string, uint32) {
	_, offsetU32, size := unpackDataAndCheckType(packedData, types.ValueTypeString)
	data := unsafe.Slice(ptrToData[byte](uint64(offsetU32)), int(size))
	return string(data), size
}

func readAny(packedData PackedData) (any, uint32) {

	valueType, _, size := utils.UnpackUI64(uint64(packedData))
	var value any

	switch valueType {
	case types.ValueTypeBytes:
		value, _ = readBytes(packedData)
	case types.ValueTypeByte:
		value = readByte(packedData)
	case types.ValueTypeI32:
		value = readI32(packedData)
	case types.ValueTypeI64:
		value = readI64(packedData)
	case types.ValueTypeF32:
		value = readF32(packedData)
	case types.ValueTypeF64:
		value = readF64(packedData)
	case types.ValueTypeString:
		value, size = readString(packedData)
	default:
		LogError("can't read the datatype: %s", valueType)
		return nil, 0
	}

	return value, size
}

// free deallocates the memory previously allocated by a call to malloc (C.malloc).
// The offset parameter is a uint64 representing the starting address of the block
// of linear memory to be deallocated.
func free(offset uint64) {
	C.free(unsafe.Pointer(uintptr(offset)))
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

func unpackDataAndCheckType(packedData PackedData, expectedType types.ValueType) (types.ValueType, uint32, uint32) {
	valueType, offsetU32, size := utils.UnpackUI64(uint64(packedData))
	if valueType != expectedType {
		LogError("Unexpected data type. Expected %s, but got %s", expectedType, valueType)
		return 0, 0, 0
	}
	return valueType, offsetU32, size
}
