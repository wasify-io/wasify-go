package mdk

import (
	"errors"
	"fmt"
	"runtime"
	"unsafe"

	"github.com/wasify-io/wasify-go/internal/types"
	"github.com/wasify-io/wasify-go/internal/utils"
)

type PackedData uint64
type MultiPackedData uint64

// Arg prepares data for passing as an argument to a host function in WebAssembly.
// It accepts a generic data input and returns an ArgData which packs the memory offset and size of the data.
// This function abstracts away the complexity of memory management and conversion for the user.
//
// The runtime.KeepAlive call is used to ensure that the 'value' object is not garbage collected
// until the function finishes execution.
//
// ⚠️ Note: The ArgData returned by the Arg function does not need to be manually deallocated.
// The memory management is handled on the host side, where the allocated memory is automatically deallocated.
func Arg(value any) PackedData {

	packedData, err := AllocPack(value)
	if err != nil {
		LogError("can't allocate memory", err.Error())
		return 0
	}

	runtime.KeepAlive(value)

	return packedData
}

// TODO: add comment
func ReadMultiPackData(packedDataArray MultiPackedData) []PackedData {

	if packedDataArray == 0 {
		return nil
	}

	t, offsetU32, size := utils.UnpackUI64(uint64(packedDataArray))

	if t != types.ValueTypePack {
		err := fmt.Errorf("can't unpack guest data, the type is not a valueTypePack. expected %d, got %d", types.ValueTypePack, t)
		LogError("can't read results", err.Error())
		return nil
	}

	// calculate the number of elements in the array
	count := size / 8

	// read the packed pointers and sizes from the array
	packedData := unsafe.Slice(ptrToData[PackedData](uint64(offsetU32)), count)

	data := make([]PackedData, count)

	copy(data, packedData)

	return data
}

func ReadBytes(packedData PackedData) []byte {
	_, offsetU32, size := unpackDataAndCheckType(packedData, types.ValueTypeBytes)
	return unsafe.Slice(ptrToData[byte](uint64(offsetU32)), int(size))
}

func ReadByte(packedData PackedData) byte {
	_, offsetU32, _ := unpackDataAndCheckType(packedData, types.ValueTypeByte)
	return *ptrToData[byte](uint64(offsetU32))
}

func ReadI32(packedData PackedData) uint32 {
	_, offsetU32, _ := unpackDataAndCheckType(packedData, types.ValueTypeI32)
	return *ptrToData[uint32](uint64(offsetU32))
}

func ReadI64(packedData PackedData) uint64 {
	_, offsetU32, _ := unpackDataAndCheckType(packedData, types.ValueTypeI64)
	return *ptrToData[uint64](uint64(offsetU32))
}

func ReadF32(packedData PackedData) float32 {
	_, offsetU32, _ := unpackDataAndCheckType(packedData, types.ValueTypeF32)
	return *ptrToData[float32](uint64(offsetU32))
}

func ReadF64(packedData PackedData) float64 {
	_, offsetU32, _ := unpackDataAndCheckType(packedData, types.ValueTypeF64)
	return *ptrToData[float64](uint64(offsetU32))
}

func ReadString(packedData PackedData) string {
	_, offsetU32, size := unpackDataAndCheckType(packedData, types.ValueTypeString)
	data := unsafe.Slice(ptrToData[byte](uint64(offsetU32)), int(size))
	return string(data)
}

// AllocPack prepares data for interaction with WebAssembly by allocating the necessary memory.
// It accepts a generic input and returns a uint64 value that combines the data type, memory offset, size.
func AllocPack(data any) (PackedData, error) {

	dataType, offsetSize, err := types.GetOffsetSizeAndDataTypeByConversion(data)
	if err != nil {
		return 0, err
	}

	var offset uint64

	switch dataType {
	case types.ValueTypeBytes:
		offset = AllocBytes(data.([]byte), offsetSize)
	case types.ValueTypeByte:
		offset = AllocByte(data.(byte))
	case types.ValueTypeI32:
		offset = AllocUint32(data.(uint32))
	case types.ValueTypeI64:
		offset = AllocUint64(data.(uint64))
	case types.ValueTypeF32:
		offset = AllocFloat32(data.(float32))
	case types.ValueTypeF64:
		offset = AllocFloat64(data.(float64))
	case types.ValueTypeString:
		offset = AllocString(data.(string), offsetSize)
	default:
		err = fmt.Errorf("unsupported data type %s for allocation", dataType)
		LogError(err.Error())
		return 0, err
	}

	// TODO: change this
	pd, err := utils.PackUI64(dataType, uint32(offset), offsetSize)
	return PackedData(pd), err
}

func AllocBytes(data []byte, offsetSize uint32) uint64 {
	return bytesToLeakedPtr(data, offsetSize)
}
func AllocByte(data byte) uint64 {
	return byteToLeakedPtr(data)
}
func AllocUint32(data uint32) uint64 {
	return uint32ToLeakedPtr(data)
}
func AllocUint64(data uint64) uint64 {
	return uint64ToLeakedPtr(data)
}
func AllocFloat32(data float32) uint64 {
	return float32ToLeakedPtr(data)
}
func AllocFloat64(data float64) uint64 {
	return float64ToLeakedPtr(data)
}
func AllocString(data string, offsetSize uint32) uint64 {
	return stringToLeakedPtr(data, offsetSize)
}

// Return takes a variable number of parameters, packs them into a byte slice representation,
// allocates memory for the packed data, and then returns a MultiPackedData which represents the memory
// offset of the packed data. If any error occurs during the process, it logs the error and returns a MultiPackedData of 0.
//
// params ...any: A variable number of parameters that need to be packed.
//
// MultiPackedData: The offset in memory where the packed array data starts.
func Return(params ...any) MultiPackedData {

	packedDataArray := make([]uint64, len(params))

	for i, p := range params {
		// allocate memory for each value
		offsetI64, err := AllocPack(p)
		if err != nil {
			err = errors.Join(fmt.Errorf("An error occurred while attempting to alloc memory for guest func return value %s", p), err)
			LogError(err.Error())
			return 0
		}

		packedDataArray[i] = uint64(offsetI64)
	}

	// TODO: change this
	packedBytes := utils.Uint64ArrayToBytes(packedDataArray)
	packedByetesSize := uint32(len(packedBytes))

	offsetI32 := uint32(AllocBytes(packedBytes, packedByetesSize))

	offsetI64, err := utils.PackUI64(types.ValueTypePack, offsetI32, packedByetesSize)
	if err != nil {
		err = errors.Join(fmt.Errorf("Can't pack guest return data %d", packedDataArray), err)
		LogError(err.Error())
		return 0
	}

	return MultiPackedData(offsetI64)
}

// Free frees the memory.
// // exported function
//
//	func greet(var1, var2 PackedData) {
//		defer Free(var1, var2)
//
// ...
// }
func Free(packedData ...PackedData) {
	for _, p := range packedData {
		_, offset, _ := utils.UnpackUI64(uint64(p))
		free(uint64(offset))
	}
}
