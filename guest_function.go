package wasify

import (
	"errors"
	"fmt"

	"github.com/wasify-io/wasify-go/internal/types"
	"github.com/wasify-io/wasify-go/internal/utils"
)

type GuestFunctionResult struct {
	packedData uint64
	memory     Memory
}

func (r GuestFunctionResult) ReadResults() error {

	if r.packedData == 0 {
		return errors.New("packedData is empty")
	}

	t, offsetU32, _ := utils.UnpackUI64(uint64(r.packedData))

	if t != types.ValueTypePack {
		err := fmt.Errorf("can't unpack host data, the type is not a valueTypePack. expected %d, got %d", types.ValueTypePack, t)
		return err
	}

	// calculate the number of elements in the array
	// count := size / 8

	// FIXME: Read packedDatas and unpack each value
	_, _, _, _ = r.memory.Read(uint64(offsetU32))
	// fmt.Println(bytes)

	// data := make([]*Result, count)

	// // Iterate over the packedData, unpack and read data of each element into a Result
	// for i, pd := range packedData {

	// 	v, s := ReadAny(ArgData(pd))

	// 	data[i] = &Result{
	// 		Size: s,
	// 		Data: v,
	// 	}
	// }

	// return data

	return nil
}

func (r *GuestFunctionResult) Free() {

}
