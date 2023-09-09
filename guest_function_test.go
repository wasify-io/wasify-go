package wasify_test

import (
	"context"
	_ "embed"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wasify-io/wasify-go"
)

//go:embed testdata/wasm/guest_all_available_types/main.wasm
var wasm_guestAllAvailableTypes []byte

func TestGuestFunctions(t *testing.T) {

	testRuntimeConfig := wasify.RuntimeConfig{
		Runtime:     wasify.RuntimeWazero,
		LogSeverity: wasify.LogError,
	}

	testModuleConfig := wasify.ModuleConfig{
		Namespace: "guest_all_available_types",
		Wasm: wasify.Wasm{
			Binary: wasm_guestAllAvailableTypes,
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

		res, err := module.GuestFunction(ctx, "guestTest").Invoke(
			[]byte("bytes!"),
			byte(1),
			uint32(32),
			uint64(64),
			float32(32.0),
			float64(64.01),
			"Wasify",
			"any type",
		)
		assert.NoError(t, err)

		t.Log("TestGuestFunctions RES:", res)
	})
}
