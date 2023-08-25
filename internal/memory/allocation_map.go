package memory

import (
	"github.com/puzpuzpuz/xsync/v2"
)

// This is a custom map-like structure designed for managing allocations.
// It keeps track of offset and size values, and provides methods
// for storing, loading, deleting, and calculating the total size and etc.
//
// AllocationMap is employed to monitor allocations made for parameters and return values
// within host functions. These allocations can be automatically cleared later,
// relieving users from the need to manually manage them.
type AllocationMap[K xsync.IntegerConstraint, V xsync.IntegerConstraint] struct {
	Map  *xsync.MapOf[K, V]
	Size V
}

func NewAllocationMap[K xsync.IntegerConstraint, V xsync.IntegerConstraint]() *AllocationMap[K, V] {
	return &AllocationMap[K, V]{
		Map: xsync.NewIntegerMapOf[K, V](),
	}
}

func (am *AllocationMap[K, V]) Store(offset K, size V) {
	am.Map.Store(offset, size)
	am.Size += size
}

func (am *AllocationMap[K, V]) Load(offset K) (V, bool) {
	return am.Map.Load(offset)
}

func (am *AllocationMap[K, V]) Delete(offset K) {
	v, _ := am.Map.LoadAndDelete(offset)
	am.Size -= v
}

func (am *AllocationMap[K, V]) TotalSize() V {
	return am.Size
}
