package utils

import (
	"fmt"

	"github.com/wasify-io/wasify-go/internal/types"
)

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
func PackUI64(dataType types.ValueType, offset uint32, size uint32) (uint64, error) {
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
func UnpackUI64(packedData uint64) (dataType types.ValueType, offset uint32, size uint32) {
	// Extract the dataType from the highest 8 bits
	dataType = types.ValueType(packedData >> 56)

	// Extract the offset (ptr) from the next 32 bits using bitwise AND to mask the other bits
	offset = uint32((packedData >> 24) & 0xFFFFFFFF)

	// Extract the size from the lowest 24 bits
	size = uint32(packedData & 0xFFFFFF)

	return
}
