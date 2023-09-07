package wasify

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewRuntimeUnsupported(t *testing.T) {
	ctx := context.Background()
	logger := newLogger(LogDebug)

	runtimeConfig := &RuntimeConfig{
		Runtime:     255, // Assuming this is an unsupported value
		LogSeverity: LogInfo,
		log:         logger,
	}

	runtime, err := NewRuntime(ctx, runtimeConfig)
	assert.Error(t, err, "Expected error for unsupported runtime")
	assert.Nil(t, runtime, "Expected a nil runtime for unsupported value")
}

func TestRuntimeTypeString(t *testing.T) {
	assert.Equal(t, "Wazero", RuntimeWazero.String(), "Expected Wazero string representation")
}
