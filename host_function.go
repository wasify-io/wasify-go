package wasify

import (
	"context"
	"errors"
	"fmt"

	"github.com/wasify-io/wasify-go/mdk"
)

// ValueType represents the type of value used in function parameters and returns.
type ValueType uint8

// reserved value type for packedData
const valueTypePack uint8 = 255

// supported value types in params and returns
const (
	ValueTypeBytes ValueType = iota
	ValueTypeI32
	ValueTypeI64
	ValueTypeF32
	ValueTypeF64
)

// Param defines the attributes of a function parameter.
type Param struct {
	// Offset within memory.
	Offset uint32

	// Size of the data.
	Size uint32

	// Actual data value, passed from guest function as an argument.
	Value any
}
type Params []Param

// ReturnValue represents the value returned from a function.
type Result any
type Results []Result

// HostFunction defines a host function that can be invoked from a guest module.
type HostFunction struct {
	// Callback function to execute when the host function is invoked.
	Callback HostFunctionCallback

	// Name of the host function.
	Name string

	// Params specifies the types of parameters that the host function expects.
	//
	// The length of 'Params' should match the expected number of arguments
	// from the host function when called from the guest.
	Params []ValueType

	// Returns specifies the types of values that the host function returns.
	//
	// The length of 'Returns' should match the expected number of returns
	// from the host function as used in the guest.
	Returns []ValueType

	// Allocation map to track parameter and return value allocations for host func.
	allocationMap *allocationMap[uint32, uint32]

	// Configuration of the associated module.
	moduleConfig *ModuleConfig
}

// HostFunctionCallback is the function signature for the callback executed by a host function.
//
// HostFunctionCallback encapsulates the runtime's internal implementation details.
// It serves as an intermediary invoked between the processing of function parameters and the final return of the function.
type HostFunctionCallback func(ctx context.Context, moduleProxy ModuleProxy, stackParams Params) *Results

// convertParamsToStruct converts the packed stack parameters to a structured format.
// It uses the ModuleProxy instance to read data for each parameter from memory,
// creating a Params slice containing information about each parameter's offset, size, and value.
// Additionally, it stores allocation information in the host function's allocationMap.
//
// convertParamsToStruct simplifies the process of reading data by using structured information,
// allowing for easier access to parameter data instead of dealing with memory stacks and offsets.
func (hf *HostFunction) convertParamsToStruct(ctx context.Context, m ModuleProxy, stackParams []uint64) (Params, error) {

	// If user did not define params, skip the whole process, we still might get stackParams[0] = 0
	if len(hf.Params) == 0 {
		return nil, nil
	}

	if len(hf.Params) != len(stackParams) {
		return nil, fmt.Errorf("%s: params mismatch expected: %d received: %d ", hf.Name, len(hf.Params), len(stackParams))
	}

	params := make(Params, len(hf.Params))

	for i := range hf.Params {

		packedData := &stackParams[i]

		offset, offsetSize, data, err := m.Read(*packedData)
		if err != nil {
			err = errors.Join(errors.New("can't read params packed data"), err)
			return nil, err
		}

		params[i] = Param{
			Offset: offset,
			Size:   offsetSize,
			Value:  data,
		}

		hf.allocationMap.store(offset, offsetSize)

	}

	return params, nil

}

// writeResultsToMemory allocates memory to store return values and their offsets,
// and then writes the return values to memory using the ModuleProxy instance.
// It also packs and returns the data as packedDatas and the returnOffsets map.
//
// writeResultsToMemory handles the process as follows:
// It gathers all the returned values, allocates memory for each value based on its type and size,
// then compiles these individual allocations into a single uint64 called packedData and appends it to a slice.
// This packedData now contains three pieces of information:
// the first 8 bits (dataType) indicate the type of data (byte, uint32, etc.),
// the next 32 bits indicate the data's offset in memory,
// and the subsequent 24 bits represent the length or size of the data.
//
// The function then allocates memory for this slice, containing all the packedData entries, based on the slice's length.
// This results in a new packedData slice, which is stored in linear memory.
// The guest function can read this packedData slice, unpack it, and extract the required information for each item.
//
// +-----------------------------------------------------------------------------+
// |                              Packing and Storing Data                       |
// +-----------------------------------------------------------------------------+
// | Step             | Description                                              |
// +------------------+----------------------------------------------------------+
// | Data Collection  | Collect the returned values from the function.           |
// |                  | For example: V1, V2, V3, ...                             |
// +------------------+----------------------------------------------------------+
// | Data Allocation  | Allocate memory for each value based on its type & size. |
// |                  | Memory spaces are defined for different data types.      |
// +------------------+----------------------------------------------------------+
// | Packing Logic    | The packed data format (each of 64 bits or 8 bytes):     |
// |                  | - 1st byte (8 bits)  : Data type (e.g., byte, uint32)    |
// |                  | - Next 4 bytes (32 bits) : Data offset in memory         |
// |                  | - Last 3 bytes (24 bits) : Data length or size           |
// +------------------+----------------------------------------------------------+
// | Packed Data      | Compile individual allocations into packedData entries.  |
// | Creation         | For example: PD1, PD2, PD3, ...                          |
// +------------------+----------------------------------------------------------+
// | Slice Allocation | Allocate a continuous block of memory for the packedData |
// |                  | slice. This will store the array of packedData entries.  |
// +------------------+----------------------------------------------------------+
// | Storing in       | Insert the packedData slice into linear memory.          |
// | Linear Memory    | Now, the guest function can read, unpack, and extract    |
// |                  | information for each item in the slice.                  |
// +-----------------------------------------------------------------------------+
func (hf *HostFunction) writeResultsToMemory(ctx context.Context, m ModuleProxy, results *Results, stackParams []uint64) ([]uint64, map[uint32]uint32, error) {

	// If the host function does not return anything, just skip the whole process
	if results == nil {
		return nil, nil, nil
	}

	fmt.Println("FUNC: ", hf.Name)

	if len(*results) != len(hf.Returns) {
		return nil, nil, fmt.Errorf("return value missmatch %d != %d", len(*results), len(hf.Returns))
	}

	// First, allocate memory for each byte slice and store the offsets in a slice
	packedDatas := make([]uint64, len(*results))

	// +1 len because for the offset which holds all offsets
	returnOffsets := make(map[uint32]uint32, len(*results)+1)

	for i, returnValue := range *results {

		// get offset size and result value type (ValueType) by result's returnValue
		valueType, offsetSize, err := mdk.GetOffsetSizeAndDataTypeByConversion(returnValue)
		if err != nil {
			err = errors.Join(errors.New("can't convert result"), err)
			return nil, nil, err
		}

		if mdk.ValueType(valueType) != mdk.ValueType(hf.Returns[i]) {
			return nil, nil, fmt.Errorf("return value does not match actual value %d != %d", valueType, hf.Returns[i])
		}

		// allocate memory for each value
		offset, err := m.Malloc(offsetSize)
		if err != nil {
			err = errors.Join(errors.New("can't allocate memory for return value"), err)
			return nil, nil, err
		}

		returnOffsets[offset] = offsetSize

		// Add offset and offset size in the hsot function's allocationMap
		// for later cleanup.
		hf.allocationMap.store(offset, offsetSize)

		err = m.Write(offset, returnValue)
		if err != nil {
			err = errors.Join(errors.New("can't write return value"), err)
			return nil, nil, err
		}

		// Pack the offset and size into a single uint64
		packedDatas[i], err = mdk.PackUI64(uint8(valueType), offset, offsetSize)
		if err != nil {
			return nil, nil, err
		}
	}

	// Then, allocate memory for the array of packed offsets and sizes
	offsetSize := uint32(len(packedDatas) * 8)
	offset, err := m.Malloc(offsetSize)
	if err != nil {
		err = errors.Join(errors.New("can't allocate memory for offset of packed return values"), err)
		return nil, nil, err
	}

	returnOffsets[offset] = offsetSize
	// Add offset and offset size in the hsot function's allocationMap
	// for later cleanup.
	hf.allocationMap.store(offset, offsetSize)

	err = m.Write(offset, uint64ArrayToBytes(packedDatas))
	if err != nil {
		err = errors.Join(errors.New("can't write offset of packed return values"), err)
		return nil, nil, err
	}

	// Final packed data, which contains offset and size of packedDatas slice
	packedData, err := mdk.PackUI64(valueTypePack, offset, offsetSize)
	if err != nil {
		return nil, nil, err
	}

	// Store final packedData into linear memory
	stackParams[0] = packedData

	// Append final packedData to existing packedDatas slice for later cleanup
	packedDatas = append(packedDatas, packedData)

	return packedDatas, returnOffsets, nil
}

// cleanup function is responsible for releasing memory allocated during the execution
// of the host function. It iterates through the parameters and return offsets, freeing
// the associated memory allocations. The totalSize of memory released is calculated,
// and details are logged.
// cleanup will be ran at the end of the execution of host func callback.
func (hf *HostFunction) cleanup(m ModuleProxy, params Params, returnOffsets map[uint32]uint32) error {

	totalSize := hf.allocationMap.totalSize()

	for _, param := range params {
		if _, ok := hf.allocationMap.load(param.Offset); !ok {
			continue
		}

		err := m.Free(param.Offset)
		if err != nil {
			err = errors.Join(errors.New("can't free offset of param"), err)
			return err
		}

		hf.allocationMap.delete(param.Offset)
	}

	for offset := range returnOffsets {
		err := m.Free(offset)
		if err != nil {
			err = errors.Join(errors.New("can't free offset of return value"), err)
			return err
		}

		hf.allocationMap.delete(offset)
	}

	hf.moduleConfig.log.Debug(
		"cleanup: host func params and returns",
		"total_bytes", totalSize,
		"available_bytes", hf.allocationMap.totalSize(),
		"func", hf.Name,
		"module", hf.moduleConfig.Name)

	return nil
}
