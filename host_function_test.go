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
				Callback: func(ctx context.Context, m *wasify.ModuleProxy, params []wasify.PackedData) wasify.MultiPackedData {

					_bytes, _ := m.Memory.ReadBytesPack(params[0])
					assert.Equal(t, []byte("Guest: Wello Wasify!"), _bytes)

					_byte, _ := m.Memory.ReadBytePack(params[1])
					assert.Equal(t, byte(1), _byte)

					_uint32, _ := m.Memory.ReadUint32Pack(params[2])
					assert.Equal(t, uint32(11), _uint32)

					_uint64, _ := m.Memory.ReadUint64Pack(params[3])
					assert.Equal(t, uint64(2023), _uint64)

					_float32, _ := m.Memory.ReadFloat32Pack(params[4])
					assert.Equal(t, float32(11.1), _float32)

					_float64, _ := m.Memory.ReadFloat64Pack(params[5])
					assert.Equal(t, float64(11.2023), _float64)

					_string, _ := m.Memory.ReadStringPack(params[6])
					assert.Equal(t, "Guest: Wasify.", _string)

					return m.Memory.WriteMultiPack(
						m.Memory.WriteBytesPack([]byte("Some")),
						m.Memory.WriteBytePack(1),
						m.Memory.WriteUint32Pack(11),
						m.Memory.WriteUint64Pack(2023),
						m.Memory.WriteFloat32Pack(11.1),
						m.Memory.WriteFloat64Pack(11.2023),
						m.Memory.WriteStringPack("Host: Wasify."),
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
