// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024 The Ebitengine Authors

#include <stdint.h>

struct Empty{};

struct Empty ReturnEmpty() {
    struct Empty e = {};
    return e;
}

struct ThreeShorts{
    int16_t a, b, c;
};

struct ThreeShorts ReturnThreeShorts(int16_t a, int16_t b, int16_t c) {
    struct ThreeShorts e = {a, b, c};
    return e;
}

struct FourShorts{
    int16_t a, b, c, d;
};

struct FourShorts ReturnFourShorts(int16_t a, int16_t b, int16_t c, int16_t d) {
    struct FourShorts e = {a, b, c, d};
    return e;
}

struct OneLong{
    int64_t a;
};

struct OneLong ReturnOneLong(int64_t a) {
    struct OneLong e = {a};
    return e;
}

struct TwoLongs{
    int64_t a, b;
};

struct TwoLongs ReturnTwoLongs(int64_t a, int64_t b) {
    struct TwoLongs e = {a, b};
    return e;
}

struct ThreeLongs{
    int64_t a, b, c;
};

struct ThreeLongs ReturnThreeLongs(int64_t a, int64_t b, int64_t c) {
    struct ThreeLongs e = {a, b, c};
    return e;
}

struct ThreeDoubles{
    double a, b, c;
};

struct ThreeDoubles ReturnThreeDoubles(double a, double b, double c) {
    struct ThreeDoubles e = {a, b, c};
    return e;
}