package wasify

import (
	"context"
	"errors"
	"fmt"
	"reflect"

	"github.com/tetratelabs/wazero/api"
	"github.com/wasify-io/wasify-go/internal/memory"
	"github.com/wasify-io/wasify-go/internal/types"
	"github.com/wasify-io/wasify-go/internal/utils"
)

// The wazeroModule struct combines an instantiated wazero modul
// with the generic module configuration.
type wazeroModule struct {
	mod api.Module
	*ModuleConfig
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

// GuestFunction returns a GuestFunction instance associated with the wazeroModule.
// GuestFunction is used to work with exported function from this module.
//
// Example usage:
//
//	result, err = module.GuestFunction(ctx, "greet").Invoke()
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
		memory.NewAllocationMap[uint32, uint32](),
		m.ModuleConfig,
	}
}

type wazeroGuestFunction struct {
	ctx    context.Context
	fn     api.Function
	name   string
	memory Memory
	// Allocation map to track parameter and return value allocations for host func.
	allocationMap *memory.AllocationMap[uint32, uint32]
	moduleConfig  *ModuleConfig
}

// call invokes wazero's CallWithStack method, which returns ome uint64 message,
// in most cases it is used to call built in methods such as "malloc", "free"
// See wazero's CallWithStack for more details.
func (gf *wazeroGuestFunction) call(params ...uint64) (uint64, error) {

	// size of params len(params) + one size for return uint64 value
	stack := make([]uint64, len(params)+1)
	copy(stack, params)

	err := gf.fn.CallWithStack(gf.ctx, stack[:])
	if err != nil {
		err = errors.Join(fmt.Errorf("An error occurred while attempting to invoke the guest function %s", gf.name), err)
		gf.moduleConfig.log.Error(err.Error())
		return 0, err
	}

	return stack[0], nil
}

// TODO: update comment
func (gf *wazeroGuestFunction) Invoke(params ...any) (uint64, error) {

	var err error

	log := gf.moduleConfig.log.Info
	if gf.moduleConfig.Namespace == "malloc" || gf.moduleConfig.Namespace == "free" {
		log = gf.moduleConfig.log.Debug
	}

	log("calling guest function", "namespace", gf.moduleConfig.Namespace, "function", gf.name, "params", params)

	defer func() {
		err = gf.cleanup()
	}()

	stack := make([]uint64, len(params))

	for i, p := range params {
		// get offset size and result value type (ValueType) by result's returnValue
		valueType, offsetSize, err := types.GetOffsetSizeAndDataTypeByConversion(p)
		if err != nil {
			err = errors.Join(fmt.Errorf("Can't convert guest func param %s", gf.name), err)
			return 0, err
		}

		// allocate memory for each value
		offsetI32, err := gf.memory.Malloc(offsetSize)
		if err != nil {
			err = errors.Join(fmt.Errorf("An error occurred while attempting to alloc memory for guest func param in: %s", gf.name), err)
			gf.moduleConfig.log.Error(err.Error())
			return 0, err
		}

		gf.allocationMap.Store(offsetI32, offsetSize)

		err = gf.memory.Write(offsetI32, p)
		if err != nil {
			err = errors.Join(errors.New("can't write arg to"), err)
			return 0, err
		}

		stack[i], err = utils.PackUI64(valueType, offsetI32, offsetSize)
		if err != nil {
			err = errors.Join(fmt.Errorf("An error occurred while attempting to pack data for guest func param in:  %s", gf.name), err)
			gf.moduleConfig.log.Error(err.Error())
			return 0, err
		}

	}

	res, err := gf.call(stack...)
	if err != nil {
		err = errors.Join(fmt.Errorf("An error occurred while attempting to invoke the guest function: %s", gf.name), err)
		gf.moduleConfig.log.Error(err.Error())
		return 0, err
	}

	return res, err
}

// cleanup releases the memory segments that were reserved by the
// guest function for parameter passing and result retrieval.
// finishes freeing memory and returns first error (if it exists).
func (gf *wazeroGuestFunction) cleanup() error {

	var firstErr error

	totalSize := gf.allocationMap.TotalSize()

	gf.allocationMap.Range(func(key, value uint32) bool {
		err := gf.memory.Free(key)
		if firstErr == nil {
			firstErr = err
		}

		gf.allocationMap.Delete(key)
		return true
	})

	gf.moduleConfig.log.Debug(
		"cleanup: guest func params and results",
		"allocated_bytes", totalSize,
		"after_deallocate_bytes", gf.allocationMap.TotalSize(),
		"namespace", gf.moduleConfig.Namespace,
		"func", gf.name,
	)

	return firstErr

}

// Memory retrieves a Memory instance associated with the wazeroModule.
func (r *wazeroModule) Memory() Memory {
	return &wazeroMemory{r}
}

type wazeroMemory struct {
	*wazeroModule
}

// Read extracts and reads data from a packed memory location.
//
// Given a packed data representation, this function determines the type, offset, and size of the data to be read.
// It then reads the data from the specified offset and returns it.
//
// Returns:
// - offset: The memory location where the data starts.
// - size: The size or length of the data.
// - data: The actual extracted data of the determined type (i.e., byte slice, uint32, uint64, float32, float64).
// - error: An error if encountered (e.g., unsupported data type, out-of-range error).
func (m *wazeroMemory) Read(packedData uint64) (uint32, uint32, any, error) {

	var err error
	var data any

	// Unpack the packedData to extract offset and size values.
	valueType, offset, size := utils.UnpackUI64(packedData)

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
		err = fmt.Errorf("Unsupported read data type %d", valueType)
	}

	if err != nil {
		m.log.Error(err.Error())
		return 0, 0, nil, err
	}

	return offset, size, data, err
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

func (m *wazeroMemory) ReadByte(offset uint32) (byte, error) {
	buf, ok := m.mod.Memory().ReadByte(offset)
	if !ok {
		err := fmt.Errorf("Memory.ReadByte(%d, %d) out of range of memory size %d", offset, 1, m.Size())
		m.log.Error(err.Error())
		return 0, err
	}

	return buf, nil
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

func (m *wazeroMemory) ReadUint64(offset uint32) (uint64, error) {
	data, ok := m.mod.Memory().ReadUint64Le(offset)
	if !ok {
		err := fmt.Errorf("Memory.ReadUint64(%d, %d) out of range of memory size %d", offset, 8, m.Size())
		m.log.Error(err.Error())
		return 0, err
	}

	return data, nil
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

func (m *wazeroMemory) ReadFloat64(offset uint32) (float64, error) {
	data, ok := m.mod.Memory().ReadFloat64Le(offset)
	if !ok {
		err := fmt.Errorf("Memory.ReadFloat64(%d, %d) out of range of memory size %d", offset, 8, m.Size())
		m.log.Error(err.Error())
		return 0, err
	}

	return data, nil
}

func (m *wazeroMemory) ReadString(offset uint32, size uint32) (string, error) {
	buf, err := m.ReadBytes(offset, size)
	if err != nil {
		return "", err
	}

	return string(buf), err
}

// Write writes a value of type interface{} to the memory buffer managed by the wazeroMemory instance,
// starting at the given offset.
//
// The method identifies the type of the value and performs the appropriate write operation.
func (m *wazeroMemory) Write(offset uint32, v any) error {
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
		err := fmt.Errorf(" Memory.WriteBytes(%d, %d) out of range of memory size %d", offset, len(v), m.Size())
		m.log.Error(err.Error())
		return err
	}

	return nil
}

func (m *wazeroMemory) WriteByte(offset uint32, v byte) error {
	ok := m.mod.Memory().WriteByte(offset, v)
	if !ok {
		err := fmt.Errorf(" Memory.WriteByte(%d, %d) out of range of memory size %d", offset, 1, m.Size())
		m.log.Error(err.Error())
		return err
	}

	return nil
}

func (m *wazeroMemory) WriteUint32(offset uint32, v uint32) error {
	ok := m.mod.Memory().WriteUint32Le(offset, v)
	if !ok {
		err := fmt.Errorf(" Memory.WriteUint32(%d, %d) out of range of memory size %d", offset, 4, m.Size())
		m.log.Error(err.Error())
		return err
	}

	return nil
}

func (m *wazeroMemory) WriteUint64(offset uint32, v uint64) error {
	ok := m.mod.Memory().WriteUint64Le(offset, v)
	if !ok {
		err := fmt.Errorf(" Memory.WriteUint64(%d, %d) out of range of memory size %d", offset, 8, m.Size())
		m.log.Error(err.Error())
		return err
	}

	return nil
}

func (m *wazeroMemory) WriteFloat32(offset uint32, v float32) error {
	ok := m.mod.Memory().WriteFloat32Le(offset, v)
	if !ok {
		err := fmt.Errorf(" Memory.WriteFloat32(%d, %d) out of range of memory size %d", offset, 8, m.Size())
		m.log.Error(err.Error())
		return err
	}

	return nil
}

func (m *wazeroMemory) WriteFloat64(offset uint32, v float64) error {
	ok := m.mod.Memory().WriteFloat64Le(offset, v)
	if !ok {
		err := fmt.Errorf(" Memory.WriteFloat64(%d, %d) out of range of memory size %d", offset, 8, m.Size())
		m.log.Error(err.Error())
		return err
	}

	return nil
}

func (m *wazeroMemory) WriteString(offset uint32, v string) error {
	ok := m.mod.Memory().WriteString(offset, v)
	if !ok {
		err := fmt.Errorf(" Memory.WriteString(%d, %d) out of range of memory size %d", offset, len(v), m.Size())
		m.log.Error(err.Error())
		return err
	}

	return nil
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
// Note: Always make sure to free memory after allocation.
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
//
// In most cases, parameter `offset` is the value returned from Malloc func.
func (m *wazeroMemory) Free(offset uint32) error {

	_, err := m.wazeroModule.GuestFunction(m.ModuleConfig.ctx, "free").call(uint64(offset))

	if err != nil {
		err = errors.Join(fmt.Errorf("can't invoke free function"), err)
		return err
	}

	return err
}

// wazeroModuleProxy is a proxy structure for wazeroModule.
// It is used to limit access to specific methods of wazeroModule within the host function context,
// such as module closing and other operations.
// Below is the list of available operations within the host.
type wazeroModuleProxy struct {
	*wazeroModule
}

func (mp *wazeroModuleProxy) GuestFunction(ctx context.Context, name string) GuestFunction {
	return mp.wazeroModule.GuestFunction(ctx, name)
}
func (mp *wazeroModuleProxy) Read(packedData uint64) (uint32, uint32, any, error) {
	return mp.wazeroModule.Memory().Read(packedData)
}
func (mp *wazeroModuleProxy) Write(offset uint32, data any) error {
	return mp.wazeroModule.Memory().Write(offset, data)
}
func (mp *wazeroModuleProxy) Size() uint32 {
	return mp.wazeroModule.Memory().Size()
}
func (mp *wazeroModuleProxy) Malloc(size uint32) (uint32, error) {
	return mp.wazeroModule.Memory().Malloc(size)
}
func (mp *wazeroModuleProxy) Free(offset uint32) error {
	return mp.wazeroModule.Memory().Free(offset)
}

// Return constructs and returns a set of ReturnValues using the provided ReturnValue arguments.
// This method is used to create the return values in the host function,
// that will be passed back to the WebAssembly module.
//
// Example usage:
//
//	{
//		Name: "my_host_func",
//		Callback: func(ctx context.Context, m wasify.ModuleProxy, params wasify.Params) wasify.ReturnValues {
//			// ...
//			return m.Return(
//				[]byte("Hello"),
//				590110011,
//				"Hello string",
//			)
//		},
//	},
func (mp *wazeroModuleProxy) Return(args ...Result) *Results {
	returns := make(Results, len(args))
	copy(returns, args)
	return &returns
}
