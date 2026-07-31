// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 The Ebitengine Authors

#include <assert.h>
#include <inttypes.h>
#include <stdbool.h>
#include <stdint.h>
#include <stdio.h>
#include <string.h>

uint32_t stack_uint8_t(uint32_t a, uint32_t b, uint32_t c, uint32_t d, uint32_t e, uint32_t f, uint32_t g, uint32_t h, uint8_t i, uint8_t j, uint32_t k ) {
    assert(i == 1);
    assert(j == 2);
    assert(k == 1024);
    return a | b | c | d | e | f | g | h | (uint32_t)i | (uint32_t)j | k;
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

// HFA (Homogeneous Float Aggregate) struct with 2 floats
typedef struct {
    float x;
    float y;
} Float2;

// HFA struct with 4 floats
typedef struct {
    float x;
    float y;
    float z;
    float w;
} Float4;

// Non-HFA struct (mixed types)
typedef struct {
    int32_t a;
    float b;
} MixedStruct;

// Small struct that fits in one register
typedef struct {
    int32_t x;
    int32_t y;
} IntPair;

// Test: 8 int registers exhausted, then HFA struct on stack
void stack_8int_hfa2_stack(char *buf, size_t bufsize, int32_t a1, int32_t a2, int32_t a3, int32_t a4, int32_t a5, int32_t a6, int32_t a7, int32_t a8, Float2 f) {
    snprintf(buf, bufsize, "%d:%d:%d:%d:%d:%d:%d:%d:%.1f:%.1f",
             a1, a2, a3, a4, a5, a6, a7, a8, f.x, f.y);
}

// Test: 8 int registers exhausted, then multiple structs on stack
void stack_8int_2structs_stack(char *buf, size_t bufsize, int32_t a1, int32_t a2, int32_t a3, int32_t a4, int32_t a5, int32_t a6, int32_t a7, int32_t a8, IntPair p1, IntPair p2) {
    snprintf(buf, bufsize, "%d:%d:%d:%d:%d:%d:%d:%d:%d:%d:%d:%d",
             a1, a2, a3, a4, a5, a6, a7, a8, p1.x, p1.y, p2.x, p2.y);
}

// Test: 8 float registers exhausted, then HFA on stack
void stack_8float_hfa2_stack(char *buf, size_t bufsize, float f1, float f2, float f3, float f4, float f5, float f6, float f7, float f8, Float2 f) {
    snprintf(buf, bufsize, "%.1f:%.1f:%.1f:%.1f:%.1f:%.1f:%.1f:%.1f:%.1f:%.1f",
             f1, f2, f3, f4, f5, f6, f7, f8, f.x, f.y);
}

// Test: mixed - int regs exhausted, float struct can still use float regs
void stack_8int_hfa2_floatregs(char *buf, size_t bufsize, int32_t a1, int32_t a2, int32_t a3, int32_t a4, int32_t a5, int32_t a6, int32_t a7, int32_t a8, Float2 f) {
    snprintf(buf, bufsize, "%d:%d:%d:%d:%d:%d:%d:%d:%.1f:%.1f",
             a1, a2, a3, a4, a5, a6, a7, a8, f.x, f.y);
}

// Test: primitives and struct interleaved on stack
void stack_8int_int_struct_int(char *buf, size_t bufsize, int32_t a1, int32_t a2, int32_t a3, int32_t a4, int32_t a5, int32_t a6, int32_t a7, int32_t a8, int32_t a9, IntPair p, int32_t a10) {
    snprintf(buf, bufsize, "%d:%d:%d:%d:%d:%d:%d:%d:%d:%d:%d:%d",
             a1, a2, a3, a4, a5, a6, a7, a8, a9, p.x, p.y, a10);
}

// Test: HFA4 struct on stack (4 floats)
void stack_8int_hfa4_stack(char *buf, size_t bufsize, int32_t a1, int32_t a2, int32_t a3, int32_t a4, int32_t a5, int32_t a6, int32_t a7, int32_t a8, Float4 f) {
    snprintf(buf, bufsize, "%d:%d:%d:%d:%d:%d:%d:%d:%.1f:%.1f:%.1f:%.1f",
             a1, a2, a3, a4, a5, a6, a7, a8, f.x, f.y, f.z, f.w);
}

// Test: mixed type struct on stack
void stack_8int_mixed_struct(char *buf, size_t bufsize, int32_t a1, int32_t a2, int32_t a3, int32_t a4, int32_t a5, int32_t a6, int32_t a7, int32_t a8, MixedStruct m) {
    snprintf(buf, bufsize, "%d:%d:%d:%d:%d:%d:%d:%d:%d:%.1f",
             a1, a2, a3, a4, a5, a6, a7, a8, m.a, m.b);
}

void stack_10_int32(char *buf, size_t bufsize, int32_t a1, int32_t a2, int32_t a3, int32_t a4, int32_t a5, int32_t a6, int32_t a7, int32_t a8, int32_t a9, int32_t a10) {
    snprintf(buf, bufsize, "%d:%d:%d:%d:%d:%d:%d:%d:%d:%d",
             a1, a2, a3, a4, a5, a6, a7, a8, a9, a10);
}

void stack_11_int32(char *buf, size_t bufsize, int32_t a1, int32_t a2, int32_t a3, int32_t a4, int32_t a5, int32_t a6, int32_t a7, int32_t a8, int32_t a9, int32_t a10, int32_t a11) {
    snprintf(buf, bufsize, "%d:%d:%d:%d:%d:%d:%d:%d:%d:%d:%d",
             a1, a2, a3, a4, a5, a6, a7, a8, a9, a10, a11);
}

void stack_10_float32(char *buf, size_t bufsize, float f1, float f2, float f3, float f4, float f5, float f6, float f7, float f8, float f9, float f10) {
    snprintf(buf, bufsize, "%.1f:%.1f:%.1f:%.1f:%.1f:%.1f:%.1f:%.1f:%.1f:%.1f",
             f1, f2, f3, f4, f5, f6, f7, f8, f9, f10);
}

void stack_mixed_stack_4args(char *buf, size_t bufsize, int32_t a1, int32_t a2, int32_t a3, int32_t a4, int32_t a5, int32_t a6, int32_t a7, int32_t a8, const char *s1, bool b1, int32_t a9, const char *s2) {
    snprintf(buf, bufsize, "%d:%d:%d:%d:%d:%d:%d:%d:%s:%d:%d:%s",
             a1, a2, a3, a4, a5, a6, a7, a8, s1, b1, a9, s2);
}

void stack_20_int32(char *buf, size_t bufsize, int32_t a1, int32_t a2, int32_t a3, int32_t a4, int32_t a5, int32_t a6, int32_t a7, int32_t a8, int32_t a9, int32_t a10, int32_t a11, int32_t a12, int32_t a13, int32_t a14, int32_t a15, int32_t a16, int32_t a17, int32_t a18, int32_t a19, int32_t a20) {
    snprintf(buf, bufsize, "%d:%d:%d:%d:%d:%d:%d:%d:%d:%d:%d:%d:%d:%d:%d:%d:%d:%d:%d:%d",
             a1, a2, a3, a4, a5, a6, a7, a8, a9, a10, a11, a12, a13, a14, a15, a16, a17, a18, a19, a20);
}

void stack_25_int64_exceeds(char *buf, size_t bufsize, int64_t a1, int64_t a2, int64_t a3, int64_t a4, int64_t a5, int64_t a6, int64_t a7, int64_t a8, int64_t a9, int64_t a10, int64_t a11, int64_t a12, int64_t a13, int64_t a14, int64_t a15, int64_t a16, int64_t a17, int64_t a18, int64_t a19, int64_t a20, int64_t a21, int64_t a22, int64_t a23, int64_t a24, int64_t a25) {
    snprintf(buf, bufsize, "%" PRId64 ":%" PRId64 ":%" PRId64 ":%" PRId64 ":%" PRId64 ":%" PRId64 ":%" PRId64 ":%" PRId64 ":%" PRId64 ":%" PRId64 ":%" PRId64 ":%" PRId64 ":%" PRId64 ":%" PRId64 ":%" PRId64 ":%" PRId64 ":%" PRId64 ":%" PRId64 ":%" PRId64 ":%" PRId64 ":%" PRId64 ":%" PRId64 ":%" PRId64 ":%" PRId64 ":%" PRId64,
             a1, a2, a3, a4, a5, a6, a7, a8, a9, a10, a11, a12, a13, a14, a15, a16, a17, a18, a19, a20, a21, a22, a23, a24, a25);
}

uintptr_t stack_20_uintptr(
    uintptr_t a1, uintptr_t a2, uintptr_t a3, uintptr_t a4, uintptr_t a5,
    uintptr_t a6, uintptr_t a7, uintptr_t a8, uintptr_t a9, uintptr_t a10,
    uintptr_t a11, uintptr_t a12, uintptr_t a13, uintptr_t a14, uintptr_t a15,
    uintptr_t a16, uintptr_t a17, uintptr_t a18, uintptr_t a19, uintptr_t a20
) {
    return a1 + a2 + a3 + a4 + a5 + a6 + a7 + a8 + a9 + a10 +
           a11 + a12 + a13 + a14 + a15 + a16 + a17 + a18 + a19 + a20;
}

uintptr_t stack_32_uintptr(
    uintptr_t a1, uintptr_t a2, uintptr_t a3, uintptr_t a4, uintptr_t a5, uintptr_t a6, uintptr_t a7, uintptr_t a8,
    uintptr_t a9, uintptr_t a10, uintptr_t a11, uintptr_t a12, uintptr_t a13, uintptr_t a14, uintptr_t a15, uintptr_t a16,
    uintptr_t a17, uintptr_t a18, uintptr_t a19, uintptr_t a20, uintptr_t a21, uintptr_t a22, uintptr_t a23, uintptr_t a24,
    uintptr_t a25, uintptr_t a26, uintptr_t a27, uintptr_t a28, uintptr_t a29, uintptr_t a30, uintptr_t a31, uintptr_t a32
) {
    return a1 + a2 + a3 + a4 + a5 + a6 + a7 + a8 +
           a9 + a10 + a11 + a12 + a13 + a14 + a15 + a16 +
           a17 + a18 + a19 + a20 + a21 + a22 + a23 + a24 +
           a25 + a26 + a27 + a28 + a29 + a30 + a31 + a32;
}

double stack_32_mixed_int_float(
    uintptr_t a1, uintptr_t a2, uintptr_t a3, uintptr_t a4, uintptr_t a5, uintptr_t a6, uintptr_t a7, uintptr_t a8,
    uintptr_t a9, uintptr_t a10, uintptr_t a11, uintptr_t a12, uintptr_t a13, uintptr_t a14, uintptr_t a15, uintptr_t a16,
    double f1, double f2, double f3, double f4, double f5, double f6, double f7, double f8,
    double f9, double f10, double f11, double f12, double f13, double f14, double f15, double f16
) {
    return (double)a1 * 1 + (double)a2 * 2 + (double)a3 * 3 + (double)a4 * 4 +
           (double)a5 * 5 + (double)a6 * 6 + (double)a7 * 7 + (double)a8 * 8 +
           (double)a9 * 9 + (double)a10 * 10 + (double)a11 * 11 + (double)a12 * 12 +
           (double)a13 * 13 + (double)a14 * 14 + (double)a15 * 15 + (double)a16 * 16 +
           f1 * 17 + f2 * 18 + f3 * 19 + f4 * 20 +
           f5 * 21 + f6 * 22 + f7 * 23 + f8 * 24 +
           f9 * 25 + f10 * 26 + f11 * 27 + f12 * 28 +
           f13 * 29 + f14 * 30 + f15 * 31 + f16 * 32;
}

#if defined(_WIN32) && defined(__i386__)

#include <windows.h>

#define STDCALL __attribute__((stdcall))

static uint32_t win386_pointer_value = UINT32_C(0x89abcdef);

uint64_t win386_direct_mixed(
    int8_t i8, uint16_t u16, int32_t i32, uint64_t u64,
    float f32, double f64, uintptr_t pointer, bool boolean, const char *string
) {
    if (i8 != -101 || u16 != 60000 || i32 != INT32_C(-2000000000) ||
        u64 != UINT64_C(0xfedcba9876543210) || f32 != 1.25f || f64 != -2.5 ||
        pointer != (uintptr_t)&win386_pointer_value || !boolean ||
        strcmp(string, "purego-win386") != 0) {
        return 0;
    }
    return u64;
}

int64_t win386_return_int64(void) {
    return INT64_C(-0x123456789abcdef);
}

uint64_t win386_return_uint64(void) {
    return UINT64_C(0xfedcba9876543210);
}

float win386_return_float32(float value) {
    return value * 1.5f;
}

double win386_return_float64(double value) {
    return value * -2.25;
}

float win386_identity_float32(float value) {
    return value;
}

double win386_identity_float64(double value) {
    return value;
}

typedef struct {
    float value;
} win386_one_float32;

typedef struct {
    double value;
} win386_one_float64;

win386_one_float32 win386_identity_struct_float32(win386_one_float32 value) {
    return value;
}

win386_one_float64 win386_identity_struct_float64(win386_one_float64 value) {
    return value;
}

uint64_t win386_16_uint64(
    uint64_t a1, uint64_t a2, uint64_t a3, uint64_t a4,
    uint64_t a5, uint64_t a6, uint64_t a7, uint64_t a8,
    uint64_t a9, uint64_t a10, uint64_t a11, uint64_t a12,
    uint64_t a13, uint64_t a14, uint64_t a15, uint64_t a16
) {
    return a1 + a2 + a3 + a4 + a5 + a6 + a7 + a8 +
           a9 + a10 + a11 + a12 + a13 + a14 + a15 + a16;
}

double win386_16_float64(
    double a1, double a2, double a3, double a4,
    double a5, double a6, double a7, double a8,
    double a9, double a10, double a11, double a12,
    double a13, double a14, double a15, double a16
) {
    return a1 + a2 + a3 + a4 + a5 + a6 + a7 + a8 +
           a9 + a10 + a11 + a12 + a13 + a14 + a15 + a16;
}

uintptr_t win386_pointer(void) {
    return (uintptr_t)&win386_pointer_value;
}

typedef uint64_t (STDCALL *win386_stdcall_u64_callback)(
    int8_t, uint16_t, int32_t, uint64_t, float, double, uint32_t *, bool
);
typedef uint64_t (*win386_cdecl_u64_callback)(
    int8_t, uint16_t, int32_t, uint64_t, float, double, uint32_t *, bool
);
typedef int64_t (STDCALL *win386_stdcall_i64_callback)(int64_t, double);
typedef float (STDCALL *win386_stdcall_f32_callback)(uint64_t, float, double);
typedef double (*win386_cdecl_f64_callback)(uint64_t, float, double);
typedef uint64_t (STDCALL *win386_stdcall_16_u64_callback)(
    uint64_t, uint64_t, uint64_t, uint64_t,
    uint64_t, uint64_t, uint64_t, uint64_t,
    uint64_t, uint64_t, uint64_t, uint64_t,
    uint64_t, uint64_t, uint64_t, uint64_t
);
typedef float (STDCALL *win386_stdcall_identity_f32_callback)(float);
typedef double (*win386_cdecl_identity_f64_callback)(double);
typedef uint32_t (STDCALL *win386_stdcall_u32_callback)(uint32_t);
typedef uint64_t (STDCALL *win386_stdcall_reentrant_callback)(uint32_t);
typedef uint64_t (*win386_cdecl_reentrant_callback)(uint32_t);
typedef uint64_t (STDCALL *win386_stdcall_thread_callback)(uint64_t, float, double);

uint64_t win386_call_stdcall_u64(uintptr_t callback) {
    return ((win386_stdcall_u64_callback)callback)(
        -101, 60000, INT32_C(-2000000000), UINT64_C(0xfedcba9876543210),
        1.25f, -2.5, &win386_pointer_value, true
    );
}

uint64_t win386_call_cdecl_u64(uintptr_t callback) {
    return ((win386_cdecl_u64_callback)callback)(
        -101, 60000, INT32_C(-2000000000), UINT64_C(0xfedcba9876543210),
        1.25f, -2.5, &win386_pointer_value, true
    );
}

int64_t win386_call_stdcall_i64(uintptr_t callback) {
    return ((win386_stdcall_i64_callback)callback)(INT64_C(-0x123456789abcdef), 3.5);
}

float win386_call_stdcall_f32(uintptr_t callback) {
    return ((win386_stdcall_f32_callback)callback)(UINT64_C(0x123456789abcdef0), 1.25f, -2.5);
}

double win386_call_cdecl_f64(uintptr_t callback) {
    return ((win386_cdecl_f64_callback)callback)(UINT64_C(0x123456789abcdef0), 1.25f, -2.5);
}

uint64_t win386_call_stdcall_16_u64(uintptr_t callback) {
    return ((win386_stdcall_16_u64_callback)callback)(
        UINT64_C(0x100000001), UINT64_C(0x100000002), UINT64_C(0x100000003), UINT64_C(0x100000004),
        UINT64_C(0x100000005), UINT64_C(0x100000006), UINT64_C(0x100000007), UINT64_C(0x100000008),
        UINT64_C(0x100000009), UINT64_C(0x10000000a), UINT64_C(0x10000000b), UINT64_C(0x10000000c),
        UINT64_C(0x10000000d), UINT64_C(0x10000000e), UINT64_C(0x10000000f), UINT64_C(0x100000010)
    );
}

float win386_call_stdcall_identity_f32(uintptr_t callback, float value) {
    return ((win386_stdcall_identity_f32_callback)callback)(value);
}

double win386_call_cdecl_identity_f64(uintptr_t callback, double value) {
    return ((win386_cdecl_identity_f64_callback)callback)(value);
}

uint32_t win386_call_stdcall_u32(uintptr_t callback, uint32_t value) {
    return ((win386_stdcall_u32_callback)callback)(value);
}

uint64_t win386_call_stdcall_reentrant(uintptr_t callback, uint32_t depth) {
    return ((win386_stdcall_reentrant_callback)callback)(depth);
}

uint64_t win386_call_cdecl_reentrant(uintptr_t callback, uint32_t depth) {
    return ((win386_cdecl_reentrant_callback)callback)(depth);
}

typedef struct {
    win386_stdcall_thread_callback callback;
    uint64_t result;
} win386_thread_call;

static DWORD WINAPI win386_thread_entry(void *opaque) {
    win386_thread_call *call = (win386_thread_call *)opaque;
    call->result = call->callback(UINT64_C(0xfedcba9876543210), 1.25f, -2.5);
    return 0;
}

uint64_t win386_call_stdcall_on_thread(uintptr_t callback) {
    win386_thread_call call = {(win386_stdcall_thread_callback)callback, 0};
    HANDLE thread = CreateThread(NULL, 0, win386_thread_entry, &call, 0, NULL);
    if (thread == NULL) {
        return 0;
    }
    if (WaitForSingleObject(thread, INFINITE) != WAIT_OBJECT_0) {
        CloseHandle(thread);
        return 0;
    }
    CloseHandle(thread);
    return call.result;
}

typedef struct {
    uint32_t slots[32];
} win386_struct_32_slots;

uint32_t win386_take_struct_32_slots(win386_struct_32_slots value) {
    uint32_t sum = 0;
    for (size_t i = 0; i < 32; i++) {
        sum += value.slots[i];
    }
    return sum;
}

#endif
