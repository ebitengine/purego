// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 The Ebitengine Authors

#include <stdint.h>
#include <assert.h>
#include <stdio.h>
#include <string.h>

uint32_t stack_uint8_t(uint32_t a, uint32_t b, uint32_t c, uint32_t d, uint32_t e, uint32_t f, uint32_t g, uint32_t h, uint8_t i, uint8_t j, uint32_t k) {
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

uint32_t stack_string(uint32_t a, uint32_t b, uint32_t c, uint32_t d, uint32_t e, uint32_t f, uint32_t g, uint32_t h, const char *i) {
    assert(i != 0);
    assert(strcmp(i, "test") == 0);
    return a | b | c | d | e | f | g | h;
}
