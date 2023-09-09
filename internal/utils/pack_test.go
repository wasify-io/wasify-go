package utils

import (
	"testing"

	"github.com/wasify-io/wasify-go/internal/types"
)

func TestPackUnpackUI64(t *testing.T) {
	dataType := types.ValueTypeString
	ptr := uint32(0x12345678)
	size := uint32(0x123456)

	packedData, err := PackUI64(dataType, ptr, size)
	if err != nil {
		t.Fatalf("Failed to pack data: %v", err)
	}

	unpackedDataType, unpackedPtr, unpackedSize := UnpackUI64(packedData)

	if unpackedDataType != dataType || unpackedPtr != ptr || unpackedSize != size {
		t.Errorf("Unpack did not match original data. Expected: %v, %v, %v. Got: %v, %v, %v",
			dataType, ptr, size, unpackedDataType, unpackedPtr, unpackedSize)
	}

	// Test case where size exceeds 24 bits of precision
	largeSize := uint32(1<<24 + 1) // This value is one greater than what can be represented in 24 bits

	_, err = PackUI64(dataType, ptr, largeSize)
	if err == nil {
		t.Errorf("Expected error due to size exceeding 24 bits of precision but got none")
	}
}
