package oppobloom

import "sync/atomic"

// Like Filter, but has additional ability to expose the number of items currently
// in the filter, as well as the total size of the data held by the filter
type SizeTrackingFilter struct {
	underlying *StandardFilter
	numEntries uint32
	bytesUsed  uint64
}

func NewSizeTrackingFilter(size int) (*SizeTrackingFilter, error) {
	underlying, err := NewFilter(size)
	if err != nil {
		return nil, err
	}
	ret := SizeTrackingFilter{
		underlying: underlying,
		numEntries: 0,
		bytesUsed:  0,
	}
	return &ret, nil
}

func (f *SizeTrackingFilter) Contains(id []byte) bool {
	ret, _ := f.ContainsCollision(id)
	return ret
}

func (f *SizeTrackingFilter) ContainsCollision(id []byte) (contains bool, collision bool) {
	var oldId []byte
	contains, collision, oldId = f.underlying.containsCollisionOldVal(id)
	if !contains && !collision {
		atomic.AddUint32(&f.numEntries, 1)
	}
	var bytesUsedDelta int64 = int64(len(id)) - int64(len(oldId))
	if bytesUsedDelta < 0 {
		atomic.AddUint64(&f.bytesUsed, ^uint64((-1*bytesUsedDelta)-1))
	} else {
		atomic.AddUint64(&f.bytesUsed, uint64(bytesUsedDelta))
	}
	return contains, collision
}

// Indicates how many entries have been added to the set. This will increment when new entries
// are added, and they do not collide with existing entries.
func (f *SizeTrackingFilter) NumEntries() uint32 {
	return atomic.LoadUint32(&f.numEntries)
}

// Returns the total size of the data held by the Filter.
func (f *SizeTrackingFilter) BytesUsed() uint64 {
	return atomic.LoadUint64(&f.bytesUsed)
}

func (f *SizeTrackingFilter) Size() int {
	return f.underlying.Size()
}
