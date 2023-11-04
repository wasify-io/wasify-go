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
