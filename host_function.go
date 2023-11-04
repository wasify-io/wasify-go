package wasify

import (
	"context"
	"fmt"

	"github.com/wasify-io/wasify-go/internal/types"
)

// ValueType represents the type of value used in function parameters and returns.
type ValueType types.ValueType

// supported value types in params and returns
const (
	ValueTypeBytes  ValueType = ValueType(types.ValueTypeBytes)
	ValueTypeByte   ValueType = ValueType(types.ValueTypeByte)
	ValueTypeI32    ValueType = ValueType(types.ValueTypeI32)
	ValueTypeI64    ValueType = ValueType(types.ValueTypeI64)
	ValueTypeF32    ValueType = ValueType(types.ValueTypeF32)
	ValueTypeF64    ValueType = ValueType(types.ValueTypeF64)
	ValueTypeString ValueType = ValueType(types.ValueTypeString)
)

// Param defines the attributes of a function parameter.
type MultiPackedData uint64
type PackedData uint64

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

	// Results specifies the types of values that the host function Results.
	//
	// The length of 'Results' should match the expected number of Results
	// from the host function as used in the guest.
	Results []ValueType

	// Allocation map to track parameter and return value allocations for host func.

	// Configuration of the associated module.
	moduleConfig *ModuleConfig
}

// HostFunctionCallback is the function signature for the callback executed by a host function.
//
// HostFunctionCallback encapsulates the runtime's internal implementation details.
// It serves as an intermediary invoked between the processing of function parameters and the final return of the function.
type HostFunctionCallback func(ctx context.Context, moduleProxy *ModuleProxy, multiPackedData []PackedData) MultiPackedData

// preHostFunctionCallback
// prepares parameters for the host function by converting
// packed stack parameters into a slice of PackedData. It validates parameter counts
// and leverages ModuleProxy for reading the data.
func (hf *HostFunction) preHostFunctionCallback(ctx context.Context, m *ModuleProxy, stackParams []uint64) ([]PackedData, error) {

	// If user did not define params, skip the whole process, we still might get stackParams[0] = 0
	if len(hf.Params) == 0 {
		return nil, nil
	}

	if len(hf.Params) != len(stackParams) {
		return nil, fmt.Errorf("%s: params mismatch expected: %d received: %d ", hf.Name, len(hf.Params), len(stackParams))
	}

	pds := make([]PackedData, len(hf.Params))

	for i := range hf.Params {
		pds[i] = PackedData(stackParams[i])
	}

	return pds, nil

}

// postHostFunctionCallback
// stores the resulting MultiPackedData into linear memory after the host function execution.
func (hf *HostFunction) postHostFunctionCallback(ctx context.Context, m *ModuleProxy, mpd MultiPackedData, stackParams []uint64) {
	// Store final MultiPackedData into linear memory
	stackParams[0] = uint64(mpd)
}
