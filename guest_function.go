package wasify

import (
	"errors"
	"fmt"

	"github.com/wasify-io/wasify-go/internal/types"
	"github.com/wasify-io/wasify-go/internal/utils"
)

type GuestFunctionResult struct {
	multiPackedData uint64
	memory          Memory
}

// ReadPacks decodes the packedData from a GuestFunctionResult instance and retrieves a sequence of packed datas.
// NOTE: Frees multiPackedData, which means ReadPacks should be called once.
func (r GuestFunctionResult) ReadPacks() ([]PackedData, error) {

	if r.multiPackedData == 0 {
		return nil, errors.New("packedData is empty")
	}

	t, offsetU32, size := utils.UnpackUI64(uint64(r.multiPackedData))

	if t != types.ValueTypePack {
		err := fmt.Errorf("Can't unpack host data, the type is not a valueTypePack. expected %d, got %d", types.ValueTypePack, t)
		return nil, err
	}

	bytes, err := r.memory.ReadBytes(offsetU32, size)
	if err != nil {
		err := errors.Join(errors.New("ReadPacks error, can't read bytes:"), err)
		return nil, err
	}

	err = r.memory.FreePack(PackedData(r.multiPackedData))
	if err != nil {
		err := errors.Join(errors.New("ReadPacks error, can't free multiPackedData:"), err)
		return nil, err
	}

	packedDataArray := utils.BytesToUint64Array(bytes)

	// calculate the number of elements in the array
	count := size / 8

	results := make([]PackedData, count)

	// Iterate over the packedData, unpack and read data of each element into a Result
	for i, pd := range packedDataArray {
		results[i] = PackedData(pd)
	}

	return results, nil
}
