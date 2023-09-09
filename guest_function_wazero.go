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

// Invoke calls the guest function with the provided parameters.
func (gf *wazeroGuestFunction) Invoke(params ...any) (uint64, error) {

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
			return 0, err
		}

		// allocate memory for each value
		offsetI32, err := gf.memory.Malloc(offsetSize)
		if err != nil {
			err = errors.Join(fmt.Errorf("An error occurred while attempting to alloc memory for guest func param in: %s", gf.name), err)
			gf.moduleConfig.log.Error(err.Error())
			return 0, err
		}

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
