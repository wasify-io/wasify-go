package wasify

import (
	"context"
	_ "embed"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wasify-io/wasify-go/internal/utils"
)

//go:embed test/_data/empty_host_func.wasm
var emptyHostFunc []byte

var runtimeConfig = RuntimeConfig{
	Runtime:     RuntimeWazero,
	LogSeverity: LogWarning,
}

var moduleConfig = ModuleConfig{
	Namespace:   "myEnv",
	LogSeverity: LogError,
	FSConfig: FSConfig{
		Enabled:  true,
		HostDir:  "test/_data/",
		GuestDir: "/",
	},
	Wasm: Wasm{
		Binary: emptyHostFunc,
	},
	HostFunctions: []HostFunction{
		{
			Name: "hostFunc",
			Callback: func(ctx context.Context, m ModuleProxy, params Params) *Results {
				return nil
			},
			Params:  []ValueType{},
			Returns: []ValueType{},
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
		runtime, err := NewRuntime(ctx, &runtimeConfig)
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

		runtime, err := NewRuntime(ctx, &invalidConfig)
		assert.Error(t, err, "Expected an error due to invalid config")
		assert.Nil(t, runtime, "Expected a nil runtime due to invalid config")
	})

	t.Run("failure due to invalid hash", func(t *testing.T) {
		runtime, err := NewRuntime(ctx, &runtimeConfig)
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
		runtime, err := NewRuntime(ctx, &runtimeConfig)
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
