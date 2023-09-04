package wasify

import (
	"context"
	"fmt"
)

// hostFunctions is a list of pre-defined host functions
type hostFunctions struct {
	hostFunctions *[]*HostFunction
}

func newHostFunctions() *hostFunctions {

	return &hostFunctions{}
}

// log data from the guest module to the host machine to avoid stdin/stdout calls,
// to ensure sandboxing.
func (hf *hostFunctions) log() *HostFunction {

	log := &HostFunction{
		Name: "log",
		Callback: func(ctx context.Context, m ModuleProxy, params Params) *Results {

			fmt.Println("Host func param 0: ", string(params[0].Value))
			fmt.Println("Host func param 1: ", string(params[1].Value))

			a := m.Return(
				[]byte("Hello"),
				[]byte("There!"),
			)

			return a

		},
		Params:  []ValueType{ValueTypeByte},
		Returns: []ValueType{},
	}

	return log
}
