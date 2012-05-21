// Copyright 2012 Jeff Hodges. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
package com.somethingsimilar.opposite_of_a_bloom_filter;

import java.math.RoundingMode;
import java.util.Arrays;
import java.util.concurrent.atomic.AtomicReferenceArray;

import com.google.common.hash.HashCode;
import com.google.common.hash.HashFunction;
import com.google.common.hash.Hashing;
import com.google.common.math.IntMath;

/**
 * ByteArrayFilter is used to filter out duplicate byte arrays from a given dataset or stream. It is
 * guaranteed to never return a false positive (that is, it will never say that an item has already
 * been seen by the filter when it has not) but may return a false negative.
 *
 * ByteArrayFilter is thread-safe.
 */
public class ByteArrayFilter {
  private static final HashFunction HASH_FUNC = Hashing.murmur3_32();
  private final int sizeMask;
  private final AtomicReferenceArray<byte[]> array;
  private static final int MAX_SIZE = 1 << 30;

  /**
   * Constructs a ByteArrayFilter with an underlying array of the given size, rounded up to the next
   * power of two.
   *
   * This rounding occurs because the hashing is much faster on an array the size of a power of two.
   * If you really want a different sized array, used the AtomicReferenceArray constructor.
   *
   * @param size The size of the underlying array.
   */
  public ByteArrayFilter(int size) {
    if (size <= 0) {
      throw new IllegalArgumentException("array size must be greater than zero, was " + size);
    }
    if (size > MAX_SIZE) {
      throw new IllegalArgumentException(
          "array size may not be larger than 2**31-1, but will be rounded to larger. was " + size);
    }
    // round to the next largest power of two
    int poweredSize = IntMath.pow(2, IntMath.log2(size, RoundingMode.CEILING));
    this.sizeMask = poweredSize - 1;
    this.array = new AtomicReferenceArray<byte[]>(poweredSize);
  }

  /**
   * Returns whether the given byte array has been previously seen by this array. That is, if a byte
   * array with the same bytes as id has been passed to to this method before.
   *
   * This method may return false when it has seen an id before. This occurs if the id passed in
   * hashes to the same index in the underlying array as another id previously checked. On the
   * flip side, this method will never return true incorrectly.
   *
   * @param id The byte array that may have been previously seen.
   * @return Whether the byte array is contained in the ByteArrayFilter.
   */
  public boolean containsAndAdd(byte[] id) {
    HashCode code = HASH_FUNC.hashBytes(id);
    int index = Math.abs(code.asInt()) & sizeMask;
    byte[] oldId = array.getAndSet(index, id);
    return Arrays.equals(id, oldId);
  }

  /**
   * Returns the size of the underlying array. Welp.
   *
   * @return The size of the underlying array.
   */
  public int getSize() {
    return array.length();
  }
}
