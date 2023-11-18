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
		expectedUnsigned         = 0xdeadbeef
		expectedSigned           = -123
		expectedOdd              = 12 + 23 + 46
		expectedLong     uint64  = 0xdeadbeefcafebabe
		expectedFloat    float32 = 10
		expectedDouble   float64 = 10
	)

	{
		type Empty struct{}
		var NoStruct func(Empty) int64
		purego.RegisterLibFunc(&NoStruct, lib, "NoStruct")
		if ret := NoStruct(Empty{}); ret != expectedUnsigned {
			t.Fatalf("NoStruct returned %#x wanted %#x", ret, expectedUnsigned)
		}
	}
	{
		type GreaterThan16Bytes struct {
			x, y, z *int64
		}
		var x, y, z int64 = 0xEF, 0xBE00, 0xDEAD0000
		var GreaterThan16BytesFn func(GreaterThan16Bytes) int64
		purego.RegisterLibFunc(&GreaterThan16BytesFn, lib, "GreaterThan16Bytes")
		if ret := GreaterThan16BytesFn(GreaterThan16Bytes{x: &x, y: &y, z: &z}); ret != expectedUnsigned {
			t.Fatalf("GreaterThan16Bytes returned %#x wanted %#x", ret, expectedUnsigned)
		}
	}
	{
		type GreaterThan16Bytes struct {
			x, y, z *int64
		}
		var x, y, z int64 = 0xEF, 0xBE00, 0xDEAD0000
		var AfterRegisters func(a, b, c, d, e, f, g, h int, bytes GreaterThan16Bytes) int64
		purego.RegisterLibFunc(&AfterRegisters, lib, "AfterRegisters")
		if ret := AfterRegisters(0xD0000000, 0xE000000, 0xA00000, 0xD0000, 0xB000, 0xE00, 0xE0, 0xF, GreaterThan16Bytes{x: &x, y: &y, z: &z}); ret != expectedUnsigned {
			t.Fatalf("AfterRegisters returned %#x wanted %#x", ret, expectedUnsigned)
		}
	}
	{
		type IntLessThan16Bytes struct {
			x, y int64
		}
		var IntLessThan16BytesFn func(bytes IntLessThan16Bytes) int64
		purego.RegisterLibFunc(&IntLessThan16BytesFn, lib, "IntLessThan16Bytes")
		if ret := IntLessThan16BytesFn(IntLessThan16Bytes{0xDEAD0000, 0xBEEF}); ret != expectedUnsigned {
			t.Fatalf("IntLessThan16BytesFn returned %#x wanted %#x", ret, expectedUnsigned)
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
		type UnsignedChar4Bytes struct {
			a, b, c, d byte
		}
		var UnsignedChar4BytesFn func(UnsignedChar4Bytes) uint32
		purego.RegisterLibFunc(&UnsignedChar4BytesFn, lib, "UnsignedChar4Bytes")
		if ret := UnsignedChar4BytesFn(UnsignedChar4Bytes{a: 0xDE, b: 0xAD, c: 0xBE, d: 0xEF}); ret != expectedUnsigned {
			t.Fatalf("Char8Bytes returned %#x wanted %#x", ret, expectedUnsigned)
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
	//	if ret := Array4CharsFn(Array4Chars{a: [...]uint8{0xDE, 0xAD, 0xBE, 0xEF}}); ret != expectedUnsigned {
	//		t.Fatalf("LongFn returned %#x wanted %#x", ret, expectedLong)
	//	}
	//}
	{
		type Char8Bytes struct {
			a, b, c, d, e, f, g, h int8
		}
		var Char8BytesFn func(Char8Bytes) int32
		purego.RegisterLibFunc(&Char8BytesFn, lib, "Char8Bytes")
		if ret := Char8BytesFn(Char8Bytes{a: -128, b: 127, c: 3, d: -88, e: -3, f: 34, g: -48, h: -20}); ret != expectedSigned {
			t.Fatalf("Char8Bytes returned %d wanted %d", ret, expectedSigned)
		}
	}
	{
		type Odd struct {
			a, b, c byte
		}
		var OddFn func(Odd) int32
		purego.RegisterLibFunc(&OddFn, lib, "Odd")
		if ret := OddFn(Odd{a: 12, b: 23, c: 46}); ret != expectedOdd {
			t.Fatalf("OddFn returned %d wanted %d", ret, expectedOdd)
		}
	}
}
