// Copyright 2012 Jeff Hodges and Jeff Smick. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
package com.somethingsimilar.opposite_of_a_bloom_filter;

import java.nio.ByteBuffer;

import org.junit.Test;

import static org.junit.Assert.assertEquals;
import static org.junit.Assert.assertFalse;
import static org.junit.Assert.assertTrue;
import static org.junit.Assert.fail;

public class OoaBFilterTest {

  private class FElement implements OoaBFilter.Element {
    private final ByteBuffer byteBuffer;

    public FElement(int a, int b, int c) {
      this.byteBuffer = ByteBuffer.allocate(12);
      this.byteBuffer.putInt(a);
      this.byteBuffer.putInt(b);
      this.byteBuffer.putInt(c);
    }

    public ByteBuffer getByteBuffer() {
      byteBuffer.rewind();
      return byteBuffer;
    }
  }

  @Test
  public void testTheBasics() {
    OoaBFilter filter = new OoaBFilter(2, 12);
    FElement twentyNineId = new FElement(27, 28, 29);
    FElement thirtyId = new FElement(27, 28, 30);
    FElement thirtyThreeId = new FElement(27, 28, 33);
    assertFalse("nothing should be contained at all", filter.containsAndAdd(twentyNineId));
    assertTrue("now it should", filter.containsAndAdd(twentyNineId));
    assertFalse("false unless the hash collides", filter.containsAndAdd(thirtyId));
    assertTrue("original should still return true", filter.containsAndAdd(twentyNineId));
    assertTrue("new array should still return true", filter.containsAndAdd(thirtyId));

    // Handling collisions. {27, 28, 33} and {27, 28, 30} hash to the same index using the current
    // hash function inside ByteArrayFilter.
    assertFalse("colliding array returns false", filter.containsAndAdd(thirtyThreeId));
    assertTrue(
        "colliding array returns true in second call", filter.containsAndAdd(thirtyThreeId));
    assertFalse("original colliding array returns false", filter.containsAndAdd(thirtyId));
    assertTrue("original colliding array returns true", filter.containsAndAdd(thirtyId));
    assertFalse("colliding array returns false", filter.containsAndAdd(thirtyThreeId));
  }

  @Test
  public void testSizeRounding() {
    ByteArrayFilter filter = new ByteArrayFilter(3);
    assertEquals("3 should round to 4", 4, filter.getSize());
    filter = new ByteArrayFilter(4);
    assertEquals("4 should round to 4", 4, filter.getSize());
    filter = new ByteArrayFilter(129);
    assertEquals("129 should round to 256", 256, filter.getSize());
  }

  @Test
  public void testTooLargeSize() {
    int size = (1<<30) + 1;
    try {
      new ByteArrayFilter(size);
      fail("should have thrown an IllegalArgumentException");
    } catch (IllegalArgumentException e) {
      String msg =
          "array size may not be larger than 2**31-1, but will be rounded to larger. was " + size;
      assertEquals(msg, e.getMessage());
    }
  }

  @Test
  public void testTooSmallSize() {
    int size = 0;
    try {
      new ByteArrayFilter(size);
      fail("zero passed in directly should have thrown an IllegalArgumentException");
    } catch (IllegalArgumentException e) {
      String msg = "array size must be greater than zero, was 0";
      assertEquals("zero passed in directly should error correctly", msg, e.getMessage());
    }
  }
}

