package wasify

import (
	"context"
	"fmt"
)

// hostFunctions is a list of pre-defined host functions
type hostFunctions struct {
	moduleConfig *ModuleConfig
}

func newHostFunctions(moduleConfig *ModuleConfig) *hostFunctions {
	return &hostFunctions{moduleConfig}
}

// log data from the guest module to the host machine to avoid stdin/stdout calls,
// to ensure sandboxing.
func (hf *hostFunctions) newLog() *HostFunction {

	log := &HostFunction{
		Name: "log",
		Callback: func(ctx context.Context, m ModuleProxy, params Params) *Results {

			fmt.Println(string(params[0].Value))

			return nil

		},
		Params:  []ValueType{ValueTypeByte},
		Returns: []ValueType{},
	}

	log.moduleConfig = hf.moduleConfig
	log.allocationMap = newAllocationMap[uint32, uint32]()

	return log
}
