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
	Read(packedData uint64) (data any, offset uint32, size uint32, err error)
	Write(offset uint32, data any) error
	Free(offset uint32) error
	Malloc(size uint32) (offset uint32, err error)
	Size() uint32
	Return(...Result) MultiPackedData
}

type GuestFunction interface {
	Invoke(args ...any) (*GuestFunctionResult, error)
	call(args ...uint64) (uint64, error)
}

type Memory interface {
	Read(packedData uint64) (data any, offset uint32, size uint32, err error)
	ReadBytes(offset uint32, size uint32) ([]byte, error)
	ReadByte(offset uint32) (byte, error)
	ReadUint32(offset uint32) (uint32, error)
	ReadUint64(offset uint32) (uint64, error)
	ReadFloat32(offset uint32) (float32, error)
	ReadFloat64(offset uint32) (float64, error)
	ReadString(offset uint32, size uint32) (string, error)
	Write(offset uint32, data any) error
	WriteBytes(offset uint32, v []byte) error
	WriteByte(offset uint32, v byte) error
	WriteUint32(offset uint32, v uint32) error
	WriteUint64(offset uint32, v uint64) error
	WriteFloat32(offset uint32, v float32) error
	WriteFloat64(offset uint32, v float64) error
	WriteString(offset uint32, v string) error
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

// getGuestDir gets the default path for guest module.
func (fs *FSConfig) getGuestDir() string {

	if fs.GuestDir == "" {
		return "/"
	}

	return fs.GuestDir
}
