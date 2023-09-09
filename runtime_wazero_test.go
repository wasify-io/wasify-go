package wasify

import (
	"context"
	_ "embed"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tetratelabs/wazero/api"
	"github.com/wasify-io/wasify-go/internal/utils"
)

//go:embed testdata/wasm/empty_host_func/main.wasm
var wasm_emptyHostFunc []byte

func TestNewModuleInstantaion(t *testing.T) {

	testRuntimeConfig := RuntimeConfig{
		Runtime:     RuntimeWazero,
		LogSeverity: LogError,
	}

	testModuleConfig := ModuleConfig{
		Namespace:   "empty_host_func",
		LogSeverity: LogError,
		FSConfig: FSConfig{
			Enabled:  true,
			HostDir:  "test/_data/",
			GuestDir: "/",
		},
		Wasm: Wasm{
			Binary: wasm_emptyHostFunc,
		},
		HostFunctions: []HostFunction{
			{
				Name: "hostFunc",
				Callback: func(ctx context.Context, m ModuleProxy, params Params) *Results {
					return nil
				},
				Params:  []ValueType{},
				Results: []ValueType{},
			},
		},
	}

	hash, err := utils.CalculateHash(wasm_emptyHostFunc)
	assert.NoError(t, err, "Expected no error while calculating hash")
	testModuleConfig.Wasm.Hash = hash

	ctx := context.Background()

	t.Run("successful instantiation", func(t *testing.T) {
		runtime, err := NewRuntime(ctx, &testRuntimeConfig)
		assert.NoError(t, err, "Expected no error while creating runtime")
		assert.NotNil(t, runtime, "Expected a non-nil runtime")

		defer func() {
			err = runtime.Close(ctx)
			assert.NoError(t, err, "Expected no error while closing runtime")
		}()

		module, err := runtime.NewModule(ctx, &testModuleConfig)
		assert.NoError(t, err, "Expected no error while creating module")
		assert.NotNil(t, module, "Expected a non-nil module")

		defer func() {
			err = module.Close(ctx)
			assert.Nil(t, err, "Expected no error while closing module")
		}()
	})

	t.Run("failure due to invalid runtime", func(t *testing.T) {
		invalidConfig := testRuntimeConfig
		invalidConfig.Runtime = 255

		runtime, err := NewRuntime(ctx, &invalidConfig)
		assert.Error(t, err, "Expected an error due to invalid config")
		assert.Nil(t, runtime, "Expected a nil runtime due to invalid config")
	})

	t.Run("failure due to invalid hash", func(t *testing.T) {
		runtime, err := NewRuntime(ctx, &testRuntimeConfig)
		assert.NoError(t, err, "Expected no error while creating runtime")

		defer func() {
			err = runtime.Close(ctx)
			assert.NoError(t, err, "Expected no error while closing runtime")
		}()

		invalidtestModuleConfig := testModuleConfig
		invalidtestModuleConfig.Wasm.Hash = "invalid_hash"

		module, err := runtime.NewModule(ctx, &invalidtestModuleConfig)
		assert.Error(t, err)

		defer func() {
			assert.Nil(t, module)
		}()
	})

	t.Run("failure due to invalid wasm", func(t *testing.T) {
		runtime, err := NewRuntime(ctx, &testRuntimeConfig)
		assert.NoError(t, err, "Expected no error while creating runtime")

		defer func() {
			err = runtime.Close(ctx)
			assert.NoError(t, err, "Expected no error while closing runtime")
		}()

		invalidtestModuleConfig := testModuleConfig
		invalidtestModuleConfig.Wasm.Binary = []byte("invalid wasm data")
		invalidtestModuleConfig.Wasm.Hash = ""

		module, err := runtime.NewModule(ctx, &invalidtestModuleConfig)
		assert.Error(t, err)
		assert.Nil(t, module)
	})

	t.Run("test convertToAPIValueTypes", func(t *testing.T) {
		r := &wazeroRuntime{}

		for vt := ValueTypeBytes; vt <= ValueTypeString; vt++ {
			converted := r.convertToAPIValueTypes([]ValueType{vt})
			assert.Len(t, converted, 1, "Unexpected length of converted types for %v", vt)
			assert.Equal(t, api.ValueTypeI64, converted[0], "Unexpected conversion for %v", vt)
		}
	})

}
