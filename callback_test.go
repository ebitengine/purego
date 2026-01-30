// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2023 The Ebitengine Authors

//go:build darwin || (linux && (386 || amd64 || arm || arm64 || loong64 || riscv64))

package purego_test

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"unsafe"

	"github.com/ebitengine/purego"
)

// TestCallGoFromSharedLib is a test that checks for stack corruption on arm64
// when C calls Go code from a non-Go thread in a dynamically loaded share library.
func TestCallGoFromSharedLib(t *testing.T) {
	libFileName := filepath.Join(t.TempDir(), "libcbtest.so")
	t.Logf("Build %v", libFileName)

	if err := buildSharedLib("CC", libFileName, filepath.Join("testdata", "libcbtest", "callback_test.c")); err != nil {
		t.Fatal(err)
	}
	defer os.Remove(libFileName)

	lib, err := purego.Dlopen(libFileName, purego.RTLD_NOW|purego.RTLD_GLOBAL)
	if err != nil {
		t.Fatalf("Dlopen(%q) failed: %v", libFileName, err)
	}

	var callCallback func(p uintptr, s string) int
	purego.RegisterLibFunc(&callCallback, lib, "callCallback")

	goFunc := func(cstr *byte, n int) int {
		s := string(unsafe.Slice(cstr, n))
		t.Logf("FROM Go: %s\n", s)
		return 1
	}

	const want = 10101
	cb := purego.NewCallback(goFunc)
	for i := 0; i < 10; i++ {
		got := callCallback(cb, "a test string")
		if got != want {
			t.Fatalf("%d: callCallback() got %v want %v", i, got, want)
		}
	}
}

func TestNewCallbackFloat64(t *testing.T) {
	// This tests the maximum number of arguments a function to NewCallback can take
	const (
		expectCbTotal    = -3
		expectedCbTotalF = float64(36)
	)
	var cbTotal int
	var cbTotalF float64
	imp := purego.NewCallback(func(a1, a2, a3, a4, a5, a6, a7, a8, a9 int,
		f1, f2, f3, f4, f5, f6, f7, f8 float64,
	) {
		cbTotal = a1 + a2 + a3 + a4 + a5 + a6 + a7 + a8 + a9
		cbTotalF = f1 + f2 + f3 + f4 + f5 + f6 + f7 + f8
	})
	var fn func(a1, a2, a3, a4, a5, a6, a7, a8, a9 int,
		f1, f2, f3, f4, f5, f6, f7, f8 float64)
	purego.RegisterFunc(&fn, imp)
	fn(1, 2, -3, 4, -5, 6, -7, 8, -9,
		1, 2, 3, 4, 5, 6, 7, 8)

	if cbTotal != expectCbTotal {
		t.Errorf("cbTotal not correct got %d but wanted %d", cbTotal, expectCbTotal)
	}
	if cbTotalF != expectedCbTotalF {
		t.Errorf("cbTotalF not correct got %f but wanted %f", cbTotalF, expectedCbTotalF)
	}
}

func TestNewCallbackFloat64AndIntMix(t *testing.T) {
	// This tests interleaving float and integer arguments to NewCallback
	const (
		expectCbTotal = 54.75
	)
	var cbTotal float64
	imp := purego.NewCallback(func(a1, a2 float64, a3, a4, a5 int, a6, a7, a8 float64, a9 int) {
		cbTotal = a1 + a2 + float64(a3) + float64(a4) + float64(a5) + a6 + a7 + a8 + float64(a9)
	})
	var fn func(a1, a2 float64, a3, a4, a5 int, a6, a7, a8 float64, a9 int)
	purego.RegisterFunc(&fn, imp)
	fn(1.25, 3.25, 4, 5, 6, 7.5, 8.25, 9.5, 10)

	if cbTotal != expectCbTotal {
		t.Errorf("cbTotal not correct got %f but wanted %f", cbTotal, expectCbTotal)
	}
}

func TestNewCallbackFloat32(t *testing.T) {
	// This tests the maximum number of float32 arguments a function to NewCallback can take
	const (
		expectCbTotal    = 6
		expectedCbTotalF = float32(45)
	)
	var cbTotal int
	var cbTotalF float32
	imp := purego.NewCallback(func(a1, a2, a3, a4, a5, a6, a7, a8 int,
		f1, f2, f3, f4, f5, f6, f7, f8, f9 float32,
	) {
		cbTotal = a1 + a2 + a3 + a4 + a5 + a6 + a7 + a8
		cbTotalF = f1 + f2 + f3 + f4 + f5 + f6 + f7 + f8 + f9
	})
	var fn func(a1, a2, a3, a4, a5, a6, a7, a8 int,
		f1, f2, f3, f4, f5, f6, f7, f8, f9 float32)
	purego.RegisterFunc(&fn, imp)
	fn(1, 2, -3, 4, -5, 6, -7, 8,
		1, 2, 3, 4, 5, 6, 7, 8, 9)

	if cbTotal != expectCbTotal {
		t.Errorf("cbTotal not correct got %d but wanted %d", cbTotal, expectCbTotal)
	}
	if cbTotalF != expectedCbTotalF {
		t.Errorf("cbTotalF not correct got %f but wanted %f", cbTotalF, expectedCbTotalF)
	}
}

func TestNewCallbackFloat32AndFloat64(t *testing.T) {
	// This tests that calling a function with a mix of float32 and float64 arguments works
	const (
		expectedCbTotalF32 = float32(72)
		expectedCbTotalF64 = float64(48)
	)
	var cbTotalF32 float32
	var cbTotalF64 float64
	imp := purego.NewCallback(func(f1, f2, f3 float32, f4, f5, f6 float64, f7, f8, f9 float32, f10, f11, f12 float64, f13, f14, f15 float32) {
		cbTotalF32 = f1 + f2 + f3 + f7 + f8 + f9 + f13 + f14 + f15
		cbTotalF64 = f4 + f5 + f6 + f10 + f11 + f12
	})
	var fn func(f1, f2, f3 float32, f4, f5, f6 float64, f7, f8, f9 float32, f10, f11, f12 float64, f13, f14, f15 float32)
	purego.RegisterFunc(&fn, imp)
	fn(1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15)

	if cbTotalF32 != expectedCbTotalF32 {
		t.Errorf("cbTotalF32 not correct got %f but wanted %f", cbTotalF32, expectedCbTotalF32)
	}
	if cbTotalF64 != expectedCbTotalF64 {
		t.Errorf("cbTotalF64 not correct got %f but wanted %f", cbTotalF64, expectedCbTotalF64)
	}
}

func ExampleNewCallback() {
	cb := purego.NewCallback(func(a1, a2, a3, a4, a5, a6, a7, a8, a9, a10, a11, a12, a13, a14, a15 int) int {
		fmt.Println(a1, a2, a3, a4, a5, a6, a7, a8, a9, a10, a11, a12, a13, a14, a15)
		return a1 + a2 + a3 + a4 + a5 + a6 + a7 + a8 + a9 + a10 + a11 + a12 + a13 + a14 + a15
	})

	var fn func(a1, a2, a3, a4, a5, a6, a7, a8, a9, a10, a11, a12, a13, a14, a15 int) int
	purego.RegisterFunc(&fn, cb)

	ret := fn(1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15)
	fmt.Println(ret)

	// Output: 1 2 3 4 5 6 7 8 9 10 11 12 13 14 15
	// 120
}

func ExampleNewCallback_cdecl() {
	fn := func(_ purego.CDecl, a int) {
		fmt.Println(a)
	}
	cb := purego.NewCallback(fn)
	purego.SyscallN(cb, 83)

	// Output: 83
}

func TestCallbackInt32Packing(t *testing.T) {
	if runtime.GOOS != "darwin" || runtime.GOARCH != "arm64" {
		t.Skip("callback tight packing only applies to darwin/arm64")
	}

	libFileName := filepath.Join(t.TempDir(), "libcbtest_packing.so")
	if err := buildSharedLib("CC", libFileName, filepath.Join("testdata", "libcbtest", "callback_packing_test.c")); err != nil {
		t.Fatal(err)
	}
	defer os.Remove(libFileName)

	lib, err := purego.Dlopen(libFileName, purego.RTLD_NOW|purego.RTLD_GLOBAL)
	if err != nil {
		t.Fatalf("Dlopen(%q) failed: %v", libFileName, err)
	}

	var callCallback12Int32 func(cb uintptr) int32
	purego.RegisterLibFunc(&callCallback12Int32, lib, "callCallback12Int32")

	// Go callback that sums the 12 int32 arguments (prime numbers: 2,3,5,7,11,13,17,19,23,29,31,37)
	goFunc := func(a1, a2, a3, a4, a5, a6, a7, a8, a9, a10, a11, a12 int32) int32 {
		return a1 + a2 + a3 + a4 + a5 + a6 + a7 + a8 + a9 + a10 + a11 + a12
	}

	cb := purego.NewCallback(goFunc)
	got := callCallback12Int32(cb)
	want := int32(2 + 3 + 5 + 7 + 11 + 13 + 17 + 19 + 23 + 29 + 31 + 37) // 197
	if got != want {
		t.Errorf("callCallback12Int32() = %d, want %d", got, want)
	}
}

func TestCallbackMixedStackPacking(t *testing.T) {
	if runtime.GOOS != "darwin" || runtime.GOARCH != "arm64" {
		t.Skip("callback tight packing only applies to darwin/arm64")
	}

	libFileName := filepath.Join(t.TempDir(), "libcbtest_packing.so")
	if err := buildSharedLib("CC", libFileName, filepath.Join("testdata", "libcbtest", "callback_packing_test.c")); err != nil {
		t.Fatal(err)
	}
	defer os.Remove(libFileName)

	lib, err := purego.Dlopen(libFileName, purego.RTLD_NOW|purego.RTLD_GLOBAL)
	if err != nil {
		t.Fatalf("Dlopen(%q) failed: %v", libFileName, err)
	}

	var callCallbackMixedStack func(cb uintptr) int64
	purego.RegisterLibFunc(&callCallbackMixedStack, lib, "callCallbackMixedStack")

	// Go callback: 8 int64s in regs, then int32(100), int64(200), int32(300) on stack
	goFunc := func(a1, a2, a3, a4, a5, a6, a7, a8 int64, s1 int32, s2 int64, s3 int32) int64 {
		// Return sum to verify all args received correctly
		return a1 + a2 + a3 + a4 + a5 + a6 + a7 + a8 + int64(s1) + s2 + int64(s3)
	}

	cb := purego.NewCallback(goFunc)
	got := callCallbackMixedStack(cb)
	want := int64(1 + 2 + 3 + 4 + 5 + 6 + 7 + 8 + 100 + 200 + 300) // 636
	if got != want {
		t.Errorf("callCallbackMixedStack() = %d, want %d", got, want)
	}
}

func TestCallbackSmallTypesPacking(t *testing.T) {
	if runtime.GOOS != "darwin" || runtime.GOARCH != "arm64" {
		t.Skip("callback tight packing only applies to darwin/arm64")
	}

	libFileName := filepath.Join(t.TempDir(), "libcbtest_packing.so")
	if err := buildSharedLib("CC", libFileName, filepath.Join("testdata", "libcbtest", "callback_packing_test.c")); err != nil {
		t.Fatal(err)
	}
	defer os.Remove(libFileName)

	lib, err := purego.Dlopen(libFileName, purego.RTLD_NOW|purego.RTLD_GLOBAL)
	if err != nil {
		t.Fatalf("Dlopen(%q) failed: %v", libFileName, err)
	}

	var callCallbackSmallTypes func(cb uintptr) int64
	purego.RegisterLibFunc(&callCallbackSmallTypes, lib, "callCallbackSmallTypes")

	// Values: 1,2,3,4,5,6,7,8 (regs), true, -42, 200, -1000, 50000, 123456 (stack)
	var gotBool bool
	var gotI8 int8
	var gotU8 uint8
	var gotI16 int16
	var gotU16 uint16
	var gotI32 int32

	goFunc := func(a1, a2, a3, a4, a5, a6, a7, a8 int64,
		b bool, i8 int8, u8 uint8, i16 int16, u16 uint16, i32 int32) int64 {
		gotBool = b
		gotI8 = i8
		gotU8 = u8
		gotI16 = i16
		gotU16 = u16
		gotI32 = i32
		return a1 + a2 + a3 + a4 + a5 + a6 + a7 + a8
	}

	cb := purego.NewCallback(goFunc)
	got := callCallbackSmallTypes(cb)

	// Check register args sum
	want := int64(1 + 2 + 3 + 4 + 5 + 6 + 7 + 8) // 36
	if got != want {
		t.Errorf("register args sum = %d, want %d", got, want)
	}

	// Check stack args
	if gotBool != true {
		t.Errorf("bool = %v, want true", gotBool)
	}
	if gotI8 != -42 {
		t.Errorf("int8 = %d, want -42", gotI8)
	}
	if gotU8 != 200 {
		t.Errorf("uint8 = %d, want 200", gotU8)
	}
	if gotI16 != -1000 {
		t.Errorf("int16 = %d, want -1000", gotI16)
	}
	if gotU16 != 50000 {
		t.Errorf("uint16 = %d, want 50000", gotU16)
	}
	if gotI32 != 123456 {
		t.Errorf("int32 = %d, want 123456", gotI32)
	}
}

func TestCallback10Int32Packing(t *testing.T) {
	if runtime.GOOS != "darwin" || runtime.GOARCH != "arm64" {
		t.Skip("callback tight packing only applies to darwin/arm64")
	}

	libFileName := filepath.Join(t.TempDir(), "libcbtest_packing.so")
	if err := buildSharedLib("CC", libFileName, filepath.Join("testdata", "libcbtest", "callback_packing_test.c")); err != nil {
		t.Fatal(err)
	}
	defer os.Remove(libFileName)

	lib, err := purego.Dlopen(libFileName, purego.RTLD_NOW|purego.RTLD_GLOBAL)
	if err != nil {
		t.Fatalf("Dlopen(%q) failed: %v", libFileName, err)
	}

	var callCallback10Int32 func(cb uintptr) int32
	purego.RegisterLibFunc(&callCallback10Int32, lib, "callCallback10Int32")

	goFunc := func(a1, a2, a3, a4, a5, a6, a7, a8, a9, a10 int32) int32 {
		return a1 + a2 + a3 + a4 + a5 + a6 + a7 + a8 + a9 + a10
	}

	cb := purego.NewCallback(goFunc)
	got := callCallback10Int32(cb)
	want := int32(1 + 2 + 3 + 4 + 5 + 6 + 7 + 8 + 9 + 10) // 55
	if got != want {
		t.Errorf("callCallback10Int32() = %d, want %d", got, want)
	}
}

func TestCallbackFloat64StackPacking(t *testing.T) {
	if runtime.GOOS != "darwin" || runtime.GOARCH != "arm64" {
		t.Skip("callback tight packing only applies to darwin/arm64")
	}

	libFileName := filepath.Join(t.TempDir(), "libcbtest_packing.so")
	if err := buildSharedLib("CC", libFileName, filepath.Join("testdata", "libcbtest", "callback_packing_test.c")); err != nil {
		t.Fatal(err)
	}
	defer os.Remove(libFileName)

	lib, err := purego.Dlopen(libFileName, purego.RTLD_NOW|purego.RTLD_GLOBAL)
	if err != nil {
		t.Fatalf("Dlopen(%q) failed: %v", libFileName, err)
	}

	var callCallback10Float64 func(cb uintptr) int64
	purego.RegisterLibFunc(&callCallback10Float64, lib, "callCallback10Float64")

	// 10 float64s: 8 in registers, 2 on stack
	// Return int64 since callbacks don't support float returns
	goFunc := func(f1, f2, f3, f4, f5, f6, f7, f8, f9, f10 float64) int64 {
		sum := f1 + f2 + f3 + f4 + f5 + f6 + f7 + f8 + f9 + f10
		return int64(sum * 10) // 60.0 * 10 = 600
	}

	cb := purego.NewCallback(goFunc)
	got := callCallback10Float64(cb)
	want := int64(600)
	if got != want {
		t.Errorf("callCallback10Float64() = %d, want %d", got, want)
	}
}

func TestCallbackFloat32StackPacking(t *testing.T) {
	if runtime.GOOS != "darwin" || runtime.GOARCH != "arm64" {
		t.Skip("callback tight packing only applies to darwin/arm64")
	}

	libFileName := filepath.Join(t.TempDir(), "libcbtest_packing.so")
	if err := buildSharedLib("CC", libFileName, filepath.Join("testdata", "libcbtest", "callback_packing_test.c")); err != nil {
		t.Fatal(err)
	}
	defer os.Remove(libFileName)

	lib, err := purego.Dlopen(libFileName, purego.RTLD_NOW|purego.RTLD_GLOBAL)
	if err != nil {
		t.Fatalf("Dlopen(%q) failed: %v", libFileName, err)
	}

	var callCallback12Float32 func(cb uintptr) int64
	purego.RegisterLibFunc(&callCallback12Float32, lib, "callCallback12Float32")

	// 12 float32s: 8 in registers, 4 on stack
	// Return int64 since callbacks don't support float returns
	goFunc := func(f1, f2, f3, f4, f5, f6, f7, f8, f9, f10, f11, f12 float32) int64 {
		sum := f1 + f2 + f3 + f4 + f5 + f6 + f7 + f8 + f9 + f10 + f11 + f12
		return int64(sum) // 78
	}

	cb := purego.NewCallback(goFunc)
	got := callCallback12Float32(cb)
	want := int64(78)
	if got != want {
		t.Errorf("callCallback12Float32() = %d, want %d", got, want)
	}
}
