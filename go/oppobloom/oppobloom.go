// Copyright 2012 Jeff Hodges. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package oppobloom implements a filter data structure that may report false
// negatives but no false positives.
package oppobloom

import (
	"bytes"
	"crypto/md5"
	"errors"
	"hash"
	"math"
	"sync/atomic"
	"unsafe"
)

type Filter struct {
	array      []*[]byte
	sizeMask   uint32
	numEntries uint32
	bytesUsed  uint64
}

var ErrSizeTooLarge = errors.New("oppobloom: size given too large to round to a power of 2")
var ErrSizeTooSmall = errors.New("oppobloom: filter cannot have a zero or negative size")
var MaxFilterSize = 1 << 30

func NewFilter(size int) (*Filter, error) {
	if size > MaxFilterSize {
		return nil, ErrSizeTooLarge
	}
	if size <= 0 {
		return nil, ErrSizeTooSmall
	}
	// round to the next largest power of two
	size = int(math.Pow(2, math.Ceil(math.Log2(float64(size)))))
	slice := make([]*[]byte, size)
	sizeMask := uint32(size - 1)
	return &Filter{array: slice, sizeMask: sizeMask}, nil
}

// Adds the given bytes to the set, and indicates if they were already present in the set.
// A true value here is definitive; a false value may be a false negative.
func (f *Filter) Contains(id []byte) bool {
	ret, _ := f.ContainsCollision(id)
	return ret
}

// Like Contains, but also indicates if there was a collision on the key. If both are false,
// then you can be sure that it is not a false negative. It may also be interested to track
// how often collisons happen-- that tracking is left to external concerns.
func (f *Filter) ContainsCollision(id []byte) (contains bool, collision bool) {
	h := md5UintHash{md5.New()}
	h.Write(id)
	uindex := h.Sum32() & f.sizeMask
	index := int32(uindex)
	oldId := getAndSet(f.array, index, id)
	contains = bytes.Equal(oldId, id)
	collision = len(oldId) != 0 && !contains
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
func (f *Filter) NumEntries() uint32 {
	return atomic.LoadUint32(&f.numEntries)
}

// Returns the total size of the data held by the Filter.
func (f *Filter) BytesUsed() uint64 {
	return atomic.LoadUint64(&f.bytesUsed)
}

func (f *Filter) Size() int {
	return len(f.array)
}

type md5UintHash struct {
	hash.Hash // a hack with knowledge of how md5 works
}

func (m md5UintHash) Sum32() uint32 {
	sum := m.Sum(nil)
	x := uint32(sum[0])
	for _, val := range sum[1:3] {
		x = x << 3
		x += uint32(val)
	}
	return x
}

// Returns the id that was in the slice at the given index after putting the
// new id in the slice at that index, atomically.
func getAndSet(arr []*[]byte, index int32, id []byte) []byte {
	indexPtr := (*unsafe.Pointer)(unsafe.Pointer(&arr[index]))
	idUnsafe := unsafe.Pointer(&id)
	var oldId []byte
	for {
		oldIdUnsafe := atomic.LoadPointer(indexPtr)
		if atomic.CompareAndSwapPointer(indexPtr, oldIdUnsafe, idUnsafe) {
			oldIdPtr := (*[]byte)(oldIdUnsafe)
			if oldIdPtr != nil {
				oldId = *oldIdPtr
			}
			break
		}
	}
	return oldId
}
