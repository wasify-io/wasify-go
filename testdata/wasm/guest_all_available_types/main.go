package main

import (
	"fmt"

	"github.com/wasify-io/wasify-go/mdk"
)

func main() {}

//export guestTest
func _guestTest(
	_bytes mdk.PackedData,
	_byte mdk.PackedData,
	_i32 mdk.PackedData,
	_i64 mdk.PackedData,
	_f32 mdk.PackedData,
	_f64 mdk.PackedData,
	_string mdk.PackedData,
) {

	v1 := mdk.ReadBytesPack(_bytes)
	v2 := mdk.ReadBytePack(_byte)
	v3 := mdk.ReadI32Pack(_i32)
	v4 := mdk.ReadI64Pack(_i64)
	v5 := mdk.ReadF32Pack(_f32)
	v6 := mdk.ReadF64Pack(_f64)
	v7 := mdk.ReadStringPack(_string)

	fmt.Println("Test Resu: ",
		v1,
		v2,
		v3,
		v4,
		v5,
		v6,
		v7,
	)

	defer mdk.Free(
		_bytes,
		_byte,
		_i32,
		_i64,
		_f32,
		_f64,
		_string,
	)
}
