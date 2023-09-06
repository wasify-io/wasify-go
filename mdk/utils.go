package mdk

// #include <stdlib.h>
// #include <string.h>
import "C"
import (
	"fmt"
	"reflect"
	"unsafe"
)

// free deallocates the memory previously allocated by a call to malloc (C.malloc).
// The offset parameter is a uint32 representing the starting address of the block
// of linear memory to be deallocated.
func free(offset uint32) {
	C.free(unsafe.Pointer(uintptr(offset)))
}

// bytesToLeakedPtr converts a byte slice to an offset and size pair.
// It allocates memory of size 'len(data)' and copies the data into this memory.
// It returns the offset to the allocated memory and the size of the data.
func bytesToLeakedPtr(data []byte, offsetSize uint32) (offset uint32) {
	ptr := unsafe.Pointer(C.malloc(C.ulong(offsetSize)))
	copy(unsafe.Slice((*byte)(ptr), C.ulong(offsetSize)), data)

	return uint32(uintptr(ptr))
}

// uint32ToLeakedPtr allocates memory for a uint32 and stores the value in that memory.
// It returns the offset to the allocated memory.
func uint32ToLeakedPtr(value uint32) (offset uint32) {
	ptr := unsafe.Pointer(C.malloc(4))
	*(*uint32)(ptr) = value

	return uint32(uintptr(ptr))
}

// uint64ToLeakedPtr allocates memory for a uint64 and stores the value in that memory.
// It returns the offset to the allocated memory.
func uint64ToLeakedPtr(value uint64) (offset uint32) {
	ptr := unsafe.Pointer(C.malloc(4))
	*(*uint64)(ptr) = value

	return uint32(uintptr(ptr))
}

// float32ToLeakedPtr allocates memory for a float32 and stores the value in that memory.
// It returns the offset to the allocated memory.
func float32ToLeakedPtr(value float32) (offset uint32) {
	ptr := unsafe.Pointer(C.malloc(4))
	*(*float32)(ptr) = value

	return uint32(uintptr(ptr))
}

// float64ToLeakedPtr allocates memory for a float64 and stores the value in that memory.
// It returns the offset to the allocated memory.
func float64ToLeakedPtr(value float64) (offset uint32) {
	ptr := unsafe.Pointer(C.malloc(4))
	*(*float64)(ptr) = value

	return uint32(uintptr(ptr))
}

// GetOffsetSizeAndDataTypeByConversion determines the memory size (offsetSize) and ValueType
// of a given data. The function supports several data types.
func GetOffsetSizeAndDataTypeByConversion(data any) (dataType ValueType, offsetSize uint32, err error) {

	switch vTyped := data.(type) {
	case []byte:
		offsetSize = uint32(len(vTyped))
		dataType = ValueTypeBytes
	case uint32:
		offsetSize = 4
		dataType = ValueTypeI32
	case uint64:
		offsetSize = 8
		dataType = ValueTypeI64
	case float32:
		offsetSize = 4
		dataType = ValueTypeF32
	case float64:
		offsetSize = 8
		dataType = ValueTypeF64
	default:
		err = fmt.Errorf("unsupported data type %s", reflect.TypeOf(data))
		return
	}

	return dataType, offsetSize, err
}
