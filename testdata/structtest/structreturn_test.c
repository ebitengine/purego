// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024 The Ebitengine Authors

#include <stdint.h>

struct Empty{};

struct Empty ReturnEmpty() {
    struct Empty e = {};
    return e;
}

struct StructInStruct{
    struct{ int16_t a; } a;
    struct{ int16_t b; } b;
    struct{ int16_t c; } c;
};

struct StructInStruct ReturnStructInStruct(int16_t a, int16_t b, int16_t c) {
    struct StructInStruct e = {{a}, {b}, {c}};
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

struct OneFloat{
    float a;
};

struct TwoFloats{
    float a, b;
};

struct TwoFloats ReturnTwoFloats(float a, float b) {
    struct TwoFloats e = {a-b, a*b};
    return e;
}

struct ThreeFloats{
    float a, b, c;
};

struct ThreeFloats ReturnThreeFloats(float a, float b, float c) {
    struct ThreeFloats e = {a, b, c};
    return e;
}

struct OneFloat ReturnOneFloat(float a) {
    struct OneFloat e = {a};
    return e;
}

struct OneDouble{
    double a;
};

struct OneDouble ReturnOneDouble(double a) {
    struct OneDouble e = {a};
    return e;
}

struct TwoDoubles{
    double a, b;
};

struct TwoDoubles ReturnTwoDoubles(double a, double b) {
    struct TwoDoubles e = {a, b};
    return e;
}

struct ThreeDoubles{
    double a, b, c;
};

struct ThreeDoubles ReturnThreeDoubles(double a, double b, double c) {
    struct ThreeDoubles e = {a, b, c};
    return e;
}

struct FourDoubles{
    double a, b, c, d;
};

struct FourDoubles ReturnFourDoubles(double a, double b, double c, double d) {
    struct FourDoubles e = {a, b, c, d};
    return e;
}

struct FourDoublesInternal{
    struct {
        double a, b;
    } f;
    struct {
        double c, d;
    } g;
};

struct FourDoublesInternal ReturnFourDoublesInternal(double a, double b, double c, double d) {
    struct FourDoublesInternal e = { {a, b}, {c, d} };
    return e;
}

struct FiveDoubles{
    double a, b, c, d, e;
};

struct FiveDoubles ReturnFiveDoubles(double a, double b, double c, double d, double e) {
    struct FiveDoubles s = {a, b, c, d, e};
    return s;
}

struct OneFloatOneDouble{
    float a;
    double b;
};

struct OneFloatOneDouble ReturnOneFloatOneDouble(float a, double b) {
    struct OneFloatOneDouble e = {a, b};
    return e;
}

struct OneDoubleOneFloat{
    double a;
    float b;
};

struct OneDoubleOneFloat ReturnOneDoubleOneFloat(double a, float b) {
    struct OneDoubleOneFloat e = {a, b};
    return e;
}

struct Unaligned1{
    int8_t  a;
    int16_t b;
    int64_t c;
};

struct Unaligned1 ReturnUnaligned1(int8_t a, int16_t b, int64_t c) {
    struct Unaligned1 e = {a, b, c};
    return e;
}

struct Mixed1{
     float a;
     int32_t b;
};

struct Mixed1 ReturnMixed1(float a, int32_t b) {
    struct Mixed1 e = {a, b};
    return e;
}

struct Mixed2{
     float a;
     int32_t b;
     float c;
     int32_t d;
};

struct Mixed2 ReturnMixed2(float a, int32_t b, float c, int32_t d) {
    struct Mixed2 e = {a, b, c, d};
    return e;
}

struct Mixed3{
     float a;
     uint32_t b;
     double c;
};

struct Mixed3 ReturnMixed3(float a, uint32_t b, double c) {
    struct Mixed3 s = {a, b, c};
    return s;
}

struct Mixed4{
     double a;
     uint32_t b;
     float c;
};

struct Mixed4 ReturnMixed4(double a, uint32_t b, float c) {
    struct Mixed4 s = {a, b, c};
    return s;
}

struct Ptr1{
     int64_t *a;
     void *b;
};

struct Ptr1 ReturnPtr1(int64_t *a, void *b) {
    struct Ptr1 s = {a, b};
    return s;
}
