// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024 The Ebitengine Authors

#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>

// Empty is empty
struct Empty {};

// NoStruct tests that an empty struct doesn't cause issues
uint64_t NoStruct(struct Empty e) {
    return 0xdeadbeef;
}

struct EmptyEmpty {
    struct Empty x;
};

uint64_t EmptyEmpty(struct EmptyEmpty e) {
    return 0xdeadbeef;
}

uint64_t EmptyEmptyWithReg(unsigned int x, struct EmptyEmpty e, unsigned int y) {
    return (x << 16) | y;
}

// GreaterThan16Bytes is 24 bytes on 64 bit systems
struct GreaterThan16Bytes {
   int64_t  *x, *y, *z;
};

// GreaterThan16Bytes is a basic test for structs bigger than 16 bytes
uint64_t GreaterThan16Bytes(struct GreaterThan16Bytes g) {
    return *g.x + *g.y + *g.z;
}

// AfterRegisters tests to make sure that structs placed on the stack work properly
uint64_t AfterRegisters(intptr_t a, intptr_t b, intptr_t c, intptr_t d, intptr_t e, intptr_t f, intptr_t g, intptr_t h, struct GreaterThan16Bytes bytes) {
    intptr_t registers = a + b + c + d + e + f + g + h;
    int64_t stack =  *bytes.x + *bytes.y + *bytes.z;
    if (registers != stack) {
        return 0xbadbad;
    }
    if (stack != 0xdeadbeef) {
        return 0xcafebad;
    }
    return stack;
}

uint64_t BeforeRegisters(struct GreaterThan16Bytes bytes, int64_t a, int64_t b) {
    return *bytes.x + *bytes.y + *bytes.z + a + b;
}

struct GreaterThan16BytesStruct {
    struct {
        int64_t  *x, *y, *z;
    } a ;
};

uint64_t GreaterThan16BytesStruct(struct GreaterThan16BytesStruct g) {
    return *(g.a.x) + *(g.a.y) + *(g.a.z);
}

struct IntLessThan16Bytes {
    int64_t x, y;
};

uint64_t IntLessThan16Bytes(struct IntLessThan16Bytes l) {
    return l.x + l.y;
}

struct FloatLessThan16Bytes {
    float x, y;
};

float FloatLessThan16Bytes(struct FloatLessThan16Bytes f) {
    return f.x + f.y;
}

struct ThreeSmallFields {
    float x, y, z;
};

float ThreeSmallFields(struct ThreeSmallFields f) {
    return f.x + f.y + f.z;
}

struct FloatAndInt {
    float x;
    int   y;
};

float FloatAndInt(struct FloatAndInt f) {
    return f.x + f.y;
}

struct DoubleStruct {
    double x;
};

double DoubleStruct(struct DoubleStruct d) {
    return d.x;
}

struct TwoDoubleStruct {
    double x, y;
};

double TwoDoubleStruct(struct TwoDoubleStruct d) {
    return d.x + d.y;
}

struct TwoDoubleTwoStruct {
    struct {
        double x, y;
    } s;
};

double TwoDoubleTwoStruct(struct TwoDoubleTwoStruct d) {
    return d.s.x + d.s.y;
}

struct ThreeDoubleStruct {
    double x, y, z;
};

double ThreeDoubleStruct(struct ThreeDoubleStruct d) {
    return d.x + d.y + d.z;
}

struct LargeFloatStruct {
    double a, b, c, d, e, f;
};

double LargeFloatStruct(struct LargeFloatStruct s) {
    return s.a + s.b + s.c + s.d + s.e + s.f;
}

double LargeFloatStructWithRegs(double a, double b, double c, struct LargeFloatStruct s) {
    return a + b + c + s.a + s.b + s.c + s.d + s.e + s.f;
}

struct Rect {
    double x, y, w, h;
};

double Rectangle(struct Rect rect) {
    return rect.x + rect.y + rect.w + rect.h;
}

double RectangleSubtract(struct Rect rect) {
    return (rect.x + rect.y) - (rect.w + rect.h);
}

double RectangleWithRegs(double a, double b, double c, double d, double e, struct Rect rect) {
    return a + b + c + d + e + rect.x + rect.y + rect.w + rect.h;
}

struct FloatArray {
    double a[2];
};

double FloatArray(struct FloatArray f) {
    return f.a[0] + f.a[1];
}

struct UnsignedChar4Bytes {
    unsigned char a, b, c, d;
};

unsigned int UnsignedChar4Bytes(struct UnsignedChar4Bytes b) {
    return (((int)b.a)<<24) | (((int)b.b)<<16) | (((int)b.c)<<8) | (((int)b.d)<<0);
}

struct UnsignedChar4BytesStruct {
    struct {
        unsigned char a;
    } x;
    struct {
        unsigned char b;
    } y;
    struct {
        unsigned char c;
    } z;
    struct {
        unsigned char d;
    } w;
};

unsigned int UnsignedChar4BytesStruct(struct UnsignedChar4BytesStruct b) {
    return (((int)b.x.a)<<24) | (((int)b.y.b)<<16) | (((int)b.z.c)<<8) | (((int)b.w.d)<<0);
}

struct Short {
    unsigned short a, b, c, d;
};

uint64_t Short(struct Short s) {
    return (uint64_t)s.a << 48 | (uint64_t)s.b << 32 | (uint64_t)s.c << 16 | (uint64_t)s.d << 0;
}

struct Int {
    unsigned int a, b;
};

uint64_t Int(struct Int i) {
    return (uint64_t)i.a << 32 | (uint64_t)i.b << 0;
}

struct Long {
    uint64_t a;
};

uint64_t Long(struct Long l) {
    return l.a;
}

struct Char8Bytes {
    signed char a, b, c, d, e, f, g, h;
};

int Char8Bytes(struct Char8Bytes b) {
    return (int)b.a + (int)b.b + (int)b.c + (int)b.d + (int)b.e + (int)b.f + (int)b.g + (int)b.h;
}

struct Odd {
    unsigned char a, b, c;
};

int Odd(struct Odd o) {
    return (int)o.a + (int)o.b + (int)o.c;
}

struct Char2Short1 {
    unsigned char a, b;
    unsigned short c;
};

int Char2Short1s(struct Char2Short1 s) {
    return (int)s.a + (int)s.b + (int)s.c;
}

struct SignedChar2Short1 {
    signed char a, b;
    signed short c;
};

int SignedChar2Short1(struct SignedChar2Short1 s) {
    return s.a + s.b + s.c;
}

struct Array4UnsignedChars {
    unsigned char a[4];
};

unsigned int Array4UnsignedChars(struct Array4UnsignedChars a) {
    return (((int)a.a[0])<<24) | (((int)a.a[1])<<16) | (((int)a.a[2])<<8) | (((int)a.a[3])<<0);
}

struct Array3UnsignedChar {
    unsigned char a[3];
};

unsigned int Array3UnsignedChars(struct Array3UnsignedChar a) {
    return (((int)a.a[0])<<24) | (((int)a.a[1])<<16) | (((int)a.a[2])<<8) | 0xef;
}

struct Array2UnsignedShort {
    unsigned short a[2];
};

unsigned int Array2UnsignedShorts(struct Array2UnsignedShort a) {
    return (((int)a.a[0])<<16) | (((int)a.a[1])<<0);
}

struct Array4Chars {
    signed char a[4];
};

int Array4Chars(struct Array4Chars a) {
    return (int)a.a[0] + (int)a.a[1] + (int)a.a[2] + (int)a.a[3];
}

struct Array2Short {
    short a[2];
};

int Array2Shorts(struct Array2Short a) {
    return (int)a.a[0] + (int)a.a[1];
}

struct Array3Short {
    short a[3];
};

int Array3Shorts(struct Array3Short a) {
    return (int)a.a[0] + (int)a.a[1] + (int)a.a[2];
}

struct BoolStruct {
    _Bool b;
};

_Bool BoolStruct(struct BoolStruct b) {
    return b.b;
}

struct BoolFloat {
    _Bool b;
    float f;
};

float BoolFloat(struct BoolFloat s) {
    if (s.b)
        return s.f;
    return -s.f;
}

struct Content {
      struct { double x, y; } point;
      struct { double width, height; } size;
};

uint64_t InitWithContentRect(int *win, struct Content c, int style, int backing, _Bool flag) {
  if (win == 0)
      return 0xBAD;
  if (!flag)
      return 0xF1A6; // FLAG
  return (uint64_t)(c.point.x + c.point.y + c.size.width + c.size.height) / (style - backing);
}

struct GoInt4 {
    intptr_t a, b, c, d;
};

intptr_t GoInt4(struct GoInt4 g) {
    return g.a + g.b - g.c + g.d;
}

struct GoUint4 {
    uintptr_t a, b, c, d;
};

uintptr_t GoUint4(struct GoUint4 g) {
    return g.a + g.b + g.c + g.d;
}

uintptr_t TakeGoUintAndReturn(uintptr_t a) {
    return a;
}

struct FloatAndBool {
    float value;
    _Bool has_value;
};

int FloatAndBool(struct FloatAndBool f) {
    return f.has_value;
}

struct FourInt32s {
    int32_t f0;
    int32_t f1;
    int32_t f2;
    int32_t f3;
};

int32_t FourInt32s(struct FourInt32s s) {
    return s.f0 + s.f1 + s.f2 + s.f3;
}

struct PointerWrapper {
    void* ctx;
};

uintptr_t ExtractPointer(struct PointerWrapper wrapper) {
    return (uintptr_t)wrapper.ctx;
}

struct TwoPointers {
    void* ptr1;
    void* ptr2;
};

uintptr_t AddPointers(struct TwoPointers wrapper) {
    return (uintptr_t)wrapper.ptr1 + (uintptr_t)wrapper.ptr2;
}

// Identity functions for round-trip testing of struct arguments

struct OneInt64 {
    int64_t a;
};

struct OneInt64 IdentityOneInt64(struct OneInt64 s) {
    return s;
}

struct IntLessThan16Bytes IdentityIntLessThan16Bytes(struct IntLessThan16Bytes s) {
    return s;
}

struct TwoDoubleStruct IdentityTwoDoubleStruct(struct TwoDoubleStruct s) {
    return s;
}

struct FourFloat32 {
    float a, b, c, d;
};

struct FourFloat32 IdentityFourFloat32(struct FourFloat32 s) {
    return s;
}

struct FloatAndInt IdentityFloatAndInt(struct FloatAndInt s) {
    return s;
}

struct ThreeInt64 {
    int64_t a, b, c;
};

struct ThreeInt64 IdentityThreeInt64(struct ThreeInt64 s) {
    return s;
}

struct PtrInt64Ptr {
    int64_t *a;
    int64_t b;
    int64_t *c;
};

struct PtrInt64Ptr IdentityPtrInt64Ptr(struct PtrInt64Ptr s) {
    return s;
}

struct IntLessThan16Bytes IdentityTwoInt64AfterPrims(int64_t x, double y, struct IntLessThan16Bytes s) {
    return s;
}

struct FloatLessThan16Bytes IdentityTwoFloat32AfterFloats(double x, double y, struct FloatLessThan16Bytes s) {
    return s;
}

struct Mixed5Args {
    int64_t *a;
    int32_t b;
    float c;
    int32_t d;
};

struct Mixed5Args IdentityMixed5Args(struct Mixed5Args s) {
    return s;
}
