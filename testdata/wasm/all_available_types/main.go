package main

import (
	"github.com/wasify-io/wasify-go/mdk"
)

func main() {}

//go:wasmimport all_available_types hostTest
func hostTest(
	mdk.ArgData,
	mdk.ArgData,
	mdk.ArgData,
	mdk.ArgData,
	mdk.ArgData,
	mdk.ArgData,
	mdk.ArgData,
) mdk.ResultOffset

//export guestTest
func _guestTest() {
	hostTest(
		mdk.Arg([]byte("Guest: Wello Wasify!")),
		mdk.Arg(byte(1)),
		mdk.Arg(uint32(11)),
		mdk.Arg(uint64(2023)),
		mdk.Arg(float32(11.1)),
		mdk.Arg(float64(11.2023)),
		mdk.Arg("Guest: Wasify."),
	)
}
