// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 The Ebitengine Authors

#import <Foundation/Foundation.h>
#include <Block.h>

typedef int64_t (^PuregoBlock)(int64_t, double);

// purego_heap_block returns a heap block that adds base, i, and f.
void *purego_heap_block(int64_t base) {
    PuregoBlock b = ^int64_t(int64_t i, double f) {
        return base + i + (int64_t)f;
    };
    return Block_copy(b);
}

// purego_with_stack_block calls cb with a block that is still on the stack.
// The block is only valid until cb returns.
void purego_with_stack_block(int64_t base, void (*cb)(void *block)) {
    PuregoBlock b = ^int64_t(int64_t i, double f) {
        return base + i + (int64_t)f;
    };
    cb((void *)b);
}

typedef struct {
    int64_t a, b, c, d;
} Big;

// purego_big_block returns a block that returns a struct in memory.
void *purego_big_block(void) {
    Big (^b)(int64_t) = ^Big(int64_t x) {
        Big r = {x, x + 1, x + 2, x + 3};
        return r;
    };
    return Block_copy(b);
}

// purego_fnptr_block returns a block that calls a function pointer.
void *purego_fnptr_block(void) {
    void (^b)(void (*)(void)) = ^(void (*f)(void)) {
        f();
    };
    return Block_copy(b);
}
