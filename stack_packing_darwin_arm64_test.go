// Copyright 2025 The Ebitengine Authors
// SPDX-License-Identifier: Apache-2.0

//go:build darwin || freebsd || (linux && (amd64 || arm64 || loong64)) || netbsd

package purego_test

import (
	"math"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"unsafe"

	"github.com/ebitengine/purego"
	"github.com/ebitengine/purego/internal/load"
)

// loadTestLib loads the comprehensive stack test library
func loadTestLib(t *testing.T) uintptr {
	t.Helper()

	libFileName := "libcomprehensive_stack_test.so"
	t.Logf("Build %v", libFileName)

	if err := buildSharedLib("CC", libFileName, filepath.Join("testdata", "stacktest", "comprehensive_stack_test.c")); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		os.Remove(libFileName)
	})

	lib, err := load.OpenLibrary(libFileName)
	if err != nil {
		t.Fatalf("Failed to load library: %v", err)
	}
	t.Cleanup(func() {
		if err := load.CloseLibrary(lib); err != nil {
			t.Errorf("Failed to close library: %v", err)
		}
	})
	return lib
}

// testResult helps check expected vs actual with detailed output.
// Some tests are known to fail on Darwin ARM64 due to int32/float32 stack packing bugs.
// On other platforms, all tests should pass.
func testResult(t *testing.T, name string, got, want interface{}, expectedFailOnDarwinARM64 bool) {
	t.Helper()
	isDarwinARM64 := runtime.GOOS == "darwin" && runtime.GOARCH == "arm64"

	if got != want {
		if isDarwinARM64 && expectedFailOnDarwinARM64 {
			t.Logf("%s: got %v, want %v (KNOWN BUG on Darwin ARM64)", name, got, want)
		} else {
			t.Errorf("%s: got %v, want %v ❌ FAIL", name, got, want)
		}
	} else {
		if isDarwinARM64 && expectedFailOnDarwinARM64 {
			t.Logf("%s: got %v ✓ PASS (bug appears to be fixed!)", name, got)
		} else {
			t.Logf("%s: got %v ✓ PASS", name, got)
		}
	}
}

// testFloatResult checks float results with tolerance
func testFloatResult(t *testing.T, name string, got, want float64, expectedFailOnDarwinARM64 bool) {
	t.Helper()
	isDarwinARM64 := runtime.GOOS == "darwin" && runtime.GOARCH == "arm64"

	matches := math.Abs(got-want) < 0.01
	if !matches {
		if isDarwinARM64 && expectedFailOnDarwinARM64 {
			t.Logf("%s: got %v, want %v (KNOWN BUG on Darwin ARM64)", name, got, want)
		} else {
			t.Errorf("%s: got %v, want %v ❌ FAIL", name, got, want)
		}
	} else {
		if isDarwinARM64 && expectedFailOnDarwinARM64 {
			t.Logf("%s: got %v ✓ PASS (bug appears to be fixed!)", name, got)
		} else {
			t.Logf("%s: got %v ✓ PASS", name, got)
		}
	}
}

// ============================================================================
// UNIFORM TYPE TESTS - RegisterFunc
// ============================================================================

func TestDarwin_ARM64_RegisterFunc_11_int8(t *testing.T) {
	lib := loadTestLib(t)
	var fn func(int8, int8, int8, int8, int8, int8, int8, int8, int8, int8, int8) int8
	purego.RegisterLibFunc(&fn, lib, "test_11_int8")

	result := fn(1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11)
	testResult(t, "test_11_int8", result, int8(66), false)
}

func TestDarwin_ARM64_RegisterFunc_11_int16(t *testing.T) {
	lib := loadTestLib(t)
	var fn func(int16, int16, int16, int16, int16, int16, int16, int16, int16, int16, int16) int16
	purego.RegisterLibFunc(&fn, lib, "test_11_int16")

	result := fn(1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11)
	testResult(t, "test_11_int16", result, int16(66), false)
}

func TestDarwin_ARM64_RegisterFunc_11_int32(t *testing.T) {
	lib := loadTestLib(t)
	var fn func(int32, int32, int32, int32, int32, int32, int32, int32, int32, int32, int32) int32
	purego.RegisterLibFunc(&fn, lib, "test_11_int32")

	result := fn(1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11)
	testResult(t, "test_11_int32", result, int32(66), true)
}

func TestDarwin_ARM64_RegisterFunc_11_int64(t *testing.T) {
	lib := loadTestLib(t)
	var fn func(int64, int64, int64, int64, int64, int64, int64, int64, int64, int64, int64) int64
	purego.RegisterLibFunc(&fn, lib, "test_11_int64")

	result := fn(1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11)
	testResult(t, "test_11_int64", result, int64(66), false)
}

// ============================================================================
// EDGE CASES - RegisterFunc
// ============================================================================

func TestDarwin_ARM64_RegisterFunc_9_int32(t *testing.T) {
	lib := loadTestLib(t)
	var fn func(int32, int32, int32, int32, int32, int32, int32, int32, int32) int32
	purego.RegisterLibFunc(&fn, lib, "test_9_int32")

	result := fn(1, 2, 3, 4, 5, 6, 7, 8, 9)
	testResult(t, "test_9_int32", result, int32(45), false)
}

func TestDarwin_ARM64_RegisterFunc_15_int32(t *testing.T) {
	lib := loadTestLib(t)
	var fn func(int32, int32, int32, int32, int32, int32, int32, int32, int32, int32, int32, int32, int32, int32, int32) int32
	purego.RegisterLibFunc(&fn, lib, "test_15_int32")

	result := fn(1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15)
	testResult(t, "test_15_int32", result, int32(120), true)
}

// ============================================================================
// MIXED TYPES - RegisterFunc
// ============================================================================

func TestDarwin_ARM64_RegisterFunc_Mixed_8r_2u8_1u32(t *testing.T) {
	lib := loadTestLib(t)
	var fn func(uint32, uint32, uint32, uint32, uint32, uint32, uint32, uint32, uint8, uint8, uint32) uint32
	purego.RegisterLibFunc(&fn, lib, "test_mixed_8r_2u8_1u32")

	result := fn(256, 512, 4, 8, 16, 32, 64, 128, 1, 2, 1024)
	testResult(t, "test_mixed_8r_2u8_1u32", result, uint32(2047), false)
}

func TestDarwin_ARM64_RegisterFunc_Mixed_8i32_3i16(t *testing.T) {
	lib := loadTestLib(t)
	var fn func(int32, int32, int32, int32, int32, int32, int32, int32, int16, int16, int16) int32
	purego.RegisterLibFunc(&fn, lib, "test_mixed_8i32_3i16")

	result := fn(1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11)
	testResult(t, "test_mixed_8i32_3i16", result, int32(42), true)
}

func TestDarwin_ARM64_RegisterFunc_Mixed_Varied(t *testing.T) {
	lib := loadTestLib(t)
	var fn func(int32, int32, int32, int32, int32, int32, int32, int32, int8, int16, int32) int32
	purego.RegisterLibFunc(&fn, lib, "test_mixed_varied")

	result := fn(1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11)
	testResult(t, "test_mixed_varied", result, int32(42), true)
}

// ============================================================================
// BOOL TESTS - RegisterFunc
// ============================================================================

func TestDarwin_ARM64_RegisterFunc_11_bool(t *testing.T) {
	lib := loadTestLib(t)
	var fn func(bool, bool, bool, bool, bool, bool, bool, bool, bool, bool, bool) int32
	purego.RegisterLibFunc(&fn, lib, "test_11_bool")

	result := fn(true, false, true, false, true, false, true, false, true, false, true)
	testResult(t, "test_11_bool", result, int32(6), false)
}

func TestDarwin_ARM64_RegisterFunc_Mixed_8i32_3bool(t *testing.T) {
	lib := loadTestLib(t)
	var fn func(int32, int32, int32, int32, int32, int32, int32, int32, bool, bool, bool) int32
	purego.RegisterLibFunc(&fn, lib, "test_mixed_8i32_3bool")

	result := fn(1, 2, 3, 4, 5, 6, 7, 8, true, true, true)
	testResult(t, "test_mixed_8i32_3bool", result, int32(39), false)
}

// ============================================================================
// POINTER TESTS - RegisterFunc
// ============================================================================

func TestDarwin_ARM64_RegisterFunc_9_ptrs(t *testing.T) {
	lib := loadTestLib(t)
	var fn func(uintptr, uintptr, uintptr, uintptr, uintptr,
		uintptr, uintptr, uintptr, uintptr) uint32
	purego.RegisterLibFunc(&fn, lib, "test_9_ptrs")

	dummy := 42
	p := uintptr(unsafe.Pointer(&dummy))
	result := fn(p, p, p, p, p, p, p, p, p)
	testResult(t, "test_9_ptrs", result, uint32(9), false)
}

// ============================================================================
// FLOAT TESTS - RegisterFunc
// ============================================================================

func TestDarwin_ARM64_RegisterFunc_11_float32(t *testing.T) {
	lib := loadTestLib(t)
	var fn func(float32, float32, float32, float32, float32, float32, float32, float32, float32, float32, float32) float32
	purego.RegisterLibFunc(&fn, lib, "test_11_float32")

	result := fn(1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11)
	testFloatResult(t, "test_11_float32", float64(result), 66.0, true)
}

func TestDarwin_ARM64_RegisterFunc_11_float64(t *testing.T) {
	lib := loadTestLib(t)
	var fn func(float64, float64, float64, float64, float64, float64, float64, float64, float64, float64, float64) float64
	purego.RegisterLibFunc(&fn, lib, "test_11_float64")

	result := fn(1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11)
	testFloatResult(t, "test_11_float64", result, 66.0, false)
}

func TestDarwin_ARM64_RegisterFunc_Mixed_8i32_3f32(t *testing.T) {
	lib := loadTestLib(t)
	var fn func(int32, int32, int32, int32, int32, int32, int32, int32, float32, float32, float32) float32
	purego.RegisterLibFunc(&fn, lib, "test_mixed_8i32_3f32")

	result := fn(1, 2, 3, 4, 5, 6, 7, 8, 9.0, 10.0, 11.0)
	testFloatResult(t, "test_mixed_8i32_3f32", float64(result), 42.0, true)
}

// ============================================================================
// COMPLEX MIXED - RegisterFunc
// ============================================================================

func TestDarwin_ARM64_RegisterFunc_KitchenSink(t *testing.T) {
	lib := loadTestLib(t)
	var fn func(int32, int32, int32, int32, int32, int32, int32, int32, int8, int16, int32, bool, uintptr) int32
	purego.RegisterLibFunc(&fn, lib, "test_mixed_kitchen_sink")

	dummy := 42
	result := fn(1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, true, uintptr(unsafe.Pointer(&dummy)))
	testResult(t, "test_mixed_kitchen_sink", result, int32(43), true)
}

func TestDarwin_ARM64_RegisterFunc_Alternating(t *testing.T) {
	lib := loadTestLib(t)
	var fn func(int32, bool, int32, bool, int32, bool, int32, bool, int32, bool, int32) int32
	purego.RegisterLibFunc(&fn, lib, "test_alternating_i32_bool")

	result := fn(1, true, 2, true, 3, true, 4, true, 5, true, 6)
	testResult(t, "test_alternating_i32_bool", result, int32(26), true)
}

// ============================================================================
// REGRESSION TEST
// ============================================================================

func TestDarwin_ARM64_RegisterFunc_Conv2D_Signature(t *testing.T) {
	lib := loadTestLib(t)
	var fn func(uintptr, uintptr, uintptr, int32, int32, int32, int32, int32, int32, int32, uintptr) int32
	purego.RegisterLibFunc(&fn, lib, "test_conv2d_signature")

	dummy := 42
	p := uintptr(unsafe.Pointer(&dummy))
	result := fn(p, p, p, 1, 1, 0, 0, 1, 1, 1, p)
	testResult(t, "test_conv2d_signature", result, int32(7), true)
}

// ============================================================================
// SYSCALLN TESTS - Key scenarios
// ============================================================================

func TestDarwin_ARM64_SyscallN_11_int32(t *testing.T) {
	lib := loadTestLib(t)
	fnPtr, err := load.OpenSymbol(lib, "test_11_int32")
	if err != nil {
		t.Fatal(err)
	}

	result, _, _ := purego.SyscallN(fnPtr, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11)
	testResult(t, "SyscallN_test_11_int32", int32(result), int32(66), true)
}

func TestDarwin_ARM64_SyscallN_11_int64(t *testing.T) {
	lib := loadTestLib(t)
	fnPtr, err := load.OpenSymbol(lib, "test_11_int64")
	if err != nil {
		t.Fatal(err)
	}

	result, _, _ := purego.SyscallN(fnPtr, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11)
	testResult(t, "SyscallN_test_11_int64", int64(result), int64(66), false)
}

func TestDarwin_ARM64_SyscallN_9_ptrs(t *testing.T) {
	lib := loadTestLib(t)
	fnPtr, err := load.OpenSymbol(lib, "test_9_ptrs")
	if err != nil {
		t.Fatal(err)
	}

	dummy := 42
	p := uintptr(unsafe.Pointer(&dummy))
	result, _, _ := purego.SyscallN(fnPtr, p, p, p, p, p, p, p, p, p)
	testResult(t, "SyscallN_test_9_ptrs", uint32(result), uint32(9), false)
}
