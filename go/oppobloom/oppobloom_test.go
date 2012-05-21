// Copyright 2012 Jeff Hodges. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
package oppobloom

import (
	"testing"
)

func TestTheBasics(t *testing.T) {
	f, _ := NewFilter(2)
    twentyNineId := []byte{27, 28, 29}
    thirtyId := []byte{27, 28, 30}
    thirtyThreeId := []byte{27, 28, 33}
	shouldNotContain(t, "nothing should be contained at all", f, twentyNineId)
    shouldContain(t, "now it should", f, twentyNineId)
    shouldNotContain(t, "false unless the hash collides", f, thirtyId)
    shouldContain(t, "original should still return true", f, twentyNineId)
    shouldContain(t, "new array should still return true", f, thirtyId)

    // Handling collisions. {27, 28, 33} and {27, 28, 30} hash to the same
    // index using the current hash function inside Filter.
    shouldNotContain(t, "colliding array returns false", f, thirtyThreeId)
    shouldContain(t, 
        "colliding array returns true in second call", f, thirtyThreeId)
    shouldNotContain(t, "original colliding array returns false", f, thirtyId)
    shouldContain(t, "original colliding array returns true", f, thirtyId)
    shouldNotContain(t, "colliding array returns false", f, thirtyThreeId)
}

func TestSizeRounding(t *testing.T) {
    f, _ := NewFilter(3);
    if f.Size() != 4 {
		t.Errorf("3 should round to 4, rounded to: ", f.Size())
	}
	f, _ = NewFilter(4);
	if f.Size() != 4 {
		t.Errorf("4 should round to 4", f.Size())
	}
	f, _ = NewFilter(129)
	if f.Size() != 256 {
		t.Errorf("129 should round to 256", f.Size())
	}
}

func TestTooLargeSize(t *testing.T) {
    size := (1<<30) + 1;
    f, err := NewFilter(size)
	if (err != ErrSizeTooLarge) {
		t.Errorf("did not error out on a too-large filter size")
	}
	if (f != nil) {
		t.Errorf("did not return nil on a too-large filter size")
	}
}

func TestTooSmallSize(t *testing.T) {
    f, err := NewFilter(0)
	if (err != ErrSizeTooSmall) {
		t.Errorf("did not error out on a too small filter size")
	}
	if (f != nil) {
		t.Errorf("did not return nil on a too small filter size")
	}
}

func shouldContain(t *testing.T, msg string, f *Filter, id []byte) {
	if !f.Contains(id) {
		t.Errorf("should contain, %s: id %v, array: %v", msg, id, f.array)
	}
}

func shouldNotContain(t *testing.T, msg string, f *Filter, id []byte) {
	if f.Contains(id) {
		t.Errorf("should not contain, %s: %v", msg, id)
	}
}
