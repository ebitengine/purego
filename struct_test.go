// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024 The Ebitengine Authors

//go:build darwin && (arm64 || amd64)

package purego_test

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"unsafe"

	"github.com/ebitengine/purego"
)

func TestRegisterFunc_structArgs(t *testing.T) {
	libFileName := filepath.Join(t.TempDir(), "structtest.so")
	t.Logf("Build %v", libFileName)

	if err := buildSharedLib("CC", libFileName, filepath.Join("testdata", "structtest", "struct_test.c")); err != nil {
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
		type EmptyEmpty struct{}
		var EmptyEmptyFn func(EmptyEmpty) int64
		purego.RegisterLibFunc(&EmptyEmptyFn, lib, "EmptyEmpty")
		if ret := EmptyEmptyFn(EmptyEmpty{}); ret != expectedUnsigned {
			t.Fatalf("EmptyEmpty returned %#x wanted %#x", ret, expectedUnsigned)
		}
		var EmptyEmptyWithReg func(uint32, EmptyEmpty, uint32) int64
		purego.RegisterLibFunc(&EmptyEmptyWithReg, lib, "EmptyEmptyWithReg")
		if ret := EmptyEmptyWithReg(0xdead, EmptyEmpty{}, 0xbeef); ret != expectedUnsigned {
			t.Fatalf("EmptyEmptyWithReg returned %#x wanted %#x", ret, expectedUnsigned)
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
		type GreaterThan16BytesStruct struct {
			a struct {
				x, y, z *int64
			}
		}
		var x, y, z int64 = 0xEF, 0xBE00, 0xDEAD0000
		var GreaterThan16BytesStructFn func(GreaterThan16BytesStruct) int64
		purego.RegisterLibFunc(&GreaterThan16BytesStructFn, lib, "GreaterThan16BytesStruct")
		if ret := GreaterThan16BytesStructFn(GreaterThan16BytesStruct{a: struct{ x, y, z *int64 }{x: &x, y: &y, z: &z}}); ret != expectedUnsigned {
			t.Fatalf("GreaterThan16BytesStructFn returned %#x wanted %#x", ret, expectedUnsigned)
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
		var BeforeRegisters func(bytes GreaterThan16Bytes, a, b int64) uint64
		z -= 0xFF
		purego.RegisterLibFunc(&BeforeRegisters, lib, "BeforeRegisters")
		if ret := BeforeRegisters(GreaterThan16Bytes{&x, &y, &z}, 0x0F, 0xF0); ret != expectedUnsigned {
			t.Fatalf("BeforeRegisters returned %#x wanted %#x", ret, expectedUnsigned)
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
		type ThreeSmallFields struct {
			x, y, z float32
		}
		var ThreeSmallFieldsFn func(ThreeSmallFields) float32
		purego.RegisterLibFunc(&ThreeSmallFieldsFn, lib, "ThreeSmallFields")
		if ret := ThreeSmallFieldsFn(ThreeSmallFields{1, 2, 7}); ret != expectedFloat {
			t.Fatalf("ThreeSmallFields returned %f wanted %f", ret, expectedFloat)
		}
	}
	{
		type FloatAndInt struct {
			x float32
			y int32
		}
		var FloatAndIntFn func(FloatAndInt) float32
		purego.RegisterLibFunc(&FloatAndIntFn, lib, "FloatAndInt")
		if ret := FloatAndIntFn(FloatAndInt{3, 7}); ret != expectedFloat {
			t.Fatalf("FloatAndIntFn returned %f wanted %f", ret, expectedFloat)
		}
	}
	{
		type DoubleStruct struct {
			x float64
		}
		var DoubleStructFn func(DoubleStruct) float64
		purego.RegisterLibFunc(&DoubleStructFn, lib, "DoubleStruct")
		if ret := DoubleStructFn(DoubleStruct{10}); ret != expectedDouble {
			t.Fatalf("DoubleStruct returned %f wanted %f", ret, expectedDouble)
		}
	}
	{
		type TwoDoubleStruct struct {
			x, y float64
		}
		var TwoDoubleStructFn func(TwoDoubleStruct) float64
		purego.RegisterLibFunc(&TwoDoubleStructFn, lib, "TwoDoubleStruct")
		if ret := TwoDoubleStructFn(TwoDoubleStruct{3, 7}); ret != expectedDouble {
			t.Fatalf("TwoDoubleStruct returned %f wanted %f", ret, expectedDouble)
		}
	}
	{
		type TwoDoubleTwoStruct struct {
			x struct {
				x, y float64
			}
		}
		var TwoDoubleTwoStructFn func(TwoDoubleTwoStruct) float64
		purego.RegisterLibFunc(&TwoDoubleTwoStructFn, lib, "TwoDoubleTwoStruct")
		if ret := TwoDoubleTwoStructFn(TwoDoubleTwoStruct{x: struct{ x, y float64 }{x: 3, y: 7}}); ret != expectedDouble {
			t.Fatalf("TwoDoubleTwoStruct returned %f wanted %f", ret, expectedDouble)
		}
	}
	{
		type ThreeDoubleStruct struct {
			x, y, z float64
		}
		var ThreeDoubleStructFn func(ThreeDoubleStruct) float64
		purego.RegisterLibFunc(&ThreeDoubleStructFn, lib, "ThreeDoubleStruct")
		if ret := ThreeDoubleStructFn(ThreeDoubleStruct{1, 3, 6}); ret != expectedDouble {
			t.Fatalf("ThreeDoubleStructFn returned %f wanted %f", ret, expectedDouble)
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
		purego.RegisterLibFunc(&RectangleWithRegs, lib, "RectangleWithRegs")
		if ret := RectangleWithRegs(1, 2, 3, 4, -2, Rect{1, 2, 3, -4}); ret != expectedDouble {
			t.Fatalf("RectangleWithRegs returned %f wanted %f", ret, expectedDouble)
		}
		var RectangleSubtract func(rect Rect) float64
		purego.RegisterLibFunc(&RectangleSubtract, lib, "RectangleSubtract")
		if ret := RectangleSubtract(Rect{15, 5, 3, 7}); ret != expectedDouble {
			t.Fatalf("RectangleSubtract returned %f wanted %f", ret, expectedDouble)
		}
		var Rectangle func(rect Rect) float64
		purego.RegisterLibFunc(&Rectangle, lib, "Rectangle")
		if ret := Rectangle(Rect{1, 2, 3, 4}); ret != expectedDouble {
			t.Fatalf("Rectangle returned %f wanted %f", ret, expectedFloat)
		}
	}
	{
		type FloatArray struct {
			a [2]float64
		}
		var FloatArrayFn func(rect FloatArray) float64
		purego.RegisterLibFunc(&FloatArrayFn, lib, "FloatArray")
		if ret := FloatArrayFn(FloatArray{a: [2]float64{3, 7}}); ret != expectedDouble {
			t.Fatalf("FloatArray returned %f wanted %f", ret, expectedFloat)
		}
	}
	{
		type UnsignedChar4Bytes struct {
			a, b, c, d byte
		}
		var UnsignedChar4BytesFn func(UnsignedChar4Bytes) uint32
		purego.RegisterLibFunc(&UnsignedChar4BytesFn, lib, "UnsignedChar4Bytes")
		if ret := UnsignedChar4BytesFn(UnsignedChar4Bytes{a: 0xDE, b: 0xAD, c: 0xBE, d: 0xEF}); ret != expectedUnsigned {
			t.Fatalf("UnsignedChar4BytesFn returned %#x wanted %#x", ret, expectedUnsigned)
		}
	}
	{
		type UnsignedChar4BytesStruct struct {
			x struct {
				a byte
			}
			y struct {
				b byte
			}
			z struct {
				c byte
			}
			w struct {
				d byte
			}
		}
		var UnsignedChar4BytesStructFn func(UnsignedChar4BytesStruct) uint32
		purego.RegisterLibFunc(&UnsignedChar4BytesStructFn, lib, "UnsignedChar4BytesStruct")
		if ret := UnsignedChar4BytesStructFn(UnsignedChar4BytesStruct{
			x: struct{ a byte }{a: 0xDE},
			y: struct{ b byte }{b: 0xAD},
			z: struct{ c byte }{c: 0xBE},
			w: struct{ d byte }{d: 0xEF},
		}); ret != expectedUnsigned {
			t.Fatalf("UnsignedChar4BytesStructFn returned %#x wanted %#x", ret, expectedUnsigned)
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
	{
		type Char2Short1 struct {
			a, b byte
			c    uint16
		}
		var Char2Short1s func(Char2Short1) int32
		purego.RegisterLibFunc(&Char2Short1s, lib, "Char2Short1s")
		if ret := Char2Short1s(Char2Short1{a: 12, b: 23, c: 46}); ret != expectedOdd {
			t.Fatalf("Char2Short1s returned %d wanted %d", ret, expectedOdd)
		}
	}
	{
		type SignedChar2Short1 struct {
			a, b int8
			c    int16
		}
		var SignedChar2Short1Fn func(SignedChar2Short1) int32
		purego.RegisterLibFunc(&SignedChar2Short1Fn, lib, "SignedChar2Short1")
		if ret := SignedChar2Short1Fn(SignedChar2Short1{a: 100, b: -23, c: -200}); ret != expectedSigned {
			t.Fatalf("SignedChar2Short1Fn returned %d wanted %d", ret, expectedSigned)
		}
	}
	{
		type Array4UnsignedChars struct {
			a [4]uint8
		}
		var Array4UnsignedCharsFn func(chars Array4UnsignedChars) uint32
		purego.RegisterLibFunc(&Array4UnsignedCharsFn, lib, "Array4UnsignedChars")
		if ret := Array4UnsignedCharsFn(Array4UnsignedChars{a: [...]uint8{0xDE, 0xAD, 0xBE, 0xEF}}); ret != expectedUnsigned {
			t.Fatalf("Array4UnsignedCharsFn returned %#x wanted %#x", ret, expectedUnsigned)
		}
	}
	{
		type Array3UnsignedChar struct {
			a [3]uint8
		}
		var Array3UnsignedChars func(chars Array3UnsignedChar) uint32
		purego.RegisterLibFunc(&Array3UnsignedChars, lib, "Array3UnsignedChars")
		if ret := Array3UnsignedChars(Array3UnsignedChar{a: [...]uint8{0xDE, 0xAD, 0xBE}}); ret != expectedUnsigned {
			t.Fatalf("Array4UnsignedCharsFn returned %#x wanted %#x", ret, expectedUnsigned)
		}
	}
	{
		type Array2UnsignedShort struct {
			a [2]uint16
		}
		var Array2UnsignedShorts func(chars Array2UnsignedShort) uint32
		purego.RegisterLibFunc(&Array2UnsignedShorts, lib, "Array2UnsignedShorts")
		if ret := Array2UnsignedShorts(Array2UnsignedShort{a: [...]uint16{0xDEAD, 0xBEEF}}); ret != expectedUnsigned {
			t.Fatalf("Array4UnsignedCharsFn returned %#x wanted %#x", ret, expectedUnsigned)
		}
	}
	{
		type Array4Chars struct {
			a [4]int8
		}
		var Array4CharsFn func(chars Array4Chars) int32
		purego.RegisterLibFunc(&Array4CharsFn, lib, "Array4Chars")
		if ret := Array4CharsFn(Array4Chars{a: [...]int8{100, -127, 4, -100}}); ret != expectedSigned {
			t.Fatalf("Array4CharsFn returned %#x wanted %#x", ret, expectedSigned)
		}
	}
	{
		type Array2Short struct {
			a [2]int16
		}
		var Array2Shorts func(chars Array2Short) int32
		purego.RegisterLibFunc(&Array2Shorts, lib, "Array2Shorts")
		if ret := Array2Shorts(Array2Short{a: [...]int16{-333, 210}}); ret != expectedSigned {
			t.Fatalf("Array4Shorts returned %#x wanted %#x", ret, expectedSigned)
		}
	}
	{
		type Array3Short struct {
			a [3]int16
		}
		var Array3Shorts func(chars Array3Short) int32
		purego.RegisterLibFunc(&Array3Shorts, lib, "Array3Shorts")
		if ret := Array3Shorts(Array3Short{a: [...]int16{-333, 100, 110}}); ret != expectedSigned {
			t.Fatalf("Array4Shorts returned %#x wanted %#x", ret, expectedSigned)
		}
	}
	{
		type BoolStruct struct {
			b bool
		}
		var BoolStructFn func(BoolStruct) bool
		purego.RegisterLibFunc(&BoolStructFn, lib, "BoolStruct")
		if ret := BoolStructFn(BoolStruct{true}); ret != true {
			t.Fatalf("BoolStructFn returned %v wanted %v", ret, true)
		}
		if ret := BoolStructFn(BoolStruct{false}); ret != false {
			t.Fatalf("BoolStructFn returned %v wanted %v", ret, false)
		}
	}
	{
		type BoolFloat struct {
			b bool
			_ [3]byte // purego won't do padding for you so make sure it aligns properly with C struct
			f float32
		}
		var BoolFloatFn func(BoolFloat) float32
		purego.RegisterLibFunc(&BoolFloatFn, lib, "BoolFloat")
		if ret := BoolFloatFn(BoolFloat{b: true, f: 10}); ret != expectedFloat {
			t.Fatalf("BoolFloatFn returned %f wanted %f", ret, expectedFloat)
		}
		if ret := BoolFloatFn(BoolFloat{b: false, f: 10}); ret != -expectedFloat {
			t.Fatalf("BoolFloatFn returned %f wanted %f", ret, -expectedFloat)
		}
	}
	{
		type point struct{ x, y float64 }
		type size struct{ width, height float64 }
		type Content struct {
			point point
			size  size
		}
		var InitWithContentRect func(*int, Content, int32, int32, bool) uint64
		purego.RegisterLibFunc(&InitWithContentRect, lib, "InitWithContentRect")
		if ret := InitWithContentRect(new(int),
			// These numbers are created so that when added together and then divided by 11 it produces 0xdeadbeef
			Content{point{x: 41_000_000_000, y: 95_000_000}, size{width: 214_000, height: 149}},
			15, 4, true); ret != expectedUnsigned {
			t.Fatalf("InitWithContentRect returned %d wanted %#x", ret, expectedUnsigned)
		}
	}
}

func TestRegisterFunc_structReturns(t *testing.T) {
	libFileName := filepath.Join(t.TempDir(), "structreturntest.so")
	t.Logf("Build %v", libFileName)

	if err := buildSharedLib("CC", libFileName, filepath.Join("testdata", "structtest", "structreturn_test.c")); err != nil {
		t.Fatal(err)
	}
	defer os.Remove(libFileName)

	lib, err := purego.Dlopen(libFileName, purego.RTLD_NOW|purego.RTLD_GLOBAL)
	if err != nil {
		t.Fatalf("Dlopen(%q) failed: %v", libFileName, err)
	}

	{
		type Empty struct{}
		var ReturnEmpty func() Empty
		purego.RegisterLibFunc(&ReturnEmpty, lib, "ReturnEmpty")
		ret := ReturnEmpty()
		_ = ret
	}
	{
		type inner struct{ a int16 }
		type StructInStruct struct {
			a inner
			b inner
			c inner
		}
		var ReturnStructInStruct func(a, b, c int16) StructInStruct
		purego.RegisterLibFunc(&ReturnStructInStruct, lib, "ReturnStructInStruct")
		expected := StructInStruct{inner{^int16(0)}, inner{2}, inner{3}}
		if ret := ReturnStructInStruct(^int16(0), 2, 3); ret != expected {
			t.Fatalf("StructInStruct returned %+v wanted %+v", ret, expected)
		}
	}
	{
		type ThreeShorts struct{ a, b, c int16 }
		var ReturnThreeShorts func(a, b, c int16) ThreeShorts
		purego.RegisterLibFunc(&ReturnThreeShorts, lib, "ReturnThreeShorts")
		expected := ThreeShorts{^int16(0), 2, 3}
		if ret := ReturnThreeShorts(^int16(0), 2, 3); ret != expected {
			t.Fatalf("ReturnThreeShorts returned %+v wanted %+v", ret, expected)
		}
	}
	{
		type FourShorts struct{ a, b, c, d int16 }
		var ReturnFourShorts func(a, b, c, d int16) FourShorts
		purego.RegisterLibFunc(&ReturnFourShorts, lib, "ReturnFourShorts")
		expected := FourShorts{^int16(0), 2, 3, 4}
		if ret := ReturnFourShorts(^int16(0), 2, 3, 4); ret != expected {
			t.Fatalf("ReturnFourShorts returned %+v wanted %+v", ret, expected)
		}
	}
	{
		type OneLong struct{ a int64 }
		var ReturnOneLong func(a int64) OneLong
		purego.RegisterLibFunc(&ReturnOneLong, lib, "ReturnOneLong")
		expected := OneLong{5}
		if ret := ReturnOneLong(5); ret != expected {
			t.Fatalf("ReturnOneLong returned %+v wanted %+v", ret, expected)
		}
	}
	{
		type TwoLongs struct{ a, b int64 }
		var ReturnTwoLongs func(a, b int64) TwoLongs
		purego.RegisterLibFunc(&ReturnTwoLongs, lib, "ReturnTwoLongs")
		expected := TwoLongs{1, 2}
		if ret := ReturnTwoLongs(1, 2); ret != expected {
			t.Fatalf("ReturnTwoLongs returned %+v wanted %+v", ret, expected)
		}
	}
	{
		type ThreeLongs struct{ a, b, c int64 }
		var ReturnThreeLongs func(a, b, c int64) ThreeLongs
		purego.RegisterLibFunc(&ReturnThreeLongs, lib, "ReturnThreeLongs")
		expected := ThreeLongs{1, 2, 3}
		if ret := ReturnThreeLongs(1, 2, 3); ret != expected {
			t.Fatalf("ReturnThreeLongs returned %+v wanted %+v", ret, expected)
		}
	}
	{
		type OneFloat struct{ a float32 }
		var ReturnOneFloat func(a float32) OneFloat
		purego.RegisterLibFunc(&ReturnOneFloat, lib, "ReturnOneFloat")
		expected := OneFloat{1}
		if ret := ReturnOneFloat(1); ret != expected {
			t.Fatalf("ReturnOneFloat returned %+v wanted %+v", ret, expected)
		}
	}
	{
		type TwoFloats struct{ a, b float32 }
		var ReturnTwoFloats func(a, b float32) TwoFloats
		purego.RegisterLibFunc(&ReturnTwoFloats, lib, "ReturnTwoFloats")
		expected := TwoFloats{3, 10}
		if ret := ReturnTwoFloats(5, 2); ret != expected {
			t.Fatalf("ReturnTwoFloats returned %+v wanted %+v", ret, expected)
		}
	}
	{
		type ThreeFloats struct{ a, b, c float32 }
		var ReturnThreeFloats func(a, b, c float32) ThreeFloats
		purego.RegisterLibFunc(&ReturnThreeFloats, lib, "ReturnThreeFloats")
		expected := ThreeFloats{1, 2, 3}
		if ret := ReturnThreeFloats(1, 2, 3); ret != expected {
			t.Fatalf("ReturnThreeFloats returned %+v wanted %+v", ret, expected)
		}
	}
	{
		type OneDouble struct{ a float64 }
		var ReturnOneDouble func(a float64) OneDouble
		purego.RegisterLibFunc(&ReturnOneDouble, lib, "ReturnOneDouble")
		expected := OneDouble{1}
		if ret := ReturnOneDouble(1); ret != expected {
			t.Fatalf("ReturnOneDouble returned %+v wanted %+v", ret, expected)
		}
	}
	{
		type TwoDoubles struct{ a, b float64 }
		var ReturnTwoDoubles func(a, b float64) TwoDoubles
		purego.RegisterLibFunc(&ReturnTwoDoubles, lib, "ReturnTwoDoubles")
		expected := TwoDoubles{1, 2}
		if ret := ReturnTwoDoubles(1, 2); ret != expected {
			t.Fatalf("ReturnTwoDoubles returned %+v wanted %+v", ret, expected)
		}
	}
	{
		type ThreeDoubles struct{ a, b, c float64 }
		var ReturnThreeDoubles func(a, b, c float64) ThreeDoubles
		purego.RegisterLibFunc(&ReturnThreeDoubles, lib, "ReturnThreeDoubles")
		expected := ThreeDoubles{1, 2, 3}
		if ret := ReturnThreeDoubles(1, 2, 3); ret != expected {
			t.Fatalf("ReturnThreeDoubles returned %+v wanted %+v", ret, expected)
		}
	}
	{
		type FourDoubles struct{ a, b, c, d float64 }
		var ReturnFourDoubles func(a, b, c, d float64) FourDoubles
		purego.RegisterLibFunc(&ReturnFourDoubles, lib, "ReturnFourDoubles")
		expected := FourDoubles{1, 2, 3, 4}
		if ret := ReturnFourDoubles(1, 2, 3, 4); ret != expected {
			t.Fatalf("ReturnFourDoubles returned %+v wanted %+v", ret, expected)
		}
	}
	{
		type FourDoublesInternal struct {
			f struct{ a, b float64 }
			g struct{ c, d float64 }
		}
		var ReturnFourDoublesInternal func(a, b, c, d float64) FourDoublesInternal
		purego.RegisterLibFunc(&ReturnFourDoublesInternal, lib, "ReturnFourDoublesInternal")
		expected := FourDoublesInternal{f: struct{ a, b float64 }{a: 1, b: 2}, g: struct{ c, d float64 }{c: 3, d: 4}}
		if ret := ReturnFourDoublesInternal(1, 2, 3, 4); ret != expected {
			t.Fatalf("ReturnFourDoublesInternal returned %+v wanted %+v", ret, expected)
		}
	}
	{
		type FiveDoubles struct{ a, b, c, d, e float64 }
		var ReturnFiveDoubles func(a, b, c, d, e float64) FiveDoubles
		purego.RegisterLibFunc(&ReturnFiveDoubles, lib, "ReturnFiveDoubles")
		expected := FiveDoubles{1, 2, 3, 4, 5}
		if ret := ReturnFiveDoubles(1, 2, 3, 4, 5); ret != expected {
			t.Fatalf("ReturnFiveDoubles returned %+v wanted %+v", ret, expected)
		}
	}
	{
		type OneFloatOneDouble struct {
			a float32
			_ float32
			b float64
		}
		var ReturnOneFloatOneDouble func(a float32, b float64) OneFloatOneDouble
		purego.RegisterLibFunc(&ReturnOneFloatOneDouble, lib, "ReturnOneFloatOneDouble")
		expected := OneFloatOneDouble{a: 1, b: 2}
		if ret := ReturnOneFloatOneDouble(1, 2); ret != expected {
			t.Fatalf("ReturnOneFloatOneDouble returned %+v wanted %+v", ret, expected)
		}
	}
	{
		type OneDoubleOneFloat struct {
			a float64
			b float32
		}
		var ReturnOneDoubleOneFloat func(a float64, b float32) OneDoubleOneFloat
		purego.RegisterLibFunc(&ReturnOneDoubleOneFloat, lib, "ReturnOneDoubleOneFloat")
		expected := OneDoubleOneFloat{1, 2}
		if ret := ReturnOneDoubleOneFloat(1, 2); ret != expected {
			t.Fatalf("ReturnOneDoubleOneFloat returned %+v wanted %+v", ret, expected)
		}
	}
	{
		type Unaligned1 struct {
			a int8
			_ [1]int8
			b int16
			_ [1]int32
			c int64
		}
		var ReturnUnaligned1 func(a int8, b int16, c int64) Unaligned1
		purego.RegisterLibFunc(&ReturnUnaligned1, lib, "ReturnUnaligned1")
		expected := Unaligned1{a: 1, b: 2, c: 3}
		if ret := ReturnUnaligned1(1, 2, 3); ret != expected {
			t.Fatalf("ReturnUnaligned1 returned %+v wanted %+v", ret, expected)
		}
	}
	{
		type Mixed1 struct {
			a float32
			b int32
		}
		var ReturnMixed1 func(a float32, b int32) Mixed1
		purego.RegisterLibFunc(&ReturnMixed1, lib, "ReturnMixed1")
		expected := Mixed1{1, 2}
		if ret := ReturnMixed1(1, 2); ret != expected {
			t.Fatalf("ReturnMixed1 returned %+v wanted %+v", ret, expected)
		}
	}
	{
		type Mixed2 struct {
			a float32
			b int32
			c float32
			d int32
		}
		var ReturnMixed2 func(a float32, b int32, c float32, d int32) Mixed2
		purego.RegisterLibFunc(&ReturnMixed2, lib, "ReturnMixed2")
		expected := Mixed2{1, 2, 3, 4}
		if ret := ReturnMixed2(1, 2, 3, 4); ret != expected {
			t.Fatalf("ReturnMixed2 returned %+v wanted %+v", ret, expected)
		}
	}
	{
		type Mixed3 struct {
			a float32
			b uint32
			c float64
		}
		var ReturnMixed3 func(a float32, b uint32, c float64) Mixed3
		purego.RegisterLibFunc(&ReturnMixed3, lib, "ReturnMixed3")
		expected := Mixed3{1, 2, 3}
		if ret := ReturnMixed3(1, 2, 3); ret != expected {
			t.Fatalf("ReturnMixed3 returned %+v wanted %+v", ret, expected)
		}
	}
	{
		type Mixed4 struct {
			a float64
			b uint32
			c float32
		}
		var ReturnMixed4 func(a float64, b uint32, c float32) Mixed4
		purego.RegisterLibFunc(&ReturnMixed4, lib, "ReturnMixed4")
		expected := Mixed4{1, 2, 3}
		if ret := ReturnMixed4(1, 2, 3); ret != expected {
			t.Fatalf("ReturnMixed4 returned %+v wanted %+v", ret, expected)
		}
	}
	{
		type Ptr1 struct {
			a *int64
			b unsafe.Pointer
		}
		var ReturnPtr1 func(a *int64, b unsafe.Pointer) Ptr1
		purego.RegisterLibFunc(&ReturnPtr1, lib, "ReturnPtr1")
		a, b := new(int64), new(struct{})
		expected := Ptr1{a, unsafe.Pointer(b)}
		if ret := ReturnPtr1(a, unsafe.Pointer(b)); ret != expected {
			t.Fatalf("ReturnPtr1 returned %+v wanted %+v", ret, expected)
		}
		runtime.KeepAlive(a)
		runtime.KeepAlive(b)
	}
}
