package wasify

import (
	"context"
	"errors"
	"fmt"
	"reflect"

	"github.com/tetratelabs/wazero/api"
	"github.com/wasify-io/wasify-go/internal/types"
	"github.com/wasify-io/wasify-go/internal/utils"
)

// GuestFunction returns a GuestFunction instance associated with the wazeroModule.
// GuestFunction is used to work with exported function from this module.
//
// Example usage:
//
//	result, err = module.GuestFunction(ctx, "greet").Invoke("argument1", "argument2", 123)
//	if err != nil {
//	    slog.Error(err.Error())
//	}
func (m *wazeroModule) GuestFunction(ctx context.Context, name string) GuestFunction {

	fn := m.mod.ExportedFunction(name)
	if fn == nil {
		m.log.Warn("exported function does not exist", "function", name, "namespace", m.Namespace)
	}

	return &wazeroGuestFunction{
		ctx,
		fn,
		name,
		m.Memory(),
		m.ModuleConfig,
	}
}

// Close closes the resource.
//
// Note: The context parameter is used for value lookup, such as for
// logging. A canceled or otherwise done context will not prevent Close
// from succeeding.
func (m *wazeroModule) Close(ctx context.Context) error {
	err := m.mod.Close(ctx)
	if err != nil {
		err = errors.Join(errors.New("can't close module"), err)
		m.log.Error(err.Error())
		return err
	}
	return nil
}

// Memory retrieves a Memory instance associated with the wazeroModule.
func (r *wazeroModule) Memory() Memory {
	return &wazeroMemory{r}
}

type wazeroMemory struct {
	*wazeroModule
}

// The wazeroModule struct combines an instantiated wazero modul
// with the generic module configuration.
type wazeroModule struct {
	mod api.Module
	*ModuleConfig
}

// ReadAnyPack extracts and reads data from a packed memory location.
//
// Given a packed data representation, this function determines the type, offset, and size of the data to be read.
// It then reads the data from the specified offset and returns it.
//
// Returns:
// - offset: The memory location where the data starts.
// - size: The size or length of the data.
// - data: The actual extracted data of the determined type (i.e., byte slice, uint32, uint64, float32, float64).
// - error: An error if encountered (e.g., unsupported data type, out-of-range error).
func (m *wazeroMemory) ReadAnyPack(pd PackedData) (any, uint32, uint32, error) {

	var err error
	var data any

	// Unpack the packedData to extract offset and size values.
	valueType, offset, size := utils.UnpackUI64(uint64(pd))

	switch ValueType(valueType) {
	case ValueTypeBytes:
		data, err = m.ReadBytes(offset, size)
	case ValueTypeByte:
		data, err = m.ReadByte(offset)
	case ValueTypeI32:
		data, err = m.ReadUint32(offset)
	case ValueTypeI64:
		data, err = m.ReadUint64(offset)
	case ValueTypeF32:
		data, err = m.ReadFloat32(offset)
	case ValueTypeF64:
		data, err = m.ReadFloat64(offset)
	case ValueTypeString:
		data, err = m.ReadString(offset, size)
	default:
		err = fmt.Errorf("Unsupported read data type %s", valueType)
	}

	if err != nil {
		m.log.Error(err.Error())
		return nil, 0, 0, err
	}

	return data, offset, size, err
}
func (m *wazeroMemory) ReadBytes(offset uint32, size uint32) ([]byte, error) {
	buf, ok := m.mod.Memory().Read(offset, size)
	if !ok {
		err := fmt.Errorf("Memory.ReadBytes(%d, %d) out of range of memory size %d", offset, size, m.Size())
		m.log.Error(err.Error())
		return nil, err
	}

	return buf, nil
}
func (m *wazeroMemory) ReadBytesPack(pd PackedData) ([]byte, error) {
	_, offset, size := utils.UnpackUI64(uint64(pd))
	return m.ReadBytes(offset, size)
}

func (m *wazeroMemory) ReadByte(offset uint32) (byte, error) {
	buf, ok := m.mod.Memory().ReadByte(offset)
	if !ok {
		err := fmt.Errorf("Memory.ReadByte(%d, %d) out of range of memory size %d", offset, 1, m.Size())
		m.log.Error(err.Error())
		return 0, err
	}

	return buf, nil
}
func (m *wazeroMemory) ReadBytePack(pd PackedData) (byte, error) {
	_, offset, _ := utils.UnpackUI64(uint64(pd))
	return m.ReadByte(offset)
}

func (m *wazeroMemory) ReadUint32(offset uint32) (uint32, error) {
	data, ok := m.mod.Memory().ReadUint32Le(offset)
	if !ok {
		err := fmt.Errorf("Memory.ReadUint32(%d, %d) out of range of memory size %d", offset, 4, m.Size())
		m.log.Error(err.Error())
		return 0, err
	}

	return data, nil
}
func (m *wazeroMemory) ReadUint32Pack(pd PackedData) (uint32, error) {
	_, offset, _ := utils.UnpackUI64(uint64(pd))
	return m.ReadUint32(offset)
}

func (m *wazeroMemory) ReadUint64(offset uint32) (uint64, error) {
	data, ok := m.mod.Memory().ReadUint64Le(offset)
	if !ok {
		err := fmt.Errorf("Memory.ReadUint64(%d, %d) out of range of memory size %d", offset, 8, m.Size())
		m.log.Error(err.Error())
		return 0, err
	}

	return data, nil
}
func (m *wazeroMemory) ReadUint64Pack(pd PackedData) (uint64, error) {
	_, offset, _ := utils.UnpackUI64(uint64(pd))
	return m.ReadUint64(offset)
}

func (m *wazeroMemory) ReadFloat32(offset uint32) (float32, error) {
	data, ok := m.mod.Memory().ReadFloat32Le(offset)
	if !ok {
		err := fmt.Errorf("Memory.ReadFloat32(%d, %d) out of range of memory size %d", offset, 4, m.Size())
		m.log.Error(err.Error())
		return 0, err
	}

	return data, nil
}
func (m *wazeroMemory) ReadFloat32Pack(pd PackedData) (float32, error) {
	_, offset, _ := utils.UnpackUI64(uint64(pd))
	return m.ReadFloat32(offset)
}

func (m *wazeroMemory) ReadFloat64(offset uint32) (float64, error) {
	data, ok := m.mod.Memory().ReadFloat64Le(offset)
	if !ok {
		err := fmt.Errorf("Memory.ReadFloat64(%d, %d) out of range of memory size %d", offset, 8, m.Size())
		m.log.Error(err.Error())
		return 0, err
	}

	return data, nil
}
func (m *wazeroMemory) ReadFloat64Pack(pd PackedData) (float64, error) {
	_, offset, _ := utils.UnpackUI64(uint64(pd))
	return m.ReadFloat64(offset)
}

func (m *wazeroMemory) ReadString(offset uint32, size uint32) (string, error) {
	buf, err := m.ReadBytes(offset, size)
	if err != nil {
		return "", err
	}

	return string(buf), err
}
func (m *wazeroMemory) ReadStringPack(pd PackedData) (string, error) {
	_, offset, size := utils.UnpackUI64(uint64(pd))
	return m.ReadString(offset, size)
}

// WriteAny writes a value of type interface{} to the memory buffer managed by the wazeroMemory instance,
// starting at the given offset.
//
// The method identifies the type of the value and performs the appropriate write operation.
func (m *wazeroMemory) WriteAny(offset uint32, v any) error {
	var err error

	switch vTyped := v.(type) {
	case []byte:
		err = m.WriteBytes(offset, vTyped)
	case byte:
		err = m.WriteByte(offset, vTyped)
	case uint32:
		err = m.WriteUint32(offset, vTyped)
	case uint64:
		err = m.WriteUint64(offset, vTyped)
	case float32:
		err = m.WriteFloat32(offset, vTyped)
	case float64:
		err = m.WriteFloat64(offset, vTyped)
	case string:
		err = m.WriteString(offset, vTyped)
	default:
		err := fmt.Errorf("unsupported write data type %s", reflect.TypeOf(v))
		m.log.Error(err.Error())
		return err
	}

	return err
}
func (m *wazeroMemory) WriteBytes(offset uint32, v []byte) error {
	ok := m.mod.Memory().Write(offset, v)
	if !ok {
		err := fmt.Errorf("Memory.WriteBytes(%d, %d) out of range of memory size %d", offset, len(v), m.Size())
		m.log.Error(err.Error())
		return err
	}

	return nil
}

func (m *wazeroMemory) WriteBytesPack(v []byte) PackedData {

	size := uint32(len(v))

	offset, err := m.Malloc(size)
	if err != nil {
		m.log.Error(err.Error())
		return 0
	}

	err = m.WriteBytes(offset, v)
	if err != nil {
		m.log.Error(err.Error())
		return 0
	}

	pd, err := utils.PackUI64(types.ValueTypeBytes, offset, size)
	if err != nil {
		m.log.Error(err.Error())
		return 0
	}

	return PackedData(pd)
}

func (m *wazeroMemory) WriteByte(offset uint32, v byte) error {
	ok := m.mod.Memory().WriteByte(offset, v)
	if !ok {
		err := fmt.Errorf("Memory.WriteByte(%d, %d) out of range of memory size %d", offset, 1, m.Size())
		m.log.Error(err.Error())
		return err
	}

	return nil
}
func (m *wazeroMemory) WriteBytePack(v byte) PackedData {

	offset, err := m.Malloc(1)
	if err != nil {
		m.log.Error(err.Error())
		return 0
	}

	err = m.WriteByte(offset, v)
	if err != nil {
		m.log.Error(err.Error())
		return 0
	}

	pd, err := utils.PackUI64(types.ValueTypeByte, offset, 1)
	if err != nil {
		m.log.Error(err.Error())
		return 0
	}

	return PackedData(pd)
}

func (m *wazeroMemory) WriteUint32(offset uint32, v uint32) error {
	ok := m.mod.Memory().WriteUint32Le(offset, v)
	if !ok {
		err := fmt.Errorf("Memory.WriteUint32(%d, %d) out of range of memory size %d", offset, 4, m.Size())
		m.log.Error(err.Error())
		return err
	}

	return nil
}
func (m *wazeroMemory) WriteUint32Pack(v uint32) PackedData {

	offset, err := m.Malloc(4)
	if err != nil {
		m.log.Error(err.Error())
		return 0
	}

	err = m.WriteUint32(offset, v)
	if err != nil {
		m.log.Error(err.Error())
		return 0
	}

	pd, err := utils.PackUI64(types.ValueTypeI32, offset, 4)
	if err != nil {
		m.log.Error(err.Error())
		return 0
	}

	return PackedData(pd)
}

func (m *wazeroMemory) WriteUint64(offset uint32, v uint64) error {
	ok := m.mod.Memory().WriteUint64Le(offset, v)
	if !ok {
		err := fmt.Errorf("Memory.WriteUint64(%d, %d) out of range of memory size %d", offset, 8, m.Size())
		m.log.Error(err.Error())
		return err
	}

	return nil
}
func (m *wazeroMemory) WriteUint64Pack(v uint64) PackedData {

	offset, err := m.Malloc(8)
	if err != nil {
		m.log.Error(err.Error())
		return 0
	}

	err = m.WriteUint64(offset, v)
	if err != nil {
		m.log.Error(err.Error())
		return 0
	}

	pd, err := utils.PackUI64(types.ValueTypeI32, offset, 8)
	if err != nil {
		m.log.Error(err.Error())
		return 0
	}

	return PackedData(pd)
}

func (m *wazeroMemory) WriteFloat32(offset uint32, v float32) error {
	ok := m.mod.Memory().WriteFloat32Le(offset, v)
	if !ok {
		err := fmt.Errorf("Memory.WriteFloat32(%d, %d) out of range of memory size %d", offset, 8, m.Size())
		m.log.Error(err.Error())
		return err
	}

	return nil
}
func (m *wazeroMemory) WriteFloat32Pack(v float32) PackedData {

	offset, err := m.Malloc(4)
	if err != nil {
		m.log.Error(err.Error())
		return 0
	}

	err = m.WriteFloat32(offset, v)
	if err != nil {
		m.log.Error(err.Error())
		return 0
	}

	pd, err := utils.PackUI64(types.ValueTypeF32, offset, 4)
	if err != nil {
		m.log.Error(err.Error())
		return 0
	}

	return PackedData(pd)
}

func (m *wazeroMemory) WriteFloat64(offset uint32, v float64) error {
	ok := m.mod.Memory().WriteFloat64Le(offset, v)
	if !ok {
		err := fmt.Errorf("Memory.WriteFloat64(%d, %d) out of range of memory size %d", offset, 8, m.Size())
		m.log.Error(err.Error())
		return err
	}

	return nil
}
func (m *wazeroMemory) WriteFloat64Pack(v float64) PackedData {

	offset, err := m.Malloc(8)
	if err != nil {
		m.log.Error(err.Error())
		return 0
	}

	err = m.WriteFloat64(offset, v)
	if err != nil {
		m.log.Error(err.Error())
		return 0
	}

	pd, err := utils.PackUI64(types.ValueTypeF64, offset, 8)
	if err != nil {
		m.log.Error(err.Error())
		return 0
	}

	return PackedData(pd)
}

func (m *wazeroMemory) WriteString(offset uint32, v string) error {

	ok := m.mod.Memory().WriteString(offset, v)
	if !ok {
		err := fmt.Errorf("Memory.WriteString(%d, %d) out of range of memory size %d", offset, len(v), m.Size())
		m.log.Error(err.Error())
		return err
	}

	return nil
}
func (m *wazeroMemory) WriteStringPack(v string) PackedData {

	size := uint32(len(v))

	offset, err := m.Malloc(size)
	if err != nil {
		m.log.Error(err.Error())
		return 0
	}

	err = m.WriteString(offset, v)
	if err != nil {
		m.log.Error(err.Error())
		return 0
	}

	pd, err := utils.PackUI64(types.ValueTypeString, offset, size)
	if err != nil {
		m.log.Error(err.Error())
		return 0
	}

	return PackedData(pd)
}

func (m *wazeroMemory) WriteMultiPack(pds ...PackedData) MultiPackedData {

	size := uint32(len(pds)) * 8
	if size == 0 {
		return 0
	}

	offset, err := m.Malloc(size)
	if err != nil {
		return 0
	}

	pdsU64 := make([]uint64, size)
	for _, pd := range pds {
		pdsU64 = append(pdsU64, uint64(pd))
	}

	err = m.WriteBytes(offset, utils.Uint64ArrayToBytes(pdsU64))
	if err != nil {
		return 0
	}

	pd, err := utils.PackUI64(types.ValueTypeString, offset, size)
	if err != nil {
		return 0
	}

	return MultiPackedData(pd)
}

// Size returns the size in bytes available. e.g. If the underlying memory
// has 1 page: 65536
func (r *wazeroMemory) Size() uint32 {
	return r.mod.Memory().Size()
}

// Malloc allocates memory in wasm linear memory with the specified size.
//
// It invokes the "malloc" GuestFunction of the associated wazeroModule using the provided size parameter.
// Returns the allocated memory offset and any encountered error.
//
// Malloc allows memory allocation from within a host function or externally,
// returning the allocated memory offset to be used in a guest function.
// This can be helpful, for instance, when passing string data from the host to the guest.
//
// NOTE: Always make sure to free memory after allocation.
func (m *wazeroMemory) Malloc(size uint32) (uint32, error) {

	r, err := m.wazeroModule.GuestFunction(m.wazeroModule.ctx, "malloc").call(uint64(size))
	if err != nil {
		err = errors.Join(fmt.Errorf("can't invoke malloc function "), err)
		return 0, err
	}

	offset := uint32(r)

	return offset, nil
}

// Free releases the memory block at the specified offset in wazeroMemory.
// It invokes the "free" GuestFunction of the associated wazeroModule using the provided offset parameter.
// Returns any encountered error during the memory deallocation.
func (m *wazeroMemory) Free(offsets ...uint32) error {

	for _, offset := range offsets {
		_, err := m.wazeroModule.GuestFunction(m.ModuleConfig.ctx, "free").call(uint64(offset))
		if err != nil {
			err = errors.Join(fmt.Errorf("can't invoke free function"), err)
			return err
		}
	}

	return nil
}

func (m *wazeroMemory) FreePack(pds ...PackedData) error {

	for _, pd := range pds {
		_, offset, _ := utils.UnpackUI64(uint64(pd))
		if err := m.Free(offset); err != nil {
			return err
		}
	}

	return nil
}
