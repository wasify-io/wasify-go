package memory

import (
	"sync"
)

// This is a custom map-like structure designed for managing allocations.
// It keeps track of offset and size values, and provides methods
// for storing, loading, deleting, and calculating the total size and etc.
//
// AllocationMap is employed to monitor allocations made for parameters and return values
// within host functions. These allocations can be automatically cleared later,
// relieving users from the need to manually manage them.
type AllocationMap[K uint32 | uint64, V uint32 | uint64] struct {
	Map  sync.Map
	Size V
}

func NewAllocationMap[K uint32 | uint64, V uint32 | uint64]() *AllocationMap[K, V] {
	return &AllocationMap[K, V]{
		Map: sync.Map{},
	}
}

func (am *AllocationMap[K, V]) Store(offset K, size V) {
	am.Map.Store(offset, size)
	am.Size += size
}

func (am *AllocationMap[K, V]) Load(offset K) (V, bool) {
	v, ok := am.Map.Load(offset)
	return v.(V), ok
}

func (am *AllocationMap[K, V]) Delete(offset K) {
	v, _ := am.Map.LoadAndDelete(offset)
	am.Size -= v.(V)
}

func (am *AllocationMap[K, V]) TotalSize() V {
	return am.Size
}
