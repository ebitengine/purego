// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2023 The Ebitengine Authors

//go:build darwin || (!cgo && linux)

package purego_test
import (
	"testing"

	"github.com/ebitengine/purego"
)

func TestNewCallbackFloat64(t *testing.T) {
	// This tests the maximum number of arguments a function to NewCallback can take
	const (
		expectCbTotal    = -3
		expectedCbTotalF = float64(36)
	)
	var cbTotal int
	var cbTotalF float64
	imp := purego.NewCallback(func(a1, a2, a3, a4, a5, a6, a7, a8, a9 int,
		f1, f2, f3, f4, f5, f6, f7, f8 float64) {
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

func TestNewCallbackFloat32(t *testing.T) {
	// This tests the maximum number of float32 arguments a function to NewCallback can take
	const (
		expectCbTotal    = 6
		expectedCbTotalF = float32(45)
	)
	var cbTotal int
	var cbTotalF float32
	imp := purego.NewCallback(func(a1, a2, a3, a4, a5, a6, a7, a8 int,
		f1, f2, f3, f4, f5, f6, f7, f8, f9 float32) {
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
		expectedCbTotalF32 = float32(30)
		expectedCbTotalF64 = float64(15)
	)
	var cbTotalF32 float32
	var cbTotalF64 float64
	imp := purego.NewCallback(func(f1, f2, f3 float32, f4, f5, f6 float64, f7, f8, f9 float32) {
		cbTotalF32 = f1 + f2 + f3 + f7 + f8 + f9
		cbTotalF64 = f4 + f5 + f6

	})
	var fn func(f1, f2, f3 float32, f4, f5, f6 float64, f7, f8, f9 float32)
	purego.RegisterFunc(&fn, imp)
	fn(1, 2, 3, 4, 5, 6, 7, 8, 9)

	if cbTotalF32 != expectedCbTotalF32 {
		t.Errorf("cbTotalF32 not correct got %f but wanted %f", cbTotalF32, expectedCbTotalF32)
	}
	if cbTotalF64 != expectedCbTotalF64 {
		t.Errorf("cbTotalF64 not correct got %f but wanted %f", cbTotalF64, expectedCbTotalF64)
	}
}
