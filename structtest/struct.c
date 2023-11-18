// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2023 The Ebitengine Authors

// Empty is empty
struct Empty {};

//NoStruct tests that an empty struct doesn't cause issues
unsigned long NoStruct(struct Empty e) {
    return 0xdeadbeef;
}

// GreaterThan16Bytes is 24 bytes on 64 bit systems
struct GreaterThan16Bytes {
   long  *x, *y, *z;
};

// GreaterThan16Bytes is a basic test for structs bigger than 16 bytes
unsigned long GreaterThan16Bytes(struct GreaterThan16Bytes g) {
    return *g.x + *g.y + *g.z;
}

// AfterRegisters tests to make sure that structs placed on the stack work properly
unsigned long AfterRegisters(long a, long b, long c, long d, long e, long f, long g, long h, struct GreaterThan16Bytes bytes) {
    long registers = a + b + c + d + e + f + g + h;
    long stack =  *bytes.x + *bytes.y + *bytes.z;
    if (registers != stack) {
        return 0xbadbad;
    }
    if (stack != 0xdeadbeef) {
        return 0xcafebad;
    }
    return stack;
}

struct IntLessThan16Bytes {
    long x, y;
};

unsigned long IntLessThan16Bytes(struct IntLessThan16Bytes l) {
    return l.x + l.y;
}

struct FloatLessThan16Bytes {
    float x, y;
};

float FloatLessThan16Bytes(struct FloatLessThan16Bytes f) {
    return f.x + f.y;
}

struct DoubleStruct {
    double x;
};

double DoubleStruct(struct DoubleStruct d) {
    return d.x;
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

struct Short {
    unsigned short a, b, c, d;
};

unsigned long Short(struct Short s) {
    return (long)s.a << 48 | (long)s.b << 32 | (long)s.c << 16 | (long)s.d << 0;
}

struct Int {
    unsigned int a, b;
};

unsigned long Int(struct Int i) {
    return (long)i.a << 32 | (long)i.b << 0;
}

struct Long {
    unsigned long a;
};

unsigned long Long(struct Long l) {
    return l.a;
}

struct Array4Chars {
    unsigned char a[4];
};

unsigned int Array4Chars(struct Array4Chars a) {
    return (((int)a.a[0])<<24) | (((int)a.a[1])<<16) | (((int)a.a[2])<<8) | (((int)a.a[3])<<0);
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