// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 The Ebitengine Authors

#include <stdint.h>
#include <stdbool.h>

// Test: 12 int32 arguments - fills 8 int registers, then 4 go to stack
// With tight packing, the 4 stack args pack into 16 bytes (4x4)
// Without tight packing, they would use 32 bytes (4x8)
typedef int32_t (*callback_12_int32)(int32_t, int32_t, int32_t, int32_t,
                                      int32_t, int32_t, int32_t, int32_t,
                                      int32_t, int32_t, int32_t, int32_t);

int32_t callCallback12Int32(const void *fp) {
    // Using prime numbers to detect argument misalignment
    return ((callback_12_int32)(fp))(2, 3, 5, 7, 11, 13, 17, 19, 23, 29, 31, 37);
}

// Test: mixed int64 and int32 arguments with stack spillover
// 8 int64s fill registers, then int32/int64/int32 go to stack
// Tests alignment handling when mixing sizes on stack
typedef int64_t (*callback_mixed_stack)(int64_t, int64_t, int64_t, int64_t,
                                         int64_t, int64_t, int64_t, int64_t,
                                         int32_t, int64_t, int32_t);

int64_t callCallbackMixedStack(const void *fp) {
    return ((callback_mixed_stack)(fp))(1, 2, 3, 4, 5, 6, 7, 8, 100, 200, 300);
}

// Test: various small types on stack
// 8 int64s fill registers, then bool/int8/uint8/int16/uint16/int32 go to stack
// Tests byte-level packing of different small types
typedef int64_t (*callback_small_types)(int64_t, int64_t, int64_t, int64_t,
                                         int64_t, int64_t, int64_t, int64_t,
                                         bool, int8_t, uint8_t, int16_t, uint16_t, int32_t);

int64_t callCallbackSmallTypes(const void *fp) {
    return ((callback_small_types)(fp))(1, 2, 3, 4, 5, 6, 7, 8,
                                        true, -42, 200, -1000, 50000, 123456);
}

// Test: 10 int32 arguments - simpler case
// 8 fill registers, 2 go to stack (8 bytes packed vs 16 unpacked)
typedef int32_t (*callback_10_int32)(int32_t, int32_t, int32_t, int32_t,
                                      int32_t, int32_t, int32_t, int32_t,
                                      int32_t, int32_t);

int32_t callCallback10Int32(const void *fp) {
    return ((callback_10_int32)(fp))(1, 2, 3, 4, 5, 6, 7, 8, 9, 10);
}

// Test: 10 float64 arguments - fills 8 float registers, then 2 go to stack
// Callback returns int64 since purego callbacks don't support float returns
typedef int64_t (*callback_10_float64)(double, double, double, double,
                                        double, double, double, double,
                                        double, double);

int64_t callCallback10Float64(const void *fp) {
    return ((callback_10_float64)(fp))(1.5, 2.5, 3.5, 4.5, 5.5, 6.5, 7.5, 8.5, 9.5, 10.5);
}

// Test: 12 float32 arguments - fills 8 float registers, then 4 go to stack
// Callback returns int64 since purego callbacks don't support float returns
typedef int64_t (*callback_12_float32)(float, float, float, float,
                                        float, float, float, float,
                                        float, float, float, float);

int64_t callCallback12Float32(const void *fp) {
    return ((callback_12_float32)(fp))(1.0f, 2.0f, 3.0f, 4.0f, 5.0f, 6.0f,
                                        7.0f, 8.0f, 9.0f, 10.0f, 11.0f, 12.0f);
}
