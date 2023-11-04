package wasify

import (
	"context"

	"github.com/tetratelabs/wazero/api"
)

// wazeroHostFunctionCallback returns a callback function that acts as a bridge between
// the host function and the wazero runtime. This bridge ensures the seamless integration of
// the host function within the wazero environment by managing various tasks, including:
//
//   - Initialization of wazeroModule and ModuleProxy to set up the execution environment.
//   - Converting stack parameters into structured parameters that the host function can understand.
//   - Executing the user-defined host function callback with the correctly formatted parameters.
//   - Processing the results of the host function, converting them back into packed data format,
//     and writing the final packed data into linear memory.
//
// Diagram of the wazeroHostFunctionCallback Process:
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
// |  +----------------------------+      |
// |  | 🚀 Execute User-defined    |      |
// |  | Host Function Callback     |      |
// |  +----------------------------+      |
// |                 |                    |
// |                 v                    |
// |  +-----------------------------+     |
// |  | Convert Return Values to    |     |
// |  | Packed Data using           |     |
// |  | writeResultsToMemory        |     |
// |  | and write final packedData  |     |
// |  | into linear memory          |     |
// |  +-----------------------------+     |
// |                                      |
// +--------------------------------------+
//
// Return value: A callback function that takes a context, api.Module, and a stack of parameters,
// and handles the integration of the host function within the wazero runtime.
func wazeroHostFunctionCallback(wazeroModule *wazeroModule, moduleConfig *ModuleConfig, hf *HostFunction) func(context.Context, api.Module, []uint64) {

	return func(ctx context.Context, mod api.Module, stack []uint64) {

		wazeroModule.mod = mod
		moduleProxy := &ModuleProxy{
			Memory: wazeroModule.Memory(),
		}

		params, err := hf.preHostFunctionCallback(ctx, moduleProxy, stack)
		if err != nil {
			moduleConfig.log.Error(err.Error(), "namespace", wazeroModule.Namespace, "func", hf.Name)
		}

		results := hf.Callback(ctx, moduleProxy, params)

		hf.postHostFunctionCallback(ctx, moduleProxy, results, stack)

	}
}
