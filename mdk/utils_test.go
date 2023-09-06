package mdk

import (
	"reflect"
	"testing"
)

func TestGetOffsetSizeAndDataTypeByConversion(t *testing.T) {
	tests := []struct {
		input       any
		expected    ValueType
		expectError bool
	}{
		{[]byte{1, 2, 3, 4}, ValueTypeBytes, false},
		{byte(1), ValueTypeByte, false},
		{uint32(1234567890), ValueTypeI32, false},
		{uint64(1234567890123456789), ValueTypeI64, false},
		{float32(123.456), ValueTypeF32, false},
		{float64(123.4567890123), ValueTypeF64, false},
		{"TestString", ValueTypeString, false},
		{struct{}{}, ValueType(0), true},
		{-1, ValueType(0), true},
		{int(1), ValueType(0), true},
	}

	for _, tt := range tests {
		dataType, _, err := GetOffsetSizeAndDataTypeByConversion(tt.input)

		if tt.expectError == true && err == nil {
			t.Errorf("Expected error for input type %s", reflect.TypeOf(tt.input))
		} else if dataType != tt.expected {
			t.Errorf("For type %s, expected ValueType %d, got %d", reflect.TypeOf(tt.input), tt.expected, dataType)
		}
	}
}
