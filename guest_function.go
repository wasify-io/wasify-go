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

func (r GuestFunctionResult) ReadResults() (Results, error) {

	if r.packedData == 0 {
		return nil, errors.New("packedData is empty")
	}

	t, offsetU32, size := utils.UnpackUI64(uint64(r.packedData))

	if t != types.ValueTypePack {
		err := fmt.Errorf("can't unpack host data, the type is not a valueTypePack. expected %d, got %d", types.ValueTypePack, t)
		return nil, err
	}

	fmt.Println("r.packedData", r.packedData, "T:", t, "offset32:", offsetU32, "size:", size)

	// calculate the number of elements in the array
	count := size / 8

	// FIXME: Read packedDatas and unpack each value
	bytes, _ := r.memory.ReadBytes(offsetU32, size)

	packedDatas := utils.BytesToUint64Array(bytes)

	data := make(Results, count)

	// Iterate over the packedData, unpack and read data of each element into a Result
	for i, pd := range packedDatas {

		_, _, d, _ := r.memory.Read(pd)

		data[i] = d
	}

	return data, nil
}

func (r *GuestFunctionResult) Free() {

}
