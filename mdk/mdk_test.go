package mdk

import (
	"testing"

	"github.com/wasify-io/wasify-go/internal/types"
	"github.com/wasify-io/wasify-go/internal/utils"
)

func TestArg(t *testing.T) {
	// Define a slice of test cases
	tests := []struct {
		name      string
		input     any
		valueType types.ValueType
		size      uint32
		wantPanic bool
	}{
		{
			name:      "Test with byte slice",
			input:     []byte{1, 2, 3},
			valueType: types.ValueTypeBytes,
			size:      3,
			wantPanic: false,
		},
		{
			name:      "Test with byte",
			input:     byte(5),
			valueType: types.ValueTypeByte,
			size:      1,
			wantPanic: false,
		},
		{
			name:      "Test with uint32",
			input:     uint32(123),
			valueType: types.ValueTypeI32,
			size:      4,
			wantPanic: false,
		},
		{
			name:      "Test with uint64",
			input:     uint64(123456789),
			valueType: types.ValueTypeI64,
			size:      8,
			wantPanic: false,
		},
		{
			name:      "Test with float32",
			input:     float32(1.11),
			valueType: types.ValueTypeF32,
			size:      4,
			wantPanic: false,
		},
		{
			name:      "Test with float64",
			input:     float64(1.111111),
			valueType: types.ValueTypeF64,
			size:      8,
			wantPanic: false,
		},
		{
			name:      "Test with string",
			input:     "hello",
			valueType: types.ValueTypeString,
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

			valueType, _, size := utils.UnpackUI64(uint64(packedData))

			if valueType != tt.valueType {
				t.Errorf("ValueType does not match. Want %d, got %d", tt.valueType, valueType)
			}

			if size != tt.size {
				t.Errorf("size does not match. Want %d, got %d", tt.size, size)
			}

		})
	}
}
