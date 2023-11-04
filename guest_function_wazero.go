package wasify

import (
	"context"
	"errors"
	"fmt"

	"github.com/tetratelabs/wazero/api"
	"github.com/wasify-io/wasify-go/internal/types"
	"github.com/wasify-io/wasify-go/internal/utils"
)

type wazeroGuestFunction struct {
	ctx          context.Context
	fn           api.Function
	name         string
	memory       Memory
	moduleConfig *ModuleConfig
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
		err = errors.Join(errors.New("error invoking internal call func"), err)
		gf.moduleConfig.log.Error(err.Error())
		return 0, err
	}

	return stack[0], nil
}

// Invoke calls a specified guest function with the provided parameters. It ensures proper memory management,
// data conversion, and compatibility with data types. Each parameter is converted to its packedData format,
// which provides a compact representation of its memory offset, size, and type information. This packedData
// is written into the WebAssembly memory, allowing the guest function to correctly interpret and use the data.
//
// While the method takes care of memory allocation for the parameters and writing them to memory, it does
// not handle freeing the allocated memory. If an error occurs at any step, from data conversion to memory
// allocation, or during the guest function invocation, the error is logged, and the function returns with an error.
//
// Example:
//
// res, err := module.GuestFunction(ctx, "guestTest").Invoke([]byte("bytes!"), uint32(32), float32(32.0), "Wasify")
//
// params ...any: A variadic list of parameters of any type that the user wants to pass to the guest function.
//
// Return value: The result of invoking the guest function in the form of a GuestFunctionResult pointer,
// or an error if any step in the process fails.
func (gf *wazeroGuestFunction) Invoke(params ...any) (*GuestFunctionResult, error) {

	var err error

	log := gf.moduleConfig.log.Info
	if gf.moduleConfig.Namespace == "malloc" || gf.moduleConfig.Namespace == "free" {
		log = gf.moduleConfig.log.Debug
	}

	log("calling guest function", "namespace", gf.moduleConfig.Namespace, "function", gf.name, "params", params)

	stack := make([]uint64, len(params))

	for i, p := range params {
		valueType, offsetSize, err := types.GetOffsetSizeAndDataTypeByConversion(p)
		if err != nil {
			err = errors.Join(fmt.Errorf("Can't convert guest func param %s", gf.name), err)
			return nil, err
		}

		// allocate memory for each value
		offsetI32, err := gf.memory.Malloc(offsetSize)
		if err != nil {
			err = errors.Join(fmt.Errorf("An error occurred while attempting to alloc memory for guest func param in: %s", gf.name), err)
			gf.moduleConfig.log.Error(err.Error())
			return nil, err
		}

		err = gf.memory.WriteAny(offsetI32, p)
		if err != nil {
			err = errors.Join(errors.New("Can't write arg to"), err)
			return nil, err
		}

		stack[i], err = utils.PackUI64(valueType, offsetI32, offsetSize)
		if err != nil {
			err = errors.Join(fmt.Errorf("An error occurred while attempting to pack data for guest func param in:  %s", gf.name), err)
			gf.moduleConfig.log.Error(err.Error())
			return nil, err
		}
	}

	multiPackedData, err := gf.call(stack...)
	if err != nil {
		err = errors.Join(fmt.Errorf("An error occurred while attempting to invoke the guest function: %s", gf.name), err)
		gf.moduleConfig.log.Error(err.Error())
		return nil, err
	}

	res := &GuestFunctionResult{
		multiPackedData: multiPackedData,
		memory:          gf.memory,
	}

	return res, err
}
