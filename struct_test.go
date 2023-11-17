// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2023 The Ebitengine Authors

//go:build darwin || (linux && (!cgo || amd64 || arm64))

package purego_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ebitengine/purego"
)

func TestRegisterFunc_structArgs(t *testing.T) {
	libFileName := filepath.Join(t.TempDir(), "structtest.so")
	t.Logf("Build %v", libFileName)

	if err := buildSharedLib("CC", libFileName, filepath.Join("structtest", "struct.c")); err != nil {
		t.Fatal(err)
	}
	defer os.Remove(libFileName)

	lib, err := purego.Dlopen(libFileName, purego.RTLD_NOW|purego.RTLD_GLOBAL)
	if err != nil {
		t.Fatalf("Dlopen(%q) failed: %v", libFileName, err)
	}

	const (
		expected               = 0xdeadbeef
		expectedLong   uint64  = 0xdeadbeefcafebabe
		expectedFloat  float32 = 10
		expectedDouble float64 = 10
	)

	{
		type Empty struct{}
		var NoStruct func(Empty) int64
		purego.RegisterLibFunc(&NoStruct, lib, "NoStruct")
		if ret := NoStruct(Empty{}); ret != expected {
			t.Fatalf("NoStruct returned %#x wanted %#x", ret, expected)
		}
	}
	{
		type GreaterThan16Bytes struct {
			x, y, z *int64
		}
		var x, y, z int64 = 0xEF, 0xBE00, 0xDEAD0000
		var GreaterThan16BytesFn func(GreaterThan16Bytes) int64
		purego.RegisterLibFunc(&GreaterThan16BytesFn, lib, "GreaterThan16Bytes")
		if ret := GreaterThan16BytesFn(GreaterThan16Bytes{x: &x, y: &y, z: &z}); ret != expected {
			t.Fatalf("GreaterThan16Bytes returned %#x wanted %#x", ret, expected)
		}
	}
	{
		type GreaterThan16Bytes struct {
			x, y, z *int64
		}
		var x, y, z int64 = 0xEF, 0xBE00, 0xDEAD0000
		var AfterRegisters func(a, b, c, d, e, f, g, h int, bytes GreaterThan16Bytes) int64
		purego.RegisterLibFunc(&AfterRegisters, lib, "AfterRegisters")
		if ret := AfterRegisters(0xD0000000, 0xE000000, 0xA00000, 0xD0000, 0xB000, 0xE00, 0xE0, 0xF, GreaterThan16Bytes{x: &x, y: &y, z: &z}); ret != expected {
			t.Fatalf("AfterRegisters returned %#x wanted %#x", ret, expected)
		}
	}
	{
		type IntLessThan16Bytes struct {
			x, y int64
		}
		var IntLessThan16BytesFn func(bytes IntLessThan16Bytes) int64
		purego.RegisterLibFunc(&IntLessThan16BytesFn, lib, "IntLessThan16Bytes")
		if ret := IntLessThan16BytesFn(IntLessThan16Bytes{0xDEAD0000, 0xBEEF}); ret != expected {
			t.Fatalf("IntLessThan16BytesFn returned %#x wanted %#x", ret, expected)
		}
	}
	{
		type FloatLessThan16Bytes struct {
			x, y float32
		}
		var FloatLessThan16BytesFn func(FloatLessThan16Bytes) float32
		purego.RegisterLibFunc(&FloatLessThan16BytesFn, lib, "FloatLessThan16Bytes")
		if ret := FloatLessThan16BytesFn(FloatLessThan16Bytes{3, 7}); ret != expectedFloat {
			t.Fatalf("FloatLessThan16Bytes returned %f wanted %f", ret, expectedFloat)
		}
	}
	{
		type DoubleStruct struct {
			x float64
		}
		var DoubleStructFn func(DoubleStruct) float64
		purego.RegisterLibFunc(&DoubleStructFn, lib, "DoubleStruct")
		if ret := DoubleStructFn(DoubleStruct{10}); ret != expectedDouble {
			t.Fatalf("DoubleStruct returned %f wanted %f", ret, expectedFloat)
		}
	}
	{
		type LargeFloatStruct struct {
			a, b, c, d, e, f float64
		}
		var LargeFloatStructFn func(LargeFloatStruct) float64
		purego.RegisterLibFunc(&LargeFloatStructFn, lib, "LargeFloatStruct")
		if ret := LargeFloatStructFn(LargeFloatStruct{1, 2, 3, 4, 5, -5}); ret != expectedDouble {
			t.Fatalf("LargeFloatStructFn returned %f wanted %f", ret, expectedFloat)
		}
		var LargeFloatStructWithRegs func(a, b, c float64, s LargeFloatStruct) float64
		purego.RegisterLibFunc(&LargeFloatStructWithRegs, lib, "LargeFloatStructWithRegs")
		if ret := LargeFloatStructWithRegs(1, -1, 0, LargeFloatStruct{1, 2, 3, 4, 5, -5}); ret != expectedDouble {
			t.Fatalf("LargeFloatStructWithRegs returned %f wanted %f", ret, expectedFloat)
		}
	}
	{
		type Rect struct {
			x, y, w, h float64
		}
		var RectangleWithRegs func(a, b, c, d, e float64, rect Rect) float64
		purego.RegisterLibFunc(&RectangleWithRegs, lib, "Rectangle")
		if ret := RectangleWithRegs(1, 2, 3, 4, 0, Rect{1, 2, 3, -4}); ret != expectedDouble {
			t.Fatalf("RectangleWithRegs returned %f wanted %f", ret, expectedFloat)
		}
		var Rectangle func(rect Rect) float64
		purego.RegisterLibFunc(&Rectangle, lib, "Rectangle")
		if ret := Rectangle(Rect{1, 2, 3, 4}); ret != expectedDouble {
			t.Fatalf("Rectangle returned %f wanted %f", ret, expectedFloat)
		}
	}
	{
		type Char4Bytes struct {
			a, b, c, d byte
		}
		var Char4BytesFn func(Char4Bytes) uint32
		purego.RegisterLibFunc(&Char4BytesFn, lib, "Char4Bytes")
		if ret := Char4BytesFn(Char4Bytes{a: 0xDE, b: 0xAD, c: 0xBE, d: 0xEF}); ret != expected {
			t.Fatalf("Char4Bytes returned %#x wanted %#x", ret, expected)
		}
	}
	{
		type Short struct {
			a, b, c, d uint16
		}
		var ShortFn func(Short) uint64
		purego.RegisterLibFunc(&ShortFn, lib, "Short")
		if ret := ShortFn(Short{a: 0xDEAD, b: 0xBEEF, c: 0xCAFE, d: 0xBABE}); ret != expectedLong {
			t.Fatalf("ShortFn returned %#x wanted %#x", ret, expectedLong)
		}
	}
	{
		type Int struct {
			a, b uint32
		}
		var IntFn func(Int) uint64
		purego.RegisterLibFunc(&IntFn, lib, "Int")
		if ret := IntFn(Int{a: 0xDEADBEEF, b: 0xCAFEBABE}); ret != expectedLong {
			t.Fatalf("IntFn returned %#x wanted %#x", ret, expectedLong)
		}
	}
	{
		type Long struct {
			a uint64
		}
		var LongFn func(Long) uint64
		purego.RegisterLibFunc(&LongFn, lib, "Long")
		if ret := LongFn(Long{a: 0xDEADBEEFCAFEBABE}); ret != expectedLong {
			t.Fatalf("LongFn returned %#x wanted %#x", ret, expectedLong)
		}
	}
	//{
	//	type Array4Chars struct {
	//		a [4]uint8
	//	}
	//	var Array4CharsFn func(chars Array4Chars) uint32
	//	purego.RegisterLibFunc(&Array4CharsFn, lib, "Long")
	//	if ret := Array4CharsFn(Array4Chars{a: [...]uint8{0xDE, 0xAD, 0xBE, 0xEF}}); ret != expected {
	//		t.Fatalf("LongFn returned %#x wanted %#x", ret, expectedLong)
	//	}
	//}
}
