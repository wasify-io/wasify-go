# wasify

`wasify-go` is a Go library designed to streamline the interaction with WebAssembly (Wasm) modules by providing a developer-friendly API. It abstracts the [Wazero](https://github.com/tetratelabs/wazero) runtime.

wasify simplifies communication with the WebAssembly System Interface (WASI), eliminating the need for developers to delve into intricate details or communicate using numbers, which is the standard method of interaction with modules. This library significantly eases the process of running and working with wasm modules, which has traditionally been a less straightforward task.

## Installation

```bash
go get github.com/wasify-io/wasify-go
```

## Example

### main.go

```go
package main

import (
    "context"
    _ "embed"
    "fmt"

    "github.com/wasify-io/wasify-go"
)

//go:embed module/example.wasm
var moduleData []byte

func main() {

    ctx := context.Background()

    runtime, _ := wasify.NewRuntime(ctx, &wasify.RuntimeConfig{
        Runtime:     wasify.RuntimeWazero,
        LogSeverity: wasify.LogInfo,
    })
    defer runtime.Close(ctx)

    module, _ := runtime.NewModule(ctx, &wasify.ModuleConfig{
        Name: "host_all_available_types",
        Wasm: wasify.Wasm{
            Binary: moduleData,
        },
        HostFunctions: []wasify.HostFunction{
            {
                Name: "hostTest",
                Callback: func(ctx context.Context, m *wasify.ModuleProxy, params []wasify.PackedData) wasify.MultiPackedData {

                    bytes, _ := m.Memory.ReadBytesPack(params[0])
                    fmt.Println("Param 1: ", bytes)
                    // ...

                    return m.Memory.WriteMultiPack(
						m.Memory.WriteBytesPack([]byte("Some")),
						m.Memory.WriteBytePack(1),
						m.Memory.WriteUint32Pack(11),
						m.Memory.WriteUint64Pack(2023),
						m.Memory.WriteFloat32Pack(11.1),
						m.Memory.WriteFloat64Pack(11.2023),
						m.Memory.WriteStringPack("Host: Wasify."),
					)

                },
                Params: []wasify.ValueType{
					wasify.ValueTypeBytes,
					wasify.ValueTypeByte,
					wasify.ValueTypeI32,
					wasify.ValueTypeI64,
					wasify.ValueTypeF32,
					wasify.ValueTypeF64,
					wasify.ValueTypeString,
				},
				Results: []wasify.ValueType{
					wasify.ValueTypeBytes,
					wasify.ValueTypeByte,
					wasify.ValueTypeI32,
					wasify.ValueTypeI64,
					wasify.ValueTypeF32,
					wasify.ValueTypeF64,
					wasify.ValueTypeString,
				},
            },
        },
    })

    defer module.Close(ctx)

    // Call guest function
    module.GuestFunction(ctx, "greet").Invoke("example arg", 2023)

}
```

### Wasm Module Example

```go
package main

import (
	"github.com/wasify-io/wasify-go/mdk"
)

func main() {}

//go:wasmimport host_all_available_types hostTest
func hostTest(
	mdk.PackedData,
	mdk.PackedData,
	mdk.PackedData,
	mdk.PackedData,
	mdk.PackedData,
	mdk.PackedData,
	mdk.PackedData,
) mdk.MultiPackedData

//export guestTest
func _guestTest() {
	hostTest(
		mdk.WriteBytesPack([]byte("Guest: Wello Wasify!")),
		mdk.WriteBytePack(byte(1)),
		mdk.WriteUint32Pack(uint32(11)),
		mdk.WriteUint64Pack(uint64(2023)),
		mdk.WriteFloat32Pack(float32(11.1)),
		mdk.WriteFloat64Pack(float64(11.2023)),
		mdk.WriteStringPack("Guest: Wasify."),
	)
}

```

Build module using [TinyGo](https://tinygo.org/)

```
tinygo build -o ./module/example.wasm -target wasi ./module/example.go
```

Run main.go `go run .`

## Contributing

Pull requests are welcome. For major changes, please open an issue first to discuss what you would like to change.
