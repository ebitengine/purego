// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2023 The Ebitengine Authors

//go:build darwin || (linux && (amd64 || arm64 || loong64))

package purego_test

import (
	"fmt"
	"os"
	"path/filepath"
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

func TestNewCallbackInt32Packing(t *testing.T) {
	var result int32
	cb := purego.NewCallback(func(a1, a2, a3, a4, a5, a6, a7, a8, a9, a10, a11, a12 int32) int32 {
		result = a1 + a2 + a3 + a4 + a5 + a6 + a7 + a8 + a9 + a10 + a11 + a12
		return result
	})

	var fn func(a1, a2, a3, a4, a5, a6, a7, a8, a9, a10, a11, a12 int32) int32
	purego.RegisterFunc(&fn, cb)

	got := fn(2, 3, 5, 7, 11, 13, 17, 19, 23, 29, 31, 37)
	want := int32(197)

	if got != want {
		t.Errorf("callback returned %d, want %d", got, want)
	}
}

func TestNewCallbackMixedPacking(t *testing.T) {
	var gotI32_1, gotI32_2 int32
	var gotI64 int64
	cb := purego.NewCallback(func(r1, r2, r3, r4, r5, r6, r7, r8 int64, s1 int32, s2 int64, s3 int32) {
		gotI32_1 = s1
		gotI64 = s2
		gotI32_2 = s3
	})

	var fn func(r1, r2, r3, r4, r5, r6, r7, r8 int64, s1 int32, s2 int64, s3 int32)
	purego.RegisterFunc(&fn, cb)

	fn(1, 2, 3, 4, 5, 6, 7, 8, 100, 200, 300)

	if gotI32_1 != 100 || gotI64 != 200 || gotI32_2 != 300 {
		t.Errorf("got (%d, %d, %d), want (100, 200, 300)", gotI32_1, gotI64, gotI32_2)
	}
}

func TestNewCallbackSmallTypes(t *testing.T) {
	var gotBool bool
	var gotI8 int8
	var gotU8 uint8
	var gotI16 int16
	var gotU16 uint16
	var gotI32 int32
	cb := purego.NewCallback(func(r1, r2, r3, r4, r5, r6, r7, r8 int64, b bool, i8 int8, u8 uint8, i16 int16, u16 uint16, i32 int32) {
		gotBool = b
		gotI8 = i8
		gotU8 = u8
		gotI16 = i16
		gotU16 = u16
		gotI32 = i32
	})

	var fn func(r1, r2, r3, r4, r5, r6, r7, r8 int64, b bool, i8 int8, u8 uint8, i16 int16, u16 uint16, i32 int32)
	purego.RegisterFunc(&fn, cb)

	fn(1, 2, 3, 4, 5, 6, 7, 8, true, -42, 200, -1000, 50000, 123456)

	if !gotBool || gotI8 != -42 || gotU8 != 200 || gotI16 != -1000 || gotU16 != 50000 || gotI32 != 123456 {
		t.Errorf("got (bool=%v, i8=%d, u8=%d, i16=%d, u16=%d, i32=%d), want (true, -42, 200, -1000, 50000, 123456)",
			gotBool, gotI8, gotU8, gotI16, gotU16, gotI32)
	}
}

func TestCallbackFromC(t *testing.T) {
	libFileName := filepath.Join(t.TempDir(), "libcbpackingtest.so")

	if err := buildSharedLib("CC", libFileName, filepath.Join("testdata", "libcbtest", "callback_packing_test.c")); err != nil {
		t.Fatal(err)
	}
	defer os.Remove(libFileName)

	lib, err := purego.Dlopen(libFileName, purego.RTLD_NOW|purego.RTLD_GLOBAL)
	if err != nil {
		t.Fatalf("Dlopen(%q) failed: %v", libFileName, err)
	}

	t.Run("int32_packing", func(t *testing.T) {
		var result int32
		goCallback := func(a1, a2, a3, a4, a5, a6, a7, a8, a9, a10, a11, a12 int32) int32 {
			result = a1 + a2 + a3 + a4 + a5 + a6 + a7 + a8 + a9 + a10 + a11 + a12
			return result
		}

		var callCallbackInt32Packing func(uintptr) int32
		purego.RegisterLibFunc(&callCallbackInt32Packing, lib, "callCallbackInt32Packing")

		cb := purego.NewCallback(goCallback)
		got := callCallbackInt32Packing(cb)
		want := int32(197) // sum of primes: 2+3+5+7+11+13+17+19+23+29+31+37

		if got != want {
			t.Errorf("C called callback returned %d, want %d", got, want)
		}
		if result != want {
			t.Errorf("callback received wrong args, sum=%d, want %d", result, want)
		}
	})

	t.Run("mixed_packing", func(t *testing.T) {
		var gotI32_1, gotI32_2 int32
		var gotI64 int64
		goCallback := func(r1, r2, r3, r4, r5, r6, r7, r8 int64, s1 int32, s2 int64, s3 int32) {
			gotI32_1 = s1
			gotI64 = s2
			gotI32_2 = s3
		}

		var callCallbackMixedPacking func(uintptr)
		purego.RegisterLibFunc(&callCallbackMixedPacking, lib, "callCallbackMixedPacking")

		cb := purego.NewCallback(goCallback)
		callCallbackMixedPacking(cb)

		if gotI32_1 != 100 || gotI64 != 200 || gotI32_2 != 300 {
			t.Errorf("callback received (%d, %d, %d), want (100, 200, 300)", gotI32_1, gotI64, gotI32_2)
		}
	})

	t.Run("small_types", func(t *testing.T) {
		var gotBool bool
		var gotI8 int8
		var gotU8 uint8
		var gotI16 int16
		var gotU16 uint16
		var gotI32 int32
		goCallback := func(r1, r2, r3, r4, r5, r6, r7, r8 int64, b bool, i8 int8, u8 uint8, i16 int16, u16 uint16, i32 int32) {
			gotBool = b
			gotI8 = i8
			gotU8 = u8
			gotI16 = i16
			gotU16 = u16
			gotI32 = i32
		}

		var callCallbackSmallTypes func(uintptr)
		purego.RegisterLibFunc(&callCallbackSmallTypes, lib, "callCallbackSmallTypes")

		cb := purego.NewCallback(goCallback)
		callCallbackSmallTypes(cb)

		if !gotBool || gotI8 != -42 || gotU8 != 200 || gotI16 != -1000 || gotU16 != 50000 || gotI32 != 123456 {
			t.Errorf("callback received (bool=%v, i8=%d, u8=%d, i16=%d, u16=%d, i32=%d), want (true, -42, 200, -1000, 50000, 123456)",
				gotBool, gotI8, gotU8, gotI16, gotU16, gotI32)
		}
	})
}
