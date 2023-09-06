package wasify

import (
	"context"
	"errors"
	"fmt"
	"reflect"

	"github.com/tetratelabs/wazero/api"
	"github.com/wasify-io/wasify-go/mdk"
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
		m.log.Warn("exported function does not exist", "function", name, "module", m.Name)
	}

	return &wazeroGuestFunction{ctx, fn, name, m.ModuleConfig}
}

type wazeroGuestFunction struct {
	ctx  context.Context
	fn   api.Function
	name string
	*ModuleConfig
}

// Invoke calls the function with the given parameters and returns any
// results or an error for any failure looking up or invoking the function.
//
// If the function name is not "malloc" or "free", it logs the function call details.
// It omits logging for "malloc" and "free" functions due to potential high frequency,
// which could lead to excessive log entries and complicate debugging for host funcs.
func (gf *wazeroGuestFunction) Invoke(params ...uint64) ([]uint64, error) {

	if gf.name != "malloc" && gf.name != "free" {
		gf.log.Info("calling function", "name", gf.name, "module", gf.Name, "params", params)
	}

	// TODO: Use CallWithStack
	res, err := gf.fn.Call(gf.ctx, params...)
	if err != nil {
		err = errors.Join(errors.New("can't call guest function"), err)
		gf.log.Error(err.Error())
		return nil, err
	}

	return res, nil
}

// Memory retrieves a Memory instance associated with the wazeroModule.
func (r *wazeroModule) Memory() Memory {
	return &wazeroMemory{r}
}

type wazeroMemory struct {
	*wazeroModule
}

// TODO: Update Comment
// Read reads byteCount bytes from the underlying buffer at the offset or
//
// It unpacks the packedData to obtain offset and size information, then reads
// data from the memory at the specified offset and size.
// Returns the offset, size, read data, and any potential error like if out of range.
// Packed data is a uint64 where the first 32 bits represent the offset
// and the following 32 bits represent the size of the actual data to be read.
func (m *wazeroMemory) Read(packedData uint64) (uint32, uint32, any, error) {

	var err error
	var data any

	// Unpack the packedData to extract offset and size values.
	t, offset, size := mdk.UnpackUI64(packedData)

	switch ValueType(t) {
	case ValueTypeBytes:
		data, err = m.ReadBytes(offset, size)
	case ValueTypeI32:
		data, err = m.ReadUint32Le(offset)
	case ValueTypeI64:
		data, err = m.ReadUint64Le(offset)
	case ValueTypeF32:
		data, err = m.ReadFloat32Le(offset)
	case ValueTypeF64:
		data, err = m.ReadFloat64Le(offset)
	default:
		err = fmt.Errorf("Unsupported read data type %d", t)
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

func (m *wazeroMemory) ReadUint32Le(offset uint32) (uint32, error) {
	data, ok := m.mod.Memory().ReadUint32Le(offset)
	if !ok {
		err := fmt.Errorf("Memory.ReadUint32Le(%d, %d) out of range of memory size %d", offset, 4, m.Size())
		m.log.Error(err.Error())
		return 0, err
	}

	return data, nil
}

func (m *wazeroMemory) ReadUint64Le(offset uint32) (uint64, error) {
	data, ok := m.mod.Memory().ReadUint64Le(offset)
	if !ok {
		err := fmt.Errorf("Memory.ReadUint64Le(%d, %d) out of range of memory size %d", offset, 8, m.Size())
		m.log.Error(err.Error())
		return 0, err
	}

	return data, nil
}

func (m *wazeroMemory) ReadFloat32Le(offset uint32) (float32, error) {
	data, ok := m.mod.Memory().ReadFloat32Le(offset)
	if !ok {
		err := fmt.Errorf("Memory.ReadFloat32Le(%d, %d) out of range of memory size %d", offset, 4, m.Size())
		m.log.Error(err.Error())
		return 0, err
	}

	return data, nil
}

func (m *wazeroMemory) ReadFloat64Le(offset uint32) (float64, error) {
	data, ok := m.mod.Memory().ReadFloat64Le(offset)
	if !ok {
		err := fmt.Errorf("Memory.ReadFloat64Le(%d, %d) out of range of memory size %d", offset, 8, m.Size())
		m.log.Error(err.Error())
		return 0, err
	}

	return data, nil
}

// Write writes the provided value (v) to the memory buffer managed by the wazeroMemory instance,
// starting at the specified offset.
// If the type of v is unsupported, or if the operation attempts to write out of the buffer's range,
// an error will be returned.
func (m *wazeroMemory) Write(offset uint32, v any) error {
	var err error

	switch vTyped := v.(type) {
	case []byte:
		err = m.WriteBytes(offset, vTyped)
	case uint32:
		err = m.WriteUint32Le(offset, vTyped)
	case uint64:
		err = m.WriteUint64Le(offset, vTyped)
	case float32:
		err = m.WriteFloat32Le(offset, vTyped)
	case float64:
		err = m.WriteFloat64Le(offset, vTyped)
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

func (m *wazeroMemory) WriteUint32Le(offset uint32, v uint32) error {
	ok := m.mod.Memory().WriteUint32Le(offset, v)
	if !ok {
		err := fmt.Errorf(" Memory.WriteUint32Le(%d, %d) out of range of memory size %d", offset, 4, m.Size())
		m.log.Error(err.Error())
		return err
	}

	return nil
}

func (m *wazeroMemory) WriteUint64Le(offset uint32, v uint64) error {
	ok := m.mod.Memory().WriteUint64Le(offset, v)
	if !ok {
		err := fmt.Errorf(" Memory.WriteUint64Le(%d, %d) out of range of memory size %d", offset, 8, m.Size())
		m.log.Error(err.Error())
		return err
	}

	return nil
}

func (m *wazeroMemory) WriteFloat32Le(offset uint32, v float32) error {
	ok := m.mod.Memory().WriteFloat32Le(offset, v)
	if !ok {
		err := fmt.Errorf(" Memory.WriteFloat32Le(%d, %d) out of range of memory size %d", offset, 8, m.Size())
		m.log.Error(err.Error())
		return err
	}

	return nil
}

func (m *wazeroMemory) WriteFloat64Le(offset uint32, v float64) error {
	ok := m.mod.Memory().WriteFloat64Le(offset, v)
	if !ok {
		err := fmt.Errorf(" Memory.WriteFloat64Le(%d, %d) out of range of memory size %d", offset, 8, m.Size())
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
//
// Example usage:
//
//	text := "Wasify.io"
//	size := uint32(len(text))
//	offset, err := module.Memory().Malloc(size)
//	res, _ := module.GuestFunction(ctx, "guest_function_name").Invoke(offset)
//	_, _, data, _ := module.Memory().Read(res[0])
//	if err != nil {
//		panic(err)
//	}
//
// fmt.Println("DATA: ", string(data))
//
// Note: Always make sure to free memory after allocation.
func (m *wazeroMemory) Malloc(size uint32) (uint32, error) {

	mallocRes, err := m.wazeroModule.GuestFunction(m.wazeroModule.ctx, "malloc").Invoke(uint64(size))
	if err != nil {
		err = errors.Join(fmt.Errorf("can't invoke malloc function "), err)
		return 0, err
	}
	offset := uint32(mallocRes[0])

	return offset, nil
}

// Free releases the memory block at the specified offset in wazeroMemory.
// It invokes the "free" GuestFunction of the associated wazeroModule using the provided offset parameter.
// Returns any encountered error during the memory deallocation.
//
// In most cases, parameter `offset` is the value returned from Malloc func.
func (m *wazeroMemory) Free(offset uint32) error {

	_, err := m.wazeroModule.GuestFunction(m.ModuleConfig.ctx, "free").Invoke(uint64(offset))

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
