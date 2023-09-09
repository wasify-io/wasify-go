package wasify

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewRuntime(t *testing.T) {
	ctx := context.Background()

	runtimeConfig := &RuntimeConfig{
		Runtime: RuntimeWazero,
	}

	runtime, err := NewRuntime(ctx, runtimeConfig)
	assert.NoError(t, err, "Expected no error while creating a new runtime")
	assert.NotNil(t, runtime, "Expected a non-nil runtime")
}

func TestNewRuntimeUnsupported(t *testing.T) {
	ctx := context.Background()

	runtimeConfig := &RuntimeConfig{
		Runtime: 255, // Assuming this is an unsupported value
	}

	runtime, err := NewRuntime(ctx, runtimeConfig)
	assert.Error(t, err, "Expected error for unsupported runtime")
	assert.Nil(t, runtime, "Expected a nil runtime for unsupported value")
}

func TestRuntimeTypeString(t *testing.T) {
	assert.Equal(t, "Wazero", RuntimeWazero.String(), "Expected Wazero string representation")
}
