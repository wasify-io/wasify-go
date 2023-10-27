package mdk

import (
	"fmt"
	"unsafe"

	"github.com/wasify-io/wasify-go/internal/types"
)

type PackedData uint64

func ReadBytesPack(pd PackedData) []byte {
	valueType, offsetU32, size := unpackDataAndCheckType(pd, types.ValueTypeBytes)
	if valueType != types.ValueTypeBytes {
		LogError("value type %s is not a type of %s", valueType, types.ValueTypeBytes)
		return nil
	}

	return readBytes(uint64(offsetU32), int(size))
}

func ReadBytePack(pd PackedData) byte {
	valueType, offsetU32, _ := unpackDataAndCheckType(pd, types.ValueTypeByte)
	if valueType != types.ValueTypeByte {
		LogError("value type %s is not a type of %s", valueType, types.ValueTypeByte)
		return 0
	}

	return readByte(uint64(offsetU32))
}

func ReadI32Pack(pd PackedData) uint32 {
	valueType, offsetU32, _ := unpackDataAndCheckType(pd, types.ValueTypeI32)
	if valueType != types.ValueTypeI32 {
		LogError("value type %s is not a type of %s", valueType, types.ValueTypeI32)
		return 0
	}

	return readI32(uint64(offsetU32))
}

func ReadI64Pack(pd PackedData) uint64 {
	valueType, offsetU32, _ := unpackDataAndCheckType(pd, types.ValueTypeI64)
	if valueType != types.ValueTypeI64 {
		LogError("value type %s is not a type of %s", valueType, types.ValueTypeI64)
		return 0
	}

	return readI64(uint64(offsetU32))
}

func ReadF32Pack(pd PackedData) float32 {
	valueType, offsetU32, _ := unpackDataAndCheckType(pd, types.ValueTypeF32)
	if valueType != types.ValueTypeF32 {
		LogError("value type %s is not a type of %s", valueType, types.ValueTypeF32)
		return 0
	}

	return readF32(uint64(offsetU32))
}

func ReadF64Pack(pd PackedData) float64 {
	valueType, offsetU32, _ := unpackDataAndCheckType(pd, types.ValueTypeF64)
	if valueType != types.ValueTypeF64 {
		LogError("value type %s is not a type of %s", valueType, types.ValueTypeF64)
		return 0
	}

	return readF64(uint64(offsetU32))
}

func ReadStringPack(pd PackedData) string {
	valueType, offsetU32, size := unpackDataAndCheckType(pd, types.ValueTypeString)
	if valueType != types.ValueTypeString {
		LogError("value type %s is not a type of %s", valueType, types.ValueTypeString)
		return ""
	}

	return readString(uint64(offsetU32), int(size))
}

func FreePack(pd PackedData) {
	_, offset, _ := unpackUI64(uint64(pd))
	free(uint64(offset))
}

func ReadBytes(offset uint64, size int) []byte {
	return readBytes(offset, size)
}

func ReadByte(offset uint64) byte {
	return readByte(offset)
}

func ReadI32(offset uint64) uint32 {
	return readI32(offset)
}

func ReadI64(offset uint64) uint64 {
	return readI64(offset)
}

func ReadF32(offset uint64) float32 {
	return readF32(offset)
}

func ReadF64(offset uint64) float64 {
	return readF64(offset)
}

func ReadString(offset uint64, size int) string {
	return readString(offset, size)
}

type MultiPackedData uint64

func (mpd *MultiPackedData) Read() []PackedData {

	if mpd == nil || *mpd == 0 {
		return nil
	}

	t, offsetU32, size := unpackMultiPackedData(*mpd)
	if t != types.ValueTypePack {
		err := fmt.Errorf("can't unpack guest data, the type is not a valueTypePack. expected %d, got %d", types.ValueTypePack, t)
		LogError("can't read results", err.Error())
		return nil
	}

	count := size / 8

	data := make([]PackedData, count)

	copy(data, unsafe.Slice(ptrToData[PackedData](uint64(offsetU32)), count))

	return data
}

func (mpd *MultiPackedData) Free() {

	// free offsets of each packed data in the array
	offsets := mpd.Read()
	for _, offset := range offsets {
		FreePack(offset)
	}

	// free self offset of multipackdata
	_, offset, _ := unpackUI64(uint64(*mpd))
	free(uint64(offset))
}

func WriteBytes(data []byte, offsetSize uint32) uint64 {
	return bytesToLeakedPtr(data, offsetSize)
}
func WriteByte(data byte) uint64 {
	return byteToLeakedPtr(data)
}
func WriteUint32(data uint32) uint64 {
	return uint32ToLeakedPtr(data)
}
func WriteUint64(data uint64) uint64 {
	return uint64ToLeakedPtr(data)
}
func WriteFloat32(data float32) uint64 {
	return float32ToLeakedPtr(data)
}
func WriteFloat64(data float64) uint64 {
	return float64ToLeakedPtr(data)
}
func WriteString(data string, offsetSize uint32) uint64 {
	return stringToLeakedPtr(data, offsetSize)
}

func WriteBytesPack(data []byte) PackedData {
	return PackedData(packBytes(uint32(WriteBytes(data, uint32(len(data)))), uint32(len(data))))
}
func WriteBytePack(data byte) PackedData {
	return PackedData(packByte(uint32(WriteByte(data))))
}
func WriteUint32Pack(data uint32) PackedData {
	return PackedData(packI32(uint32(WriteUint32(data))))
}
func WriteUint64Pack(data uint64) PackedData {
	return PackedData(packI64(uint32(WriteUint64(data))))
}
func WriteFloat32Pack(data float32) PackedData {
	return PackedData(packF32(uint32(WriteFloat32(data))))
}
func WriteFloat64Pack(data float64) PackedData {
	return PackedData(packF64(uint32(WriteFloat64(data))))
}
func WriteStringPack(data string) PackedData {
	return PackedData(packString(uint32(WriteString(data, uint32(len(data)))), uint32(len(data))))
}

// WriteMultiPack takes a variable number of parameters, packs them into a byte slice representation,
// allocates memory for the packed data, and then returns a MultiPackedData which represents the memory
// offset of the packed data. If any error occurs during the process, it logs the error and returns a MultiPackedData of 0.
//
// params ...any: A variable number of parameters that need to be packed.
//
// MultiPackedData: The offset in memory where the packed array data starts.
func WriteMultiPack(params ...PackedData) MultiPackedData {

	if len(params) == 0 {
		return 0
	}

	multiPackedDataArray := make([]PackedData, len(params))

	copy(multiPackedDataArray, params)

	packedBytes := multiPackedDataToBytes(multiPackedDataArray)
	packedByetesSize := uint32(len(packedBytes))

	return MultiPackedData(packUI64(types.ValueTypePack, uint32(WriteBytes(packedBytes, packedByetesSize)), packedByetesSize))
}
