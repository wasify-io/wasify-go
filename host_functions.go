package wasify

import (
	"context"
)

const WASIFY_NAMESPACE = "wasify"

// hostFunctions is a list of pre-defined host functions
type hostFunctions struct {
	moduleConfig *ModuleConfig
}

func newHostFunctions(moduleConfig *ModuleConfig) *hostFunctions {
	return &hostFunctions{moduleConfig}
}

// newLog logs data from the guest module to the host machine,
// to avoid stdin/stdout calls and ensure sandboxing.
func (hf *hostFunctions) newLog() *HostFunction {

	log := &HostFunction{
		Name: "log",
		Callback: func(ctx context.Context, m *ModuleProxy, params []PackedData) MultiPackedData {

			msg, err := m.Memory.ReadStringPack(params[0])
			if err != nil {
				panic(err)
			}

			lvl, err := m.Memory.ReadBytePack(params[0])
			if err != nil {
				panic(err)
			}

			severity := LogSeverity(lvl)

			switch severity {
			case LogDebug:
				hf.moduleConfig.log.Debug(msg)
			case LogInfo:
				hf.moduleConfig.log.Info(msg)
			case LogWarning:
				hf.moduleConfig.log.Warn(msg)
			case LogError:
				hf.moduleConfig.log.Error(msg)
			}

			return 0

		},
		Params:  []ValueType{ValueTypeBytes, ValueTypeBytes},
		Results: nil,

		// required fields
		moduleConfig: hf.moduleConfig,
	}

	return log
}
