// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 The Ebitengine Authors

#include <stdint.h>
#include <assert.h>
#include <stdio.h>
#include <string.h>

uint32_t stack_uint8_t(uint32_t a, uint32_t b, uint32_t c, uint32_t d, uint32_t e, uint32_t f, uint32_t g, uint32_t h, uint8_t i, uint8_t j, uint32_t k ) {
    assert(i == 1);
    assert(j == 2);
    assert(k == 1024);
    return a | b | c | d | e | f | g | h | (uint32_t) i | (uint32_t) j | k;
}

uint32_t reg_uint8_t(uint8_t a, uint8_t b, uint32_t c) {
    assert(a == 1);
    assert(b == 2);
    assert(c == 1024);
    return a | b | c;
}

uint32_t stack_string(uint32_t a, uint32_t b, uint32_t c, uint32_t d, uint32_t e, uint32_t f, uint32_t g, uint32_t h, const char * i) {
    assert(i != 0);
    assert(strcmp(i, "test") == 0);
    return a | b | c | d | e | f | g | h;
}

void stack_8i32_3strings(char* result, size_t size, int32_t a1, int32_t a2, int32_t a3, int32_t a4, int32_t a5, int32_t a6, int32_t a7, int32_t a8, const char* s1, const char* s2, const char* s3) {
    snprintf(result, size, "%d:%d:%d:%d:%d:%d:%d:%d:%s:%s:%s", a1, a2, a3, a4, a5, a6, a7, a8, s1, s2, s3);
}
