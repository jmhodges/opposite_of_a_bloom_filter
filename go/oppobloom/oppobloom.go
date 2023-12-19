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
	"unsafe"
)

type Filter struct {
	array    []*[]byte
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
	slice := make([]*[]byte, size)
	sizeMask := uint32(size - 1)
	return &Filter{slice, sizeMask}, nil
}

func (f *Filter) Contains(id []byte) bool {
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
