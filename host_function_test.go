package wasify_test

import (
	"context"
	_ "embed"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wasify-io/wasify-go"
)

//go:embed testdata/wasm/new.wasm
var new []byte

func TestHostFunctions(t *testing.T) {

	ctx := context.Background()

	t.Run("successful instantiation", func(t *testing.T) {
		runtime, err := wasify.NewRuntime(ctx, &testRuntimeConfig)
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
}
