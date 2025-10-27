// SPDX-License-Identifier: Apache-2.0
// Comprehensive Darwin ARM64 stack parameter tests

#include <stdint.h>
#include <stdbool.h>
#include <stdio.h>
#include <unistd.h>

// Helper to print test results to stderr
#define DEBUG_PRINT(fmt, ...) do { \
    char buf[256]; \
    int len = snprintf(buf, sizeof(buf), fmt "\n", ##__VA_ARGS__); \
    write(2, buf, len); \
} while(0)

// ============================================================================
// UNIFORM TYPE TESTS - All parameters same size
// ============================================================================

int8_t test_11_int8(int8_t a1, int8_t a2, int8_t a3, int8_t a4, int8_t a5,
                    int8_t a6, int8_t a7, int8_t a8, int8_t a9, int8_t a10, int8_t a11) {
    int8_t sum = a1 + a2 + a3 + a4 + a5 + a6 + a7 + a8 + a9 + a10 + a11;
    DEBUG_PRINT("test_11_int8: %d %d %d %d %d %d %d %d %d %d %d = %d",
                a1, a2, a3, a4, a5, a6, a7, a8, a9, a10, a11, sum);
    return sum;
}

int16_t test_11_int16(int16_t a1, int16_t a2, int16_t a3, int16_t a4, int16_t a5,
                      int16_t a6, int16_t a7, int16_t a8, int16_t a9, int16_t a10, int16_t a11) {
    int16_t sum = a1 + a2 + a3 + a4 + a5 + a6 + a7 + a8 + a9 + a10 + a11;
    DEBUG_PRINT("test_11_int16: %d %d %d %d %d %d %d %d %d %d %d = %d",
                a1, a2, a3, a4, a5, a6, a7, a8, a9, a10, a11, sum);
    return sum;
}

int32_t test_11_int32(int32_t a1, int32_t a2, int32_t a3, int32_t a4, int32_t a5,
                      int32_t a6, int32_t a7, int32_t a8, int32_t a9, int32_t a10, int32_t a11) {
    int32_t sum = a1 + a2 + a3 + a4 + a5 + a6 + a7 + a8 + a9 + a10 + a11;
    DEBUG_PRINT("test_11_int32: %d %d %d %d %d %d %d %d %d %d %d = %d",
                a1, a2, a3, a4, a5, a6, a7, a8, a9, a10, a11, sum);
    return sum;
}

int64_t test_11_int64(int64_t a1, int64_t a2, int64_t a3, int64_t a4, int64_t a5,
                      int64_t a6, int64_t a7, int64_t a8, int64_t a9, int64_t a10, int64_t a11) {
    int64_t sum = a1 + a2 + a3 + a4 + a5 + a6 + a7 + a8 + a9 + a10 + a11;
    DEBUG_PRINT("test_11_int64: %lld %lld %lld %lld %lld %lld %lld %lld %lld %lld %lld = %lld",
                a1, a2, a3, a4, a5, a6, a7, a8, a9, a10, a11, sum);
    return sum;
}

// ============================================================================
// EDGE CASE: Exactly 9 parameters (last one just hits stack)
// ============================================================================

int32_t test_9_int32(int32_t a1, int32_t a2, int32_t a3, int32_t a4, int32_t a5,
                     int32_t a6, int32_t a7, int32_t a8, int32_t a9) {
    int32_t sum = a1 + a2 + a3 + a4 + a5 + a6 + a7 + a8 + a9;
    DEBUG_PRINT("test_9_int32: sum = %d (expected %d)", sum, 45);
    return sum;
}

// ============================================================================
// EDGE CASE: Maximum 15 parameters
// ============================================================================

int32_t test_15_int32(int32_t a1, int32_t a2, int32_t a3, int32_t a4, int32_t a5,
                      int32_t a6, int32_t a7, int32_t a8, int32_t a9, int32_t a10,
                      int32_t a11, int32_t a12, int32_t a13, int32_t a14, int32_t a15) {
    int32_t sum = a1 + a2 + a3 + a4 + a5 + a6 + a7 + a8 + a9 + a10 + a11 + a12 + a13 + a14 + a15;
    DEBUG_PRINT("test_15_int32: sum = %d (expected %d)", sum, 120);
    return sum;
}

// ============================================================================
// MIXED TYPE TESTS
// ============================================================================

// 8 registers + 2 uint8 + 1 uint32 (tests packing of small types)
uint32_t test_mixed_8r_2u8_1u32(uint32_t r1, uint32_t r2, uint32_t r3, uint32_t r4,
                                 uint32_t r5, uint32_t r6, uint32_t r7, uint32_t r8,
                                 uint8_t s1, uint8_t s2, uint32_t s3) {
    uint32_t result = r1 | r2 | r3 | r4 | r5 | r6 | r7 | r8 | s1 | s2 | s3;
    DEBUG_PRINT("test_mixed_8r_2u8_1u32: result = %u (expected %u)", result, 2047);
    return result;
}

// 8 registers + 3 int16 (tests int16 packing)
int32_t test_mixed_8i32_3i16(int32_t r1, int32_t r2, int32_t r3, int32_t r4,
                              int32_t r5, int32_t r6, int32_t r7, int32_t r8,
                              int16_t s1, int16_t s2, int16_t s3) {
    int32_t sum = r1 + r2 + r3 + r4 + r5 + r6 + r7 + r8 + s1 + s2 + s3;
    DEBUG_PRINT("test_mixed_8i32_3i16: sum = %d (expected %d)", sum, 42);
    return sum;
}

// 8 registers + int8, int16, int32 (tests mixed size packing)
int32_t test_mixed_varied(int32_t r1, int32_t r2, int32_t r3, int32_t r4,
                          int32_t r5, int32_t r6, int32_t r7, int32_t r8,
                          int8_t s1, int16_t s2, int32_t s3) {
    int32_t sum = r1 + r2 + r3 + r4 + r5 + r6 + r7 + r8 + s1 + s2 + s3;
    DEBUG_PRINT("test_mixed_varied: sum = %d (expected %d)", sum, 42);
    return sum;
}

// ============================================================================
// BOOL TESTS (bool is typically 1 byte but may have special handling)
// ============================================================================

int32_t test_11_bool(bool b1, bool b2, bool b3, bool b4, bool b5,
                     bool b6, bool b7, bool b8, bool b9, bool b10, bool b11) {
    // Count true values
    int32_t count = b1 + b2 + b3 + b4 + b5 + b6 + b7 + b8 + b9 + b10 + b11;
    DEBUG_PRINT("test_11_bool: count = %d (expected %d)", count, 6);
    return count;
}

// Mixed: 8 int32 + 3 bool (tests bool packing with integers)
int32_t test_mixed_8i32_3bool(int32_t i1, int32_t i2, int32_t i3, int32_t i4,
                               int32_t i5, int32_t i6, int32_t i7, int32_t i8,
                               bool b1, bool b2, bool b3) {
    int32_t sum = i1 + i2 + i3 + i4 + i5 + i6 + i7 + i8 + (b1 ? 1 : 0) + (b2 ? 1 : 0) + (b3 ? 1 : 0);
    DEBUG_PRINT("test_mixed_8i32_3bool: sum = %d (expected %d)", sum, 39);
    return sum;
}

// ============================================================================
// POINTER TESTS
// ============================================================================

uint32_t test_9_ptrs(void* p1, void* p2, void* p3, void* p4, void* p5,
                     void* p6, void* p7, void* p8, void* p9) {
    // Count non-NULL pointers
    uint32_t count = 0;
    if (p1) count++;
    if (p2) count++;
    if (p3) count++;
    if (p4) count++;
    if (p5) count++;
    if (p6) count++;
    if (p7) count++;
    if (p8) count++;
    if (p9) count++;
    DEBUG_PRINT("test_9_ptrs: %u non-NULL pointers", count);
    return count;
}

// ============================================================================
// FLOAT TESTS
// ============================================================================

float test_11_float32(float a1, float a2, float a3, float a4, float a5,
                      float a6, float a7, float a8, float a9, float a10, float a11) {
    float sum = a1 + a2 + a3 + a4 + a5 + a6 + a7 + a8 + a9 + a10 + a11;
    DEBUG_PRINT("test_11_float32: sum = %f (expected %f)", sum, 66.0f);
    return sum;
}

double test_11_float64(double a1, double a2, double a3, double a4, double a5,
                       double a6, double a7, double a8, double a9, double a10, double a11) {
    double sum = a1 + a2 + a3 + a4 + a5 + a6 + a7 + a8 + a9 + a10 + a11;
    DEBUG_PRINT("test_11_float64: sum = %f (expected %f)", sum, 66.0);
    return sum;
}

// Mixed int and float (int registers + float on stack)
float test_mixed_8i32_3f32(int32_t i1, int32_t i2, int32_t i3, int32_t i4,
                            int32_t i5, int32_t i6, int32_t i7, int32_t i8,
                            float f1, float f2, float f3) {
    float result = (float)(i1 + i2 + i3 + i4 + i5 + i6 + i7 + i8) + f1 + f2 + f3;
    DEBUG_PRINT("test_mixed_8i32_3f32: result = %f (expected %f)", result, 42.0f);
    return result;
}

// ============================================================================
// COMPLEX MIXED: Multiple different types on stack
// ============================================================================

int32_t test_mixed_kitchen_sink(int32_t i1, int32_t i2, int32_t i3, int32_t i4,
                                 int32_t i5, int32_t i6, int32_t i7, int32_t i8,
                                 int8_t s1, int16_t s2, int32_t s3, bool b1, void* p1) {
    // 8 regs (i1-i8) + 5 stack args of different types
    int32_t sum = i1 + i2 + i3 + i4 + i5 + i6 + i7 + i8 + s1 + s2 + s3 + (b1 ? 1 : 0) + (p1 ? 1 : 0);
    DEBUG_PRINT("test_mixed_kitchen_sink: sum = %d (expected %d)", sum, 43);
    return sum;
}

// ============================================================================
// ALTERNATING TYPES: Tests alignment across different boundaries
// ============================================================================

int32_t test_alternating_i32_bool(int32_t i1, bool b1, int32_t i2, bool b2, int32_t i3,
                                   bool b3, int32_t i4, bool b4, int32_t i5, bool b5, int32_t i6) {
    // Tests alternating pattern - should hit stack after 8 args total
    int32_t sum = i1 + i2 + i3 + i4 + i5 + i6 + (b1 ? 1 : 0) + (b2 ? 1 : 0) + (b3 ? 1 : 0) + (b4 ? 1 : 0) + (b5 ? 1 : 0);
    DEBUG_PRINT("test_alternating_i32_bool: sum = %d (expected %d)", sum, 26);
    return sum;
}

// ============================================================================
// REGRESSION TEST: Original failing case (mlx_conv2d signature)
// ============================================================================

int32_t test_conv2d_signature(void* res, void* input, void* weight,
                               int32_t stride_0, int32_t stride_1,
                               int32_t padding_0, int32_t padding_1,
                               int32_t dilation_0, int32_t dilation_1,
                               int32_t groups, void* stream) {
    // Simulate mlx_conv2d parameter pattern: 3 ptrs + 7 int32 + 1 ptr
    int32_t sum = stride_0 + stride_1 + padding_0 + padding_1 +
                  dilation_0 + dilation_1 + groups;
    DEBUG_PRINT("test_conv2d_signature: sum = %d (expected %d)", sum, 7);
    return sum;
}
