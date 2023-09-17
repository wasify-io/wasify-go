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

// ReadResults decodes the packedData from a GuestFunctionResult instance and retrieves a sequence of results.
// Similar to mdk.ReadResults
func (r GuestFunctionResult) ReadResults() (Results, error) {

	if r.packedData == 0 {
		return nil, errors.New("packedData is empty")
	}

	t, offsetU32, size := utils.UnpackUI64(uint64(r.packedData))

	if t != types.ValueTypePack {
		err := fmt.Errorf("Can't unpack host data, the type is not a valueTypePack. expected %d, got %d", types.ValueTypePack, t)
		return nil, err
	}

	bytes, err := r.memory.ReadBytes(offsetU32, size)
	if err != nil {
		err := errors.Join(errors.New("ReadResults error, can't read bytes:"), err)
		return nil, err
	}

	packedDatas := utils.BytesToUint64Array(bytes)

	// calculate the number of elements in the array
	count := size / 8

	results := make(Results, count)

	// Iterate over the packedData, unpack and read data of each element into a Result
	for i, pd := range packedDatas {

		_, _, d, _ := r.memory.Read(pd)

		results[i] = d
	}

	return results, nil
}

// TODO: Free allocated date
func (r *GuestFunctionResult) Free() {

}
