package wasify

import (
	"context"

	"github.com/wasify-io/wasify-go/internal/memory"
)

const WASIFY_NAMESPACE = "wasify"

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

			msg := params[0].Value.(string)
			lvl := LogSeverity(params[1].Value.(uint8))

			switch lvl {
			case LogDebug:
				hf.moduleConfig.log.Debug(msg)
			case LogInfo:
				hf.moduleConfig.log.Info(msg)
			case LogWarning:
				hf.moduleConfig.log.Warn(msg)
			case LogError:
				hf.moduleConfig.log.Error(msg)
			}

			return nil

		},
		Params:  []ValueType{ValueTypeBytes, ValueTypeBytes},
		Results: nil,

		// required fields
		moduleConfig:  hf.moduleConfig,
		allocationMap: memory.NewAllocationMap[uint32, uint32](),
	}

	return log
}
