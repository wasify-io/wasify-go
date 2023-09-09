package main

import (
	"fmt"

	"github.com/wasify-io/wasify-go/mdk"
)

func main() {}

//export guestTest
func _guestTest(a, b mdk.ArgData) mdk.ResultOffset {

	v1, _ := mdk.ReadString(a)
	v2, _ := mdk.ReadBytes(b)
	fmt.Println("Test Resu: ", v1, string(v2))

	defer mdk.Free(a, b)

	return mdk.ResultOffset(55555)

}
