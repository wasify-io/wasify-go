package wasify

import (
	"context"
	"errors"
	"log/slog"

	"github.com/wasify-io/wasify-go/internal/utils"
)

type LogSeverity utils.LogSeverity

// The log level is initially set to "Info" for runtimes and "zero" (0) for modules.
// However, modules will adopt the log level from their parent runtime.
// If you want only "Error" level for a runtime but need to debug specific module(s),
// you can set those modules to "Debug". This will replace the inherited log level,
// allowing the module to display debug information.
const (
	LogDebug   LogSeverity = LogSeverity(utils.LogDebug)
	LogInfo    LogSeverity = LogSeverity(utils.LogInfo)
	LogWarning LogSeverity = LogSeverity(utils.LogWarning)
	LogError   LogSeverity = LogSeverity(utils.LogError)
)

type Runtime interface {
	NewModule(context.Context, *ModuleConfig) (Module, error)
	Close(ctx context.Context) error
}

// RuntimeType defines a type of WebAssembly (wasm) runtime.
//
// Currently, the only supported wasm runtime is Wazero.
// However, in the future, more runtimes could be added.
// This means that you'll be able to run modules
// on various wasm runtimes.
type RuntimeType uint8

const (
	RuntimeWazero RuntimeType = iota
)

func (rt RuntimeType) String() (runtimeName string) {

	switch rt {
	case RuntimeWazero:
		runtimeName = "Wazero"
	}

	return
}

// The RuntimeConfig struct holds configuration settings for a runtime.
type RuntimeConfig struct {
	// Specifies the type of runtime being used.
	Runtime RuntimeType
	// Determines the severity level of logging.
	LogSeverity LogSeverity
	// Pointer to a logger for recording runtime information.
	log *slog.Logger
}

// NewRuntime creates and initializes a new runtime based on the provided configuration.
// It returns the initialized runtime and any error that might occur during the process.
func NewRuntime(ctx context.Context, c *RuntimeConfig) (runtime Runtime, err error) {

	c.log = utils.NewLogger(utils.LogSeverity(c.LogSeverity))

	c.log.Info("runtime has been initialized successfully", "runtime", c.Runtime)

	// Retrieve the appropriate runtime implementation based on the configured type.
	runtime = c.getRuntime(ctx)
	if runtime == nil {
		err = errors.New("unsupported runtime")
		c.log.Error(err.Error(), "runtime", c.Runtime)
		return
	}

	return
}

// getRuntime returns an instance of the appropriate runtime implementation
// based on the configured runtime type in the RuntimeConfig.
func (c *RuntimeConfig) getRuntime(ctx context.Context) Runtime {
	switch c.Runtime {
	case RuntimeWazero:
		return getWazeroRuntime(ctx, c)
	default:
		return nil
	}
}
