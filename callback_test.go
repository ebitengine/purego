// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2023 The Ebitengine Authors

//go:build darwin || (linux && (!cgo || amd64 || arm64))

package purego_test

import (
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
