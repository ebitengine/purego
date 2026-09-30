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
