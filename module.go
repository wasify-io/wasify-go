package wasify

import (
	"context"
	"log/slog"
)

type Module interface {
	Close(ctx context.Context) error
	GuestFunction(ctx context.Context, functionName string) GuestFunction
	Memory() Memory
}

type ModuleProxy interface {
	GuestFunction(ctx context.Context, functionName string) GuestFunction
	Read(packedData uint64) (offset uint32, size uint32, data any, err error)
	Write(offset uint32, data any) error
	Free(offset uint32) error
	Malloc(size uint32) (offset uint32, err error)
	Size() uint32
	Return(...Result) *Results
}

type GuestFunction interface {
	Invoke(args ...uint64) ([]uint64, error)
}

type Memory interface {
	Read(packedData uint64) (offset uint32, size uint32, data any, err error)
	Write(offset uint32, data any) error
	Free(offset uint32) error
	Size() uint32
	Malloc(size uint32) (uint32, error)
}

type ModuleConfig struct {
	// Module Namespace. Required.
	Namespace string

	// FSConfig configures a directory to be pre-opened for access by the WASI module if Enabled is set to true.
	// If GuestDir is not provided, the default guest directory will be "/".
	// Note: If FSConfig is not provided or Enabled is false, the directory will not be attached to WASI.
	FSConfig FSConfig

	// WASM configuration. Required.
	Wasm Wasm

	// List of host functions to be registered.
	HostFunctions []HostFunction

	// Set the severity level for a particular module's logs.
	// Note: If LogSeverity isn't specified, the severity is inherited from the parent, like the runtime log severity.
	LogSeverity LogSeverity

	// Struct members for internal use.
	ctx context.Context
	log *slog.Logger
}

// Wasm configures a new wasm file.
// Binay is required.
// Hash is optional.
type Wasm struct {
	Binary []byte
	Hash   string
}

// FSConfig configures a directory to be pre-opened for access by the WASI module if Enabled is set to true.
// If GuestDir is not provided, the default guest directory will be "/".
// Note: If FSConfig is not provided or Enabled is false, the directory will not be attached to WASI.
type FSConfig struct {
	// Whether to Enabled the directory for WASI access.
	Enabled bool

	// The directory on the host system.
	// Default: "/"
	HostDir string

	// The directory accessible to the WASI module.
	GuestDir string
}

func (fs *FSConfig) getGuestDir() string {

	if fs.GuestDir == "" {
		return "/"
	}

	return fs.GuestDir
}
