package types

import (
	"reflect"
	"testing"
)

func TestGetOffsetSizeAndDataTypeByConversion(t *testing.T) {
	tests := []struct {
		input       any
		expected    ValueType
		expectError bool
		size        uint32
	}{
		{[]byte{1, 2, 3, 4}, ValueTypeBytes, false, 4},
		{byte(1), ValueTypeByte, false, 1},
		{uint32(1234567890), ValueTypeI32, false, 4},
		{uint64(1234567890123456789), ValueTypeI64, false, 8},
		{float32(123.456), ValueTypeF32, false, 4},
		{float64(123.4567890123), ValueTypeF64, false, 8},
		{"TestString", ValueTypeString, false, 10},
		{struct{}{}, ValueType(0), true, 0},
		{-1, ValueType(0), true, 0},
		{int(1), ValueType(0), true, 0},
	}

	for _, tt := range tests {
		dataType, size, err := GetOffsetSizeAndDataTypeByConversion(tt.input)

		if tt.expectError {
			if err == nil {
				t.Errorf("Expected error for input type %s but got none", reflect.TypeOf(tt.input))
			}
			continue
		}
		if err != nil {
			t.Errorf("Unexpected error for input type %s: %v", reflect.TypeOf(tt.input), err)
			continue
		}
		if dataType != tt.expected {
			t.Errorf("For type %s, expected ValueType %d, got %d", reflect.TypeOf(tt.input), tt.expected, dataType)
		}
		if size != tt.size {
			t.Errorf("For type %s, expected size %v, got %v", reflect.TypeOf(tt.input), tt.size, size)
		}
	}
}
