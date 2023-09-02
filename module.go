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
	Read(packedData uint64) (offset uint32, size uint32, data []byte, err error)
	Write(offset uint32, data []byte) error
	Free(offset uint32) error
	Malloc(size uint32) (offset uint32, err error)
	Size() uint32
	Return(...Result) Results
}

type GuestFunction interface {
	Invoke(args ...uint64) ([]uint64, error)
}

type Memory interface {
	Read(packedData uint64) (offset uint32, size uint32, data []byte, err error)
	Write(offset uint32, data []byte) error
	Free(offset uint32) error
	Size() uint32
	Malloc(size uint32) (uint32, error)
}

type ModuleConfig struct {
	Name          string
	Wasm          Wasm
	Attributes    map[string]string
	HostFunctions []HostFunction
	LogSeverity   LogSeverity
	ctx           context.Context
	log           *slog.Logger
}
type Wasm struct {
	Hash   string
	Binary []byte
}
