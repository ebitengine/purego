// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 The Ebitengine Authors

#import <Foundation/Foundation.h>

struct Result {
    int64_t a, b, c;
};

@interface PuregoSuperStructBase : NSObject
- (struct Result)result;
@end

@implementation PuregoSuperStructBase
- (struct Result)result {
    return (struct Result){.a = 12, .b = 34, .c = 56};
}
@end

@interface PuregoSuperStructChild1 : PuregoSuperStructBase
@end

@implementation PuregoSuperStructChild1
- (struct Result)result {
    return (struct Result){.a = 78, .b = 90, .c = 12};
}
@end

@interface PuregoSuperStructChild2 : PuregoSuperStructChild1
@end

@implementation PuregoSuperStructChild2
@end
