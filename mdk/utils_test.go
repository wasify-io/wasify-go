package mdk

import (
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
		readBack := unsafe.Slice(ptrToData[byte](offset), len(expected.([]byte)))
		for i, v := range readBack {
			if v != expected.([]byte)[i] {
				return false
			}
		}
		return true
	}

	readBackByte := func(offset uint64, expected any) bool {
		return *ptrToData[byte](offset) == expected.(byte)
	}

	readBackUint32 := func(offset uint64, expected any) bool {
		return *ptrToData[uint32](offset) == expected.(uint32)
	}

	readBackUint64 := func(offset uint64, expected any) bool {
		return *ptrToData[uint64](offset) == expected.(uint64)
	}

	readBackFloat32 := func(offset uint64, expected any) bool {
		return *ptrToData[float32](offset) == expected.(float32)
	}

	readBackFloat64 := func(offset uint64, expected any) bool {
		return *ptrToData[float64](offset) == expected.(float64)
	}

	readBackString := func(offset uint64, expected any) bool {
		len := len(expected.(string))
		return string(unsafe.Slice(ptrToData[byte](offset), len)) == expected.(string)
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
