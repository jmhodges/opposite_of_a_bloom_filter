// Copyright 2012 Jeff Hodges and Jeff Smick. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
package com.somethingsimilar.opposite_of_a_bloom_filter;

import java.math.RoundingMode;
import java.util.Arrays;
import java.nio.ByteBuffer;

import com.google.common.hash.HashCode;
import com.google.common.hash.HashFunction;
import com.google.common.hash.Hashing;
import com.google.common.math.IntMath;

/**
 * OoaBFilter is used to filter out duplicate elements from a given dataset or stream. It is
 * guaranteed to never return a false positive (that is, it will never say that an item has already
 * been seen by the filter when it has not) but may return a false negative.
 *
 * The check is syncronized on the individual buffer. Depending on the dataset, hash function
 * and size of the underlying array lock contention should be very low. Dataset/hash function
 * combinations that cause many collisions will result in more contention.
 *
 * OoaBFilter is thread-safe.
 */
public class OoaBFilter {
  /**
   * The interface that must be implemented by an element to be filtered.
   */
  public interface Element {
    /**
     * Provide a ByteBuffer representation of the element.
     *
     * Ensure the buffer is rewound or the equality check will not work correctly.
     *
     * @return A ByteBuffer that represents the element.
     */
    public ByteBuffer getByteBuffer();
  }

  private static final HashFunction HASH_FUNC = Hashing.murmur3_32();
  private final int sizeMask;
  private final ByteBuffer[] array;
  private static final int MAX_SIZE = 1 << 30;

  /**
   * Constructs a OoaBFilter with an underlying array of the given size, rounded up to the next
   * power of two.
   *
   * This rounding occurs because the hashing is much faster on an array the size of a power of two.
   *
   * @param size The size of the underlying array.
   * @param bufSize The size of the buffers occupying each slot in the array.
   */
  public OoaBFilter(int size, int bufSize) {
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
    this.array = new ByteBuffer[poweredSize];

    // pre-allocate a ByteBuffer for each slot in the array
    int i = 0;
    while (i < poweredSize) {
      array[i] = ByteBuffer.allocate(bufSize);
      i++;
    }
  }

  /**
   * Returns whether the given elemtn has been previously seen by this filter. That is, if a byte
   * buffer with the same bytes as elem has been passed to to this method before.
   *
   * This method may return false when it has seen an element before. This occurs if the element passed in
   * hashes to the same index in the underlying array as another element previously checked. On the
   * flip side, this method will never return true incorrectly.
   *
   * @param element The byte array that may have been previously seen.
   * @return Whether the element is contained in the OoaBFilter.
   */
  public boolean containsAndAdd(Element element) {
    ByteBuffer eBytes = element.getByteBuffer();
    HashCode code = HASH_FUNC.hashBytes(eBytes.array());
    int index = code.asInt() & sizeMask;

    boolean seen = true;
    ByteBuffer buffer = array[index];

    synchronized(buffer) {
      if (!buffer.equals(eBytes)) {
        seen = false;
        buffer.put(eBytes);
        buffer.rewind();
      }
    }

    return seen;
  }
}
