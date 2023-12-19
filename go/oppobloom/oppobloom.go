// Copyright 2012 Jeff Hodges. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package oppobloom implements a filter data structure that may report false
// negatives but no false positives.
package oppobloom

import (
	"bytes"
	"errors"
	"hash/fnv"
	"math"
	"sync/atomic"
)

type Filter struct {
	array    []atomic.Pointer[[]byte]
	sizeMask uint32
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
	slice := make([]atomic.Pointer[[]byte], size)
	sizeMask := uint32(size - 1)
	return &Filter{slice, sizeMask}, nil
}

// ContainsAndAdd adds the id to the filter and returns true if the id was
// already present in it. False positives are not possible but false negatives
// are (that is, this function will never incorrectly return true but may
// incorrectly return false). False negatives occur when the given id has been
// previously seen, but in the time since that id was last passed to this
// method, a different id that hashed to the same index in the filter was added.
//
// ContainsAndAdd is thread-safe.
func (f *Filter) ContainsAndAdd(id []byte) bool {
	h := fnv.New32()
	h.Write(id)
	uindex := h.Sum32() & f.sizeMask
	index := int32(uindex)
	oldId := getAndSet(f.array, index, id)
	return bytes.Equal(oldId, id)
}

func (f *Filter) Size() int {
	return len(f.array)
}

// Returns the id that was in the slice at the given index after putting the
// new id in the slice at that index, atomically.
func getAndSet(arr []atomic.Pointer[[]byte], index int32, id []byte) []byte {
	indexPtr := &arr[index]
	idUnsafe := &id
	var oldId []byte
	for {
		oldIdUnsafe := indexPtr.Load()
		if indexPtr.CompareAndSwap(oldIdUnsafe, idUnsafe) {
			if oldIdUnsafe != nil {
				oldId = *oldIdUnsafe
			}
			break
		}
	}
	return oldId
}
