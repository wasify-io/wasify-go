# wasify

`wasify-go` is a Go library designed to streamline the interaction with WebAssembly (Wasm) modules by providing a developer-friendly API. It abstracts the [Wazero](https://github.com/tetratelabs/wazero) runtime, paving the way for potential support for multiple runtimes in the future, thereby enabling the execution of modules across various Wasm runtimes. This capability facilitates the loading and execution of Wasm modules, the passing of parameters between host and guest environments.

wasify simplifies communication with the WebAssembly System Interface (WASI), eliminating the need for developers to delve into intricate details or communicate using numbers, which is the standard method of interaction with modules. This library significantly eases the process of running and working with wasm modules, which has traditionally been a less straightforward task.

## Installation

```bash
go get github.com/wasify-io/wasify-go
```

## Usage

### main.go

In `main.go`, a Wasm module is loaded into the `wasify` runtime (abstraction of original wasm runtime), and the guest function `greet` is invoked. The host function `hostLog` is also defined in this file, which takes two parameters from the guest function, prints them, and returns two messages.

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
                Name: "hostLog",
                Callback: func(ctx context.Context, m wasify.ModuleProxy, params wasify.Params) wasify.Results {

                    fmt.Println("Host func param 0: ", string(params[0].Value))
                    fmt.Println("Host func param 1: ", string(params[1].Value))

                    return m.Return(
                        []byte("Hello"),
                        []byte("There!"),
                    )

                },
                Params:  []wasify.ValueType{wasify.ValueTypeByte, wasify.ValueTypeByte},
                Returns: []wasify.ValueType{wasify.ValueTypeByte, wasify.ValueTypeByte},
            },
        },
    })

    defer module.Close(ctx)

    // Call guest function
    module.GuestFunction(ctx, "greet").Invoke()

}
```

### Wasm Module Example

In `module/example.go`, a host function `hostLog` is defined and a guest function `greet` is exported. The `greet` function calls the `hostLog` function with two string messages and prints the results returned from the host function.

```go
package main

import (
    "fmt"

    "github.com/wasify-io/wasify-go/mdk"
)

func main() {}

//go:wasmimport myEnv hostLog
func hostLog(mdk.ArgOffset, mdk.ArgOffset) mdk.ResultOffset

//export greet
func _greet() {
    resultOffset := hostLog(mdk.Arg("Hello"), mdk.Arg("World"))

    results := mdk.Results(resultOffset)

    for i, result := range results {
        fmt.Printf("Guest func result %d: %s\r\n", i, string(result.Data))
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

