//
// Copyright 2020 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package com.google.privacy.differentialprivacy;

import static com.google.common.base.Preconditions.checkArgument;

/** Mathematical utilities for generating secure DP noise. */
final class SecureNoiseMath {

  private SecureNoiseMath() {}

  /**
   * Returns the smallest power of 2 larger or equal to {@code x}. The value of {@code x} must be a
   * finite positive number not greater than 2^1023. The result of this method is guaranteed to be
   * an exact power of 2.
   */
  public static double ceilPowerOfTwo(double x) {
    checkArgument(x > 0.0, "Input must be positive. Provided value: %s", x);
    checkArgument(Double.isFinite(x), "Input must be finite. Provided value: %s", x);
    checkArgument(!Double.isNaN(x), "Input must be a number. Provided value: NaN");

    // The following bit masks are based on the bit layout of double values in Java, which according
    // to the IEEE standard is defined as "1*s 11*e 52*m" where "s" is the sign bit, "e" are the
    // exponent bits and "m" are the mantissa bits.
    final long exponentMask = 0x7ff0000000000000L;
    final long mantissaMask = 0x000fffffffffffffL;

    long bits = Double.doubleToLongBits(x);
    long mantissaBits = bits & mantissaMask;

    // Since x is a finite positive number, x is a power of 2 if and only if it has a mantissa of 0.
    if (mantissaBits == 0x0000000000000000L) {
      return x;
    }

    long exponentBits = bits & exponentMask;
    long maxExponentBits = Double.doubleToLongBits(Double.MAX_VALUE) & exponentMask;

    checkArgument(
        exponentBits < maxExponentBits,
        "Input must not be greater than 2^1023. Provided value: %s",
        x);

    // Increase the exponent bits by 1 to get the next power of 2. Note that this requires to add
    // 0x0010000000000000L to the exponent bits in order to skip the mantissa.
    return Double.longBitsToDouble(exponentBits + 0x0010000000000000L);
  }

  /**
   * Rounds {@code x} to the closest multiple of {@code granularity}, where {@code granularity} is a
   * power of 2. Because {@code granularity} must be a power of 2, the result is exact.
   */
  public static double roundToMultipleOfPowerOfTwo(double x, double granularity) {
    // y is a power of 2 if and only if it is a positive finite value with a mantissa of 0.
    checkArgument(
        granularity > 0.0
            && Double.isFinite(granularity)
            && (Double.doubleToLongBits(granularity) & 0x000fffffffffffffL) == 0L,
        "Granularity must be a power of 2. Provided value: %s",
        granularity);

    if (Math.abs(x / granularity) < 1L << 54) {
      // The absolute value of x / granularity is less than 2^54 and therefore it is in particular
      // less than 2^63 and can be rounded to a long value without risking an overflow.
      return Math.round(x / granularity) * granularity;
    } else {
      // The absolute value of x / granularity is greater or equal to 2^54 and therefore it is in
      // particular greater than 2^53, which in turn implies that x / granularity contains no
      // fractional bits. As a result no rounding is necessary to make x a multiple of the
      // granularity.
      return x;
    }
  }

  /**
   * Computes the closest double value that is larger than or equal to the provided long value.
   *
   * <p>Mapping from long to double for large long values (> 2^53) is inaccurate since they cannot be
   * represented as a double. Implicit/explicit conversion from long to double either rounds up or
   * down the long value to the nearest representable double. This function ensures that {@code n}
   * <= (double) {@code n} by computing the difference between the double value and {@code n}, if
   * the diff is negative, it rounds up the double value to the next representable double value
   * which is greater than {@code n}.
   */
  public static double nextLargerDouble(long n) {
    // Given long n = 288230376151711767 when converted into double it lies between
    // two representable numbers 288230376151711744 and 288230376151711808. It gets
    // rounded down to 288230376151711744 when converted to double. long diff is then
    // equal to -23. Since diff is negative, it gets rounded back up to 288230376151711808.
    double result = n;
    long dif = (long) result - n;
    return dif >= 0 ? result : Math.nextUp(result);
  }

  /**
   * See {@link #nextLargerDouble(long)}.
   *
   * <p>As opposed to the latter method, this computes the closest double value that is smaller than
   * or equal to the provided long value ( {@code n} >= (double) {@code n} ).
   */
  public static double nextSmallerDouble(long n) {
    // Given long n = 288230376151711790 when converted into double it lies between
    // two representable numbers 288230376151711744 and 288230376151711808. It gets
    // rounded up to 288230376151711808 when converted to double. long diff is then
    // equal to 18. Since diff is positive, it gets rounded back down to 288230376151711744.
    double result = n;
    long dif = (long) result - n;
    return dif <= 0 ? result : Math.nextDown(result);
  }
}
