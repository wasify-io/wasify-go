package wasify_test

import (
	"context"
	_ "embed"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wasify-io/wasify-go"
)

//go:embed testdata/wasm/host_all_available_types/main.wasm
var wasm_hostAllAvailableTypes []byte

func TestHostFunctions(t *testing.T) {

	testRuntimeConfig := wasify.RuntimeConfig{
		Runtime: wasify.RuntimeWazero,
	}

	testModuleConfig := wasify.ModuleConfig{
		Namespace: "host_all_available_types",
		Wasm: wasify.Wasm{
			Binary: wasm_hostAllAvailableTypes,
		},
		HostFunctions: []wasify.HostFunction{
			{
				Name: "hostTest",
				Callback: func(ctx context.Context, m wasify.ModuleProxy, params wasify.Params) *wasify.Results {

					_bytes, _ := params[0].Value.([]byte)
					assert.Equal(t, []byte("Guest: Wello Wasify!"), _bytes)

					_byte, _ := params[1].Value.(byte)
					assert.Equal(t, byte(1), _byte)

					_uint32, _ := params[2].Value.(uint32)
					assert.Equal(t, uint32(11), _uint32)

					_uint64, _ := params[3].Value.(uint64)
					assert.Equal(t, uint64(2023), _uint64)

					_float32, _ := params[4].Value.(float32)
					assert.Equal(t, float32(11.1), _float32)

					_float64, _ := params[5].Value.(float64)
					assert.Equal(t, float64(11.2023), _float64)

					_string, _ := params[6].Value.(string)
					assert.Equal(t, "Guest: Wasify.", _string)

					return m.Return(
						[]byte("Host: Wello Wasify!"),
						byte(1),
						uint32(11),
						uint64(2023),
						float32(11.1),
						float64(11.2023),
						"Host: Wasify.",
					)

				},
				Params: []wasify.ValueType{
					wasify.ValueTypeBytes,
					wasify.ValueTypeByte,
					wasify.ValueTypeI32,
					wasify.ValueTypeI64,
					wasify.ValueTypeF32,
					wasify.ValueTypeF64,
					wasify.ValueTypeString,
				},
				Results: []wasify.ValueType{
					wasify.ValueTypeBytes,
					wasify.ValueTypeByte,
					wasify.ValueTypeI32,
					wasify.ValueTypeI64,
					wasify.ValueTypeF32,
					wasify.ValueTypeF64,
					wasify.ValueTypeString,
				},
			},
		},
	}

	t.Run("successful instantiation", func(t *testing.T) {

		ctx := context.Background()

		runtime, err := wasify.NewRuntime(ctx, &testRuntimeConfig)
		assert.NoError(t, err)

		defer func() {
			err = runtime.Close(ctx)
			assert.NoError(t, err)
		}()

		module, err := runtime.NewModule(ctx, &testModuleConfig)
		defer func() {
			err = module.Close(ctx)
			assert.NoError(t, err)
		}()

		res, err := module.GuestFunction(ctx, "guestTest").Invoke()
		assert.NoError(t, err)

		t.Log("TestHostFunctions RES:", res)
	})
}
