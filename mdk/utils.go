package mdk

// #include <stdlib.h>
// #include <string.h>
import "C"
import "unsafe"

// free deallocates the memory previously allocated by a call to malloc (C.malloc).
// The offset parameter is a uint32 representing the starting address of the block
// of linear memory to be deallocated.
func free(offset uint32) {
	C.free(unsafe.Pointer(uintptr(offset)))
}

// bytesToLeakedPtr converts a byte slice to an offset and size pair.
func bytesToLeakedPtr(data []byte) (offset uint32, size uint32) {
	return dataToLeakedPtr(data)
}

// dataToLeakedPtr converts a byte slice to an offset and size pair.
// It allocates memory of size 'len(data)' and copies the data into this memory.
// It returns the offset to the allocated memory and the size of the data.
func dataToLeakedPtr(data []byte) (offset uint32, size uint32) {
	ptr := unsafe.Pointer(C.malloc(C.ulong(len(data))))
	copy(unsafe.Slice((*byte)(ptr), C.ulong(len(data))), data)
	return uint32(uintptr(ptr)), uint32(len(data))
}
