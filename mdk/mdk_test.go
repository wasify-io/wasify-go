package mdk

import (
	"testing"
)

func TestArg(t *testing.T) {
	// Define a slice of test cases
	tests := []struct {
		name      string
		input     any
		valueType ValueType
		size      uint32
		wantPanic bool
	}{
		{
			name:      "Test with byte slice",
			input:     []byte{1, 2, 3},
			valueType: ValueTypeBytes,
			size:      3,
			wantPanic: false,
		},
		{
			name:      "Test with byte",
			input:     byte(5),
			valueType: ValueTypeByte,
			size:      1,
			wantPanic: false,
		},
		{
			name:      "Test with uint32",
			input:     uint32(123),
			valueType: ValueTypeI32,
			size:      4,
			wantPanic: false,
		},
		{
			name:      "Test with uint64",
			input:     uint64(123456789),
			valueType: ValueTypeI64,
			size:      8,
			wantPanic: false,
		},
		{
			name:      "Test with float32",
			input:     float32(1.11),
			valueType: ValueTypeF32,
			size:      4,
			wantPanic: false,
		},
		{
			name:      "Test with float64",
			input:     float64(1.111111),
			valueType: ValueTypeF64,
			size:      8,
			wantPanic: false,
		},
		{
			name:      "Test with string",
			input:     "hello",
			valueType: ValueTypeString,
			size:      5,
			wantPanic: false,
		},
		{
			name:      "Test with unsupported type",
			input:     false,
			wantPanic: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil && !tt.wantPanic {
					t.Errorf("Arg() panicked unexpectedly: %v", r)
				} else if r == nil && tt.wantPanic {
					t.Error("Arg() did not panic as expected")
				}
			}()

			packedData := Arg(tt.input)

			valueType, _, size := UnpackUI64(uint64(packedData))

			if valueType != tt.valueType {
				t.Errorf("ValueType does not match. Want %d, got %d", tt.valueType, valueType)
			}

			if size != tt.size {
				t.Errorf("size does not match. Want %d, got %d", tt.size, size)
			}

		})
	}
}

func TestPackUnpackUI64(t *testing.T) {
	dataType := ValueTypeString
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
