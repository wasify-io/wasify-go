# wasify

`wasify-go` is a Go library designed to streamline the interaction with WebAssembly (Wasm) modules by providing a developer-friendly API. It abstracts the [Wazero](https://github.com/tetratelabs/wazero) runtime.

wasify simplifies communication with the WebAssembly System Interface (WASI), eliminating the need for developers to delve into intricate details or communicate using numbers, which is the standard method of interaction with modules. This library significantly eases the process of running and working with wasm modules, which has traditionally been a less straightforward task.


---
> [!WARNING]
> - **The project is still being worked on, and there might be significant changes coming;**
> - Requires Go v1.21

---
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
        Name: "myEnv",
        Wasm: wasify.Wasm{
            Binary: moduleData,
        },
        HostFunctions: []wasify.HostFunction{
            {
                Name: "hostFunc",
                Callback: func(ctx context.Context, m wasify.ModuleProxy, params wasify.Params) *wasify.Results {

                    fmt.Println("Host func param 0: ", params[0].Value)
                    fmt.Println("Host func param 1: ", params[1].Value)

                    return m.Return(
                        []byte("Hello"),
                        uint32(1234),
                    )

                },
                Params:  []wasify.ValueType{wasify.ValueTypeString, wasify.ValueTypeI32},
                Results: []wasify.ValueType{wasify.ValueTypeBytes, wasify.ValueTypeI32},
            },
        },
    })

    defer module.Close(ctx)

    // Call guest function
    module.GuestFunction(ctx, "greet").Invoke()

}
```

### Wasm Module Example

```go
package main

import "github.com/wasify-io/wasify-go/mdk"

func main() {}

//go:wasmimport myEnv hostFunc
func hostFunc(mdk.ArgData, mdk.ArgData) mdk.ResultOffset

//export greet
func greet() {
    resultOffset := hostLog(mdk.Arg("Hello"), mdk.Arg(uint32(2023)))

    results := mdk.ReadResults(resultOffset)

    for i, result := range results {
        mdk.Log("Guest func result %d: %s\r\n", i, string(result.Data))
    }
}
```

Build module using [TinyGo](https://tinygo.org/)
```
tinygo build -o ./module/example.wasm -target wasi ./module/example.go
```

Run main.go `go run .`

## Contributing

Pull requests are welcome. For major changes, please open an issue first to discuss what you would like to change.

