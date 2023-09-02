package wasify

import (
	"context"
	"errors"
	"fmt"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
	"github.com/wasify-io/wasify-go/mdk"
)

// getWazeroRuntime creates and returns a wazero runtime instance using the provided context and
// RuntimeConfig. It configures the runtime with specific settings and features.
func getWazeroRuntime(ctx context.Context, c *RuntimeConfig) *wazeroRuntime {
	// TODO: Add explanation of runtime setup and configuration.
	// Create a new wazero runtime instance with specified configuration options.
	runtime := wazero.NewRuntimeWithConfig(ctx, wazero.NewRuntimeConfig().
		WithCustomSections(true).
		WithCloseOnContextDone(true).
		WithCoreFeatures(api.CoreFeaturesV2).
		WithDebugInfoEnabled(true),
	)
	// Instantiate the runtime with the WASI snapshot preview1.
	wasi_snapshot_preview1.MustInstantiate(ctx, runtime)
	return &wazeroRuntime{runtime, c}
}

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

// Read reads byteCount bytes from the underlying buffer at the offset or
//
// It unpacks the packedData to obtain offset and size information, then reads
// data from the memory at the specified offset and size.
// Returns the offset, size, read data, and any potential error like if out of range.
// Packed data is a uint64 where the first 32 bits represent the offset
// and the following 32 bits represent the size of the actual data to be read.
func (m *wazeroMemory) Read(packedData uint64) (uint32, uint32, []byte, error) {

	var err error

	// Unpack the packedData to extract offset and size values.
	offset, size := mdk.UnpackUI64(packedData)

	// Read data from memory using the extracted offset and size.
	buf, ok := m.mod.Memory().Read(offset, size)

	if !ok {
		err = fmt.Errorf("Memory.Read(%d, %d) out of range of memory size %d", offset, size, m.Size())
		m.log.Error(err.Error())
		return 0, 0, nil, err
	}

	return offset, size, buf, err
}

// Write writes the slice to the underlying buffer at the offset or returns error if out of range.
func (m *wazeroMemory) Write(offset uint32, v []byte) error {
	ok := m.mod.Memory().Write(offset, v)
	if !ok {
		err := fmt.Errorf("Memory.Write(%d, %d) out of range of memory size %d", offset, len(v), m.Size())
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
func (mp *wazeroModuleProxy) Read(packedData uint64) (uint32, uint32, []byte, error) {
	return mp.wazeroModule.Memory().Read(packedData)
}
func (mp *wazeroModuleProxy) Write(offset uint32, data []byte) error {
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
//				[]byte("World"),
//			)
//		},
//		Returns: []wasify.ValueType{wasify.ValueTypeByte, wasify.ValueTypeByte},
//	},
func (mp *wazeroModuleProxy) Return(args ...Result) Results {
	returns := make(Results, len(args))
	for i, arg := range args {
		returns[i] = arg
	}
	return returns
}
