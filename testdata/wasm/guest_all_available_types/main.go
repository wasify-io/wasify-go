package main

import (
	"fmt"

	"github.com/wasify-io/wasify-go/mdk"
)

func main() {}

//export guestTest
func _guestTest(
	_bytes mdk.ArgData,
	_byte mdk.ArgData,
	_i32 mdk.ArgData,
	_i64 mdk.ArgData,
	_f32 mdk.ArgData,
	_f64 mdk.ArgData,
	_string mdk.ArgData,
	_any mdk.ArgData,
) {

	v1, _ := mdk.ReadBytes(_bytes)
	v2 := mdk.ReadByte(_byte)
	v3 := mdk.ReadI32(_i32)
	v4 := mdk.ReadI64(_i64)
	v5 := mdk.ReadF32(_f32)
	v6 := mdk.ReadF64(_f64)
	v7, _ := mdk.ReadString(_string)
	v8, _ := mdk.ReadAny(_any)

	fmt.Println("Test Resu: ",
		v1,
		v2,
		v3,
		v4,
		v5,
		v6,
		v7,
		v8,
	)

	defer mdk.Free(
		_bytes,
		_byte,
		_i32,
		_i64,
		_f32,
		_f64,
		_string,
		_any,
	)
}
