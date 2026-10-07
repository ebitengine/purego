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

// purego_blockarg_block returns a block that calls the block it is given with x + 1.
void *purego_blockarg_block(void) {
    void (^b)(void (^)(int64_t), int64_t) = ^(void (^h)(int64_t), int64_t x) {
        h(x + 1);
    };
    return Block_copy(b);
}

typedef struct {
    bool b;
    float f;
} BoolFloat;

// purego_boolfloat_block returns a block that returns a struct with padding between its members.
void *purego_boolfloat_block(void) {
    BoolFloat (^b)(float) = ^BoolFloat(float f) {
        BoolFloat r = {true, f * 2};
        return r;
    };
    return Block_copy(b);
}

// purego_many_block returns a block that takes more arguments than fit in registers.
void *purego_many_block(void) {
    int64_t (^b)(int64_t, int64_t, int64_t, int64_t, int64_t, int64_t, int64_t, int64_t, double, int64_t) =
        ^int64_t(int64_t a1, int64_t a2, int64_t a3, int64_t a4, int64_t a5, int64_t a6, int64_t a7, int64_t a8, double f, int64_t a9) {
            return a1 + a2 + a3 + a4 + a5 + a6 + a7 + a8 + (int64_t)f + a9;
        };
    return Block_copy(b);
}
