package main

import (
	"fmt"

	"github.com/wasify-io/wasify-go/mdk"
)

func main() {}

//go:wasmimport myEnv hostLog
func hostLog(mdk.ArgData, mdk.ArgData) mdk.ResultOffset

//go:wasmimport myEnv hostInt
func hostInt() mdk.ResultOffset

//export greet
func _greet() {

	var results []mdk.Result

	val := hostInt()
	results = mdk.Results(val)
	for _, r := range results {
		fmt.Println("hostInt Res: ", r.Data)
	}

	resultOffset := hostLog(mdk.Arg(uint32(123)), mdk.Arg("ლუკა"))
	results = mdk.Results(resultOffset)
	for i, result := range results {
		fmt.Printf("Guest func result %d: %s \n", i, result.Data)
	}

}
