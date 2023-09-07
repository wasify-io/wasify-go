package types

import (
	"fmt"
	"reflect"
)

// ValueType is an enumeration of supported data types for function parameters and returns.
type ValueType uint8

// ValueTypePack is a reserved ValueType used for packed data.
const ValueTypePack ValueType = 255

// These constants represent the possible data types that can be used in function parameters and returns.
const (
	ValueTypeBytes ValueType = iota
	ValueTypeByte
	ValueTypeI32
	ValueTypeI64
	ValueTypeF32
	ValueTypeF64
	ValueTypeString
)

// GetOffsetSizeAndDataTypeByConversion determines the memory size (offsetSize) and ValueType
// of a given data. The function supports several data
func GetOffsetSizeAndDataTypeByConversion(data any) (dataType ValueType, offsetSize uint32, err error) {

	switch vTyped := data.(type) {
	case []byte:
		offsetSize = uint32(len(vTyped))
		dataType = ValueTypeBytes
	case byte:
		offsetSize = 1
		dataType = ValueTypeByte
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
	case string:
		offsetSize = uint32(len(vTyped))
		dataType = ValueTypeString
	default:
		err = fmt.Errorf("unsupported conversion data type %s", reflect.TypeOf(vTyped))
		return
	}

	return dataType, offsetSize, err
}
