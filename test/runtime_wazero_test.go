package test

import (
	"context"
	_ "embed"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wasify-io/wasify-go"
	"github.com/wasify-io/wasify-go/internal/utils"
)

//go:embed _data/empty_host_func.wasm
var emptyHostFunc []byte

var runtimeConfig = wasify.RuntimeConfig{
	Runtime:     wasify.RuntimeWazero,
	LogSeverity: wasify.LogWarning,
}

var moduleConfig = wasify.ModuleConfig{
	Namespace:   "myEnv",
	LogSeverity: wasify.LogError,
	FSConfig: wasify.FSConfig{
		Enabled:  true,
		HostDir:  "test/_data/",
		GuestDir: "/",
	},
	Wasm: wasify.Wasm{
		Binary: emptyHostFunc,
	},
	HostFunctions: []wasify.HostFunction{
		{
			Name: "hostFunc",
			Callback: func(ctx context.Context, m wasify.ModuleProxy, params wasify.Params) *wasify.Results {
				return nil
			},
			Params:  []wasify.ValueType{},
			Returns: []wasify.ValueType{},
		},
	},
}

func TestMain(t *testing.T) {

	hash, err := utils.CalculateHash(emptyHostFunc)
	assert.NoError(t, err, "Expected no error while calculating hash")

	moduleConfig.Wasm.Hash = hash

}

func TestNewModuleInstantaion(t *testing.T) {

	ctx := context.Background()

	t.Run("successful instantiation", func(t *testing.T) {
		runtime, err := wasify.NewRuntime(ctx, &runtimeConfig)
		assert.NoError(t, err, "Expected no error while creating runtime")
		assert.NotNil(t, runtime, "Expected a non-nil runtime")

		defer func() {
			err = runtime.Close(ctx)
			assert.NoError(t, err, "Expected no error while closing runtime")
		}()

		module, err := runtime.NewModule(ctx, &moduleConfig)
		assert.NoError(t, err, "Expected no error while creating module")
		assert.NotNil(t, module, "Expected a non-nil module")

		defer func() {
			err = module.Close(ctx)
			assert.Nil(t, err, "Expected no error while closing module")
		}()
	})

	t.Run("failure due to invalid runtime", func(t *testing.T) {
		invalidConfig := runtimeConfig
		invalidConfig.Runtime = 255

		runtime, err := wasify.NewRuntime(ctx, &invalidConfig)
		assert.Error(t, err, "Expected an error due to invalid config")
		assert.Nil(t, runtime, "Expected a nil runtime due to invalid config")
	})

	t.Run("failure due to invalid hash", func(t *testing.T) {
		runtime, err := wasify.NewRuntime(ctx, &runtimeConfig)
		assert.NoError(t, err, "Expected no error while creating runtime")

		defer func() {
			err = runtime.Close(ctx)
			assert.NoError(t, err, "Expected no error while closing runtime")
		}()

		invalidModuleConfig := moduleConfig
		invalidModuleConfig.Wasm.Hash = "invalid_hash"

		module, err := runtime.NewModule(ctx, &invalidModuleConfig)
		assert.Error(t, err)

		defer func() {
			assert.Nil(t, module)
		}()
	})

	t.Run("failure due to invalid wasm", func(t *testing.T) {
		runtime, err := wasify.NewRuntime(ctx, &runtimeConfig)
		assert.NoError(t, err, "Expected no error while creating runtime")

		defer func() {
			err = runtime.Close(ctx)
			assert.NoError(t, err, "Expected no error while closing runtime")
		}()

		invalidModuleConfig := moduleConfig
		invalidModuleConfig.Wasm.Binary = []byte("invalid wasm data")
		invalidModuleConfig.Wasm.Hash = ""

		module, err := runtime.NewModule(ctx, &invalidModuleConfig)
		assert.Error(t, err)
		assert.Nil(t, module)
	})

}
