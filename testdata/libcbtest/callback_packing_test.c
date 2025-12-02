// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 The Ebitengine Authors

#include <stdint.h>

typedef int32_t (*callback_int32)(int32_t, int32_t, int32_t, int32_t, int32_t, int32_t, int32_t, int32_t,
                                  int32_t, int32_t, int32_t, int32_t);

int32_t callCallbackInt32Packing(const void *fp) {
    return ((callback_int32) (fp))(2, 3, 5, 7, 11, 13, 17, 19, 23, 29, 31, 37);
}

typedef void (*callback_mixed)(int64_t, int64_t, int64_t, int64_t, int64_t, int64_t, int64_t, int64_t, int32_t, int64_t, int32_t);

void callCallbackMixedPacking(const void *fp) {
    ((callback_mixed) (fp))(1, 2, 3, 4, 5, 6, 7, 8, 100, 200, 300);
}

typedef void (*callback_small_types)(int64_t, int64_t, int64_t, int64_t, int64_t, int64_t, int64_t, int64_t, _Bool, int8_t, uint8_t, int16_t, uint16_t, int32_t);

void callCallbackSmallTypes(const void *fp) {
    ((callback_small_types) (fp))(1, 2, 3, 4, 5, 6, 7, 8, 1, -42, 200, -1000, 50000, 123456);
}
