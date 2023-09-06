package wasify

import (
	"context"
	"errors"

	"github.com/tetratelabs/wazero/api"
)

// wazeroHostFunctionCallback returns a callback function that acts as a bridge
// between the host function and the wazero runtime. It handles the execution of
// the host function, parameter conversion, user-defined callback execution, return
// value handling, and memory cleanup.
//
// +--------------------------------------+
// | wazeroHostFunctionCallback           |
// |                                      |
// |  +---------------------------+       |
// |  | Initialize wazeroModule   |       |
// |  | and ModuleProxy           |       |
// |  +---------------------------+       |
// |                                      |
// |  +----------------------------+      |
// |  | Convert Stack Params to    |      |
// |  | Structured Params for      |      |
// |  | Host Function              |      |
// |  +----------------------------+      |
// |                 |                    |
// |                 v                    |
// |                                      |
// |  +----------------------------+      |
// |  | 🚀 Execute User-defined    |      |
// |  | Host Function Callback     |      |
// |  +----------------------------+      |
// |                 |                    |
// |                 v                    |
// |                                      |
// |  +-----------------------------+     |
// |  | Convert Return Values to    |     |
// |  | Packed Data using           |     |
// |  | writeReturnValues           |     |
// |  | and write final packedData  |     |
// |  | into linear memory          |     |
// |  +-----------------------------+     |
// |                 |                    |
// |                 v                    |
// |                                      |
// |  +----------------------------+      |
// |  | Perform Memory Cleanup     |      |
// |  | using Host Function's      |      |
// |  | cleanup Method             |      |
// |  +----------------------------+      |
// |                                      |
// +--------------------------------------+
func wazeroHostFunctionCallback(wazeroModule *wazeroModule, moduleConfig *ModuleConfig, hf *HostFunction) func(context.Context, api.Module, []uint64) {

	return func(ctx context.Context, mod api.Module, stack []uint64) {

		wazeroModule.mod = mod
		moduleProxy := &wazeroModuleProxy{
			wazeroModule,
		}

		params, err := hf.convertParamsToStruct(ctx, moduleProxy, stack)
		if err != nil {
			moduleConfig.log.Error(err.Error(), "func", hf.Name, "module", wazeroModule.Name)
			panic(err)
		}

		// user defined host function callback
		returnValues := hf.Callback(ctx, moduleProxy, params)

		// convert Go types to uint64 values and write them to the stack
		_, returnOffsets, err := hf.writeResultsToMemory(ctx, moduleProxy, returnValues, stack)
		if err != nil {
			err = errors.Join(errors.New("function executed, bug can't write to memory"), err)
			moduleConfig.log.Error(err.Error(), "func", hf.Name, "module", wazeroModule.Name)
			panic(err)
		}

		err = hf.cleanup(moduleProxy, params, returnOffsets)
		if err != nil {
			moduleConfig.log.Error(err.Error(), "func", hf.Name, "module", wazeroModule.Name)
			panic(err)
		}
	}
}
