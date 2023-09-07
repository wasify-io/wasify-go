package mdk

import (
	"reflect"
	"testing"
	"unsafe"
)

func TestToLeakedPtr(t *testing.T) {
	type testCase struct {
		name       string
		fn         func(any) uint64
		input      any
		expected   any
		readBackFn func(uint64, any) bool
	}

	readBackBytes := func(offset uint64, expected any) bool {
		readBack := unsafe.Slice((*byte)(unsafe.Pointer(uintptr(offset))), len(expected.([]byte)))
		for i, v := range readBack {
			if v != expected.([]byte)[i] {
				return false
			}
		}
		return true
	}

	readBackByte := func(offset uint64, expected any) bool {
		return byte(*(*byte)(unsafe.Pointer(uintptr(offset)))) == expected.(byte)
	}

	readBackUint32 := func(offset uint64, expected any) bool {
		return *(*uint32)(unsafe.Pointer(uintptr(offset))) == expected.(uint32)
	}

	readBackUint64 := func(offset uint64, expected any) bool {
		return *(*uint64)(unsafe.Pointer(uintptr(offset))) == expected.(uint64)
	}

	readBackFloat32 := func(offset uint64, expected any) bool {
		return *(*float32)(unsafe.Pointer(uintptr(offset))) == expected.(float32)
	}

	readBackFloat64 := func(offset uint64, expected any) bool {
		return *(*float64)(unsafe.Pointer(uintptr(offset))) == expected.(float64)
	}

	readBackString := func(offset uint64, expected any) bool {
		len := len(expected.(string))
		return string(unsafe.Slice((*byte)(unsafe.Pointer(uintptr(offset))), len)) == expected.(string)
	}

	tests := []testCase{
		{
			name:       "Bytes",
			fn:         func(data any) uint64 { return bytesToLeakedPtr(data.([]byte), uint32(len(data.([]byte)))) },
			input:      []byte{1, 2, 3},
			expected:   []byte{1, 2, 3},
			readBackFn: readBackBytes,
		},
		{
			name:       "Byte",
			fn:         func(data any) uint64 { return byteToLeakedPtr(data.(byte)) },
			input:      byte(65),
			expected:   byte(65),
			readBackFn: readBackByte,
		},
		{
			name:       "Uint32",
			fn:         func(data any) uint64 { return uint32ToLeakedPtr(data.(uint32)) },
			input:      uint32(1234567890),
			expected:   uint32(1234567890),
			readBackFn: readBackUint32,
		},
		{
			name:       "Uint64",
			fn:         func(data any) uint64 { return uint64ToLeakedPtr(data.(uint64)) },
			input:      uint64(1234567890123456789),
			expected:   uint64(1234567890123456789),
			readBackFn: readBackUint64,
		},
		{
			name:       "Float32",
			fn:         func(data any) uint64 { return float32ToLeakedPtr(data.(float32)) },
			input:      float32(123.456),
			expected:   float32(123.456),
			readBackFn: readBackFloat32,
		},
		{
			name:       "Float64",
			fn:         func(data any) uint64 { return float64ToLeakedPtr(data.(float64)) },
			input:      float64(123.4567890123),
			expected:   float64(123.4567890123),
			readBackFn: readBackFloat64,
		},
		{
			name:       "String",
			fn:         func(data any) uint64 { return stringToLeakedPtr(data.(string), uint32(len(data.(string)))) },
			input:      "TestString",
			expected:   "TestString",
			readBackFn: readBackString,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			offset1 := tt.fn(tt.input)

			if offset1 == 0 {
				t.Error("Expected non-zero offset")
			}

			if !tt.readBackFn(offset1, tt.expected) {
				t.Errorf("For type %T, expected %v, got different value", tt.input, tt.expected)
			}

			free(offset1)

			offset2 := tt.fn(tt.input)

			if offset1 != offset2 {
				t.Error("Expected both offsets to be the same, because first offset is freed up")
			}

			free(offset2)
		})
	}
}

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
