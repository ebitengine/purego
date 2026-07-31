// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 The Ebitengine Authors

//go:build windows && 386

package purego_test

import (
	"fmt"
	"math"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/ebitengine/purego"
	"github.com/ebitengine/purego/internal/load"
)

const (
	win386Uint64 = uint64(0xfedcba9876543210)
	win386Int64  = int64(-0x123456789abcdef)
)

func openWin386ABILibrary(t *testing.T) uintptr {
	t.Helper()
	name := filepath.Join(t.TempDir(), "abitest.dll")
	if err := buildSharedLib(t, "CC", name, filepath.Join("testdata", "abitest", "abi_test.c")); err != nil {
		t.Fatal(err)
	}
	library, err := load.OpenLibrary(name)
	if err != nil {
		t.Fatalf("OpenLibrary(%q): %v", name, err)
	}
	t.Cleanup(func() {
		if err := load.CloseLibrary(library); err != nil {
			t.Errorf("CloseLibrary(%q): %v", name, err)
		}
	})
	return library
}

func TestWindows386RegisterFuncABI(t *testing.T) {
	library := openWin386ABILibrary(t)

	var pointer func() uintptr
	purego.RegisterLibFunc(&pointer, library, "win386_pointer")

	var mixed func(int8, uint16, int32, uint64, float32, float64, uintptr, bool, string) uint64
	purego.RegisterLibFunc(&mixed, library, "win386_direct_mixed")
	if got := mixed(-101, 60000, -2000000000, win386Uint64, 1.25, -2.5, pointer(), true, "purego-win386"); got != win386Uint64 {
		t.Fatalf("mixed arguments: got %#x, want %#x", got, win386Uint64)
	}

	var returnInt64 func() int64
	var returnUint64 func() uint64
	var returnFloat32 func(float32) float32
	var returnFloat64 func(float64) float64
	purego.RegisterLibFunc(&returnInt64, library, "win386_return_int64")
	purego.RegisterLibFunc(&returnUint64, library, "win386_return_uint64")
	purego.RegisterLibFunc(&returnFloat32, library, "win386_return_float32")
	purego.RegisterLibFunc(&returnFloat64, library, "win386_return_float64")

	if got := returnInt64(); got != win386Int64 {
		t.Errorf("int64 return: got %#x, want %#x", got, win386Int64)
	}
	if got := returnUint64(); got != win386Uint64 {
		t.Errorf("uint64 return: got %#x, want %#x", got, win386Uint64)
	}
	if got := returnFloat32(2.5); got != 3.75 {
		t.Errorf("float32 return: got %v, want 3.75", got)
	}
	if got := returnFloat64(2.5); got != -5.625 {
		t.Errorf("float64 return: got %v, want -5.625", got)
	}

	var sixteenUint64 func(
		uint64, uint64, uint64, uint64, uint64, uint64, uint64, uint64,
		uint64, uint64, uint64, uint64, uint64, uint64, uint64, uint64,
	) uint64
	var sixteenFloat64 func(
		float64, float64, float64, float64, float64, float64, float64, float64,
		float64, float64, float64, float64, float64, float64, float64, float64,
	) float64
	purego.RegisterLibFunc(&sixteenUint64, library, "win386_16_uint64")
	purego.RegisterLibFunc(&sixteenFloat64, library, "win386_16_float64")
	if got := sixteenUint64(
		0x100000001, 0x100000002, 0x100000003, 0x100000004,
		0x100000005, 0x100000006, 0x100000007, 0x100000008,
		0x100000009, 0x10000000a, 0x10000000b, 0x10000000c,
		0x10000000d, 0x10000000e, 0x10000000f, 0x100000010,
	); got != 0x1000000088 {
		t.Errorf("16 uint64 arguments: got %#x, want %#x", got, uint64(0x1000000088))
	}
	if got := sixteenFloat64(1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16); got != 136 {
		t.Errorf("16 float64 arguments: got %v, want 136", got)
	}
}

func TestWindows386Callbacks(t *testing.T) {
	library := openWin386ABILibrary(t)

	checkArgs := func(i8 int8, u16 uint16, i32 int32, u64 uint64, f32 float32, f64 float64, pointer *uint32, boolean bool) {
		t.Helper()
		if i8 != -101 || u16 != 60000 || i32 != -2000000000 || u64 != win386Uint64 ||
			f32 != 1.25 || f64 != -2.5 || pointer == nil || *pointer != 0x89abcdef || !boolean {
			t.Errorf("callback arguments were decoded incorrectly: %d %d %d %#x %v %v %p %v", i8, u16, i32, u64, f32, f64, pointer, boolean)
		}
	}

	stdcall := purego.NewCallback(func(i8 int8, u16 uint16, i32 int32, u64 uint64, f32 float32, f64 float64, pointer *uint32, boolean bool) uint64 {
		checkArgs(i8, u16, i32, u64, f32, f64, pointer, boolean)
		return win386Uint64
	})
	cdecl := purego.NewCallback(func(_ purego.CDecl, i8 int8, u16 uint16, i32 int32, u64 uint64, f32 float32, f64 float64, pointer *uint32, boolean bool) uint64 {
		checkArgs(i8, u16, i32, u64, f32, f64, pointer, boolean)
		return win386Uint64
	})

	var callStdcall func(uintptr) uint64
	var callCdecl func(uintptr) uint64
	purego.RegisterLibFunc(&callStdcall, library, "win386_call_stdcall_u64")
	purego.RegisterLibFunc(&callCdecl, library, "win386_call_cdecl_u64")
	for i := 0; i < 100; i++ {
		if got := callStdcall(stdcall); got != win386Uint64 {
			t.Fatalf("stdcall uint64 return %d: got %#x, want %#x", i, got, win386Uint64)
		}
		if got := callCdecl(cdecl); got != win386Uint64 {
			t.Fatalf("cdecl uint64 return %d: got %#x, want %#x", i, got, win386Uint64)
		}
	}

	stdcallInt64 := purego.NewCallback(func(value int64, f64 float64) int64 {
		if value != win386Int64 || f64 != 3.5 {
			t.Errorf("signed callback arguments: got %#x, %v", value, f64)
		}
		return value
	})
	stdcallFloat32 := purego.NewCallback(func(u64 uint64, f32 float32, f64 float64) float32 {
		if u64 != 0x123456789abcdef0 || f32 != 1.25 || f64 != -2.5 {
			t.Errorf("float32 callback arguments: got %#x, %v, %v", u64, f32, f64)
		}
		return 6.25
	})
	cdeclFloat64 := purego.NewCallback(func(_ purego.CDecl, u64 uint64, f32 float32, f64 float64) float64 {
		if u64 != 0x123456789abcdef0 || f32 != 1.25 || f64 != -2.5 {
			t.Errorf("float64 callback arguments: got %#x, %v, %v", u64, f32, f64)
		}
		return -9.5
	})

	var callInt64 func(uintptr) int64
	var callFloat32 func(uintptr) float32
	var callFloat64 func(uintptr) float64
	purego.RegisterLibFunc(&callInt64, library, "win386_call_stdcall_i64")
	purego.RegisterLibFunc(&callFloat32, library, "win386_call_stdcall_f32")
	purego.RegisterLibFunc(&callFloat64, library, "win386_call_cdecl_f64")
	if got := callInt64(stdcallInt64); got != win386Int64 {
		t.Errorf("callback int64 return: got %#x, want %#x", got, win386Int64)
	}
	if got := callFloat32(stdcallFloat32); got != 6.25 {
		t.Errorf("callback float32 return: got %v, want 6.25", got)
	}
	if got := callFloat64(cdeclFloat64); got != -9.5 {
		t.Errorf("callback float64 return: got %v, want -9.5", got)
	}

	stdcall16Uint64 := purego.NewCallback(func(
		a1, a2, a3, a4, a5, a6, a7, a8,
		a9, a10, a11, a12, a13, a14, a15, a16 uint64,
	) uint64 {
		return a1 + a2 + a3 + a4 + a5 + a6 + a7 + a8 +
			a9 + a10 + a11 + a12 + a13 + a14 + a15 + a16
	})
	var call16Uint64 func(uintptr) uint64
	purego.RegisterLibFunc(&call16Uint64, library, "win386_call_stdcall_16_u64")
	if got := call16Uint64(stdcall16Uint64); got != 0x1000000088 {
		t.Errorf("callback 16 uint64 arguments: got %#x, want %#x", got, uint64(0x1000000088))
	}
}

func TestWindows386SlotLimit(t *testing.T) {
	t.Run("direct", func(t *testing.T) {
		var sixteenInt64 func(
			int64, int64, int64, int64, int64, int64, int64, int64,
			int64, int64, int64, int64, int64, int64, int64, int64,
		)
		purego.RegisterFunc(&sixteenInt64, 1)

		mustPanicWin386(t, "too many stack arguments", func() {
			var seventeenInt64 func(
				int64, int64, int64, int64, int64, int64, int64, int64,
				int64, int64, int64, int64, int64, int64, int64, int64,
				int64,
			)
			purego.RegisterFunc(&seventeenInt64, 1)
		})
	})

	t.Run("callback", func(t *testing.T) {
		purego.NewCallback(func(
			uint64, uint64, uint64, uint64, uint64, uint64, uint64, uint64,
			uint64, uint64, uint64, uint64, uint64, uint64, uint64, uint64,
		) {
		})

		mustPanicWin386(t, "too many callback argument slots", func() {
			purego.NewCallback(func(
				uint64, uint64, uint64, uint64, uint64, uint64, uint64, uint64,
				uint64, uint64, uint64, uint64, uint64, uint64, uint64, uint64,
				uint64,
			) {
			})
		})
		mustPanicWin386(t, "unsupported argument type: struct", func() {
			purego.NewCallback(func(struct{ Value uint32 }) {})
		})
		mustPanicWin386(t, "function must not be nil", func() {
			var nilCallback func(uint32) uint32
			purego.NewCallback(nilCallback)
		})
	})
}

func mustPanicWin386(t *testing.T, contains string, fn func()) {
	t.Helper()
	defer func() {
		panicValue := recover()
		if panicValue == nil || !strings.Contains(fmt.Sprint(panicValue), contains) {
			t.Fatalf("got panic %v, want text containing %q", panicValue, contains)
		}
	}()
	fn()
}

func TestWindows386FullSlotCalls(t *testing.T) {
	library := openWin386ABILibrary(t)

	var thirtyTwo func(
		uintptr, uintptr, uintptr, uintptr, uintptr, uintptr, uintptr, uintptr,
		uintptr, uintptr, uintptr, uintptr, uintptr, uintptr, uintptr, uintptr,
		uintptr, uintptr, uintptr, uintptr, uintptr, uintptr, uintptr, uintptr,
		uintptr, uintptr, uintptr, uintptr, uintptr, uintptr, uintptr, uintptr,
	) uintptr
	purego.RegisterLibFunc(&thirtyTwo, library, "stack_32_uintptr")
	if got := thirtyTwo(
		1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16,
		17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32,
	); got != 528 {
		t.Fatalf("32 scalar slots: got %d, want 528", got)
	}

	type slots32 struct{ Slots [32]uint32 }
	var takeStruct func(slots32) uint32
	purego.RegisterLibFunc(&takeStruct, library, "win386_take_struct_32_slots")
	var value slots32
	for i := range value.Slots {
		value.Slots[i] = uint32(i + 1)
	}
	if got := takeStruct(value); got != 528 {
		t.Fatalf("32-slot struct: got %d, want 528", got)
	}

	mustPanicWin386(t, "too many stack arguments", func() {
		type slots33 struct{ Slots [33]uint32 }
		var tooLarge func(slots33)
		purego.RegisterFunc(&tooLarge, 1)
	})
}

func TestWindows386X87SpecialValues(t *testing.T) {
	library := openWin386ABILibrary(t)

	var identity32 func(float32) float32
	var identity64 func(float64) float64
	var integer func() uint64
	purego.RegisterLibFunc(&identity32, library, "win386_identity_float32")
	purego.RegisterLibFunc(&identity64, library, "win386_identity_float64")
	purego.RegisterLibFunc(&integer, library, "win386_return_uint64")

	callback32 := purego.NewCallback(func(value float32) float32 { return value })
	callback64 := purego.NewCallback(func(_ purego.CDecl, value float64) float64 { return value })
	var callCallback32 func(uintptr, float32) float32
	var callCallback64 func(uintptr, float64) float64
	purego.RegisterLibFunc(&callCallback32, library, "win386_call_stdcall_identity_f32")
	purego.RegisterLibFunc(&callCallback64, library, "win386_call_cdecl_identity_f64")

	float32Values := []float32{
		float32(math.Copysign(0, -1)),
		float32(math.Inf(1)),
		float32(math.Inf(-1)),
		math.SmallestNonzeroFloat32,
		math.Float32frombits(0x7fc12345),
	}
	float64Values := []float64{
		math.Copysign(0, -1),
		math.Inf(1),
		math.Inf(-1),
		math.SmallestNonzeroFloat64,
		math.Float64frombits(0x7ff8123456789abc),
	}
	for i := range 2000 {
		value32 := float32Values[i%len(float32Values)]
		checkFloat32Win386(t, "direct", value32, identity32(value32))
		checkFloat32Win386(t, "callback", value32, callCallback32(callback32, value32))

		value64 := float64Values[i%len(float64Values)]
		checkFloat64Win386(t, "direct", value64, identity64(value64))
		checkFloat64Win386(t, "callback", value64, callCallback64(callback64, value64))
		if got := integer(); got != win386Uint64 {
			t.Fatalf("integer call after x87 call %d: got %#x", i, got)
		}
	}

	type oneFloat32 struct{ Value float32 }
	type oneFloat64 struct{ Value float64 }
	var struct32 func(oneFloat32) oneFloat32
	var struct64 func(oneFloat64) oneFloat64
	purego.RegisterLibFunc(&struct32, library, "win386_identity_struct_float32")
	purego.RegisterLibFunc(&struct64, library, "win386_identity_struct_float64")
	for _, value := range float32Values {
		checkFloat32Win386(t, "struct", value, struct32(oneFloat32{value}).Value)
	}
	for _, value := range float64Values {
		checkFloat64Win386(t, "struct", value, struct64(oneFloat64{value}).Value)
	}
}

func checkFloat32Win386(t *testing.T, path string, want, got float32) {
	t.Helper()
	if math.IsNaN(float64(want)) {
		if !math.IsNaN(float64(got)) {
			t.Fatalf("%s float32: got %v, want NaN", path, got)
		}
		return
	}
	if math.Float32bits(got) != math.Float32bits(want) {
		t.Fatalf("%s float32: got %#08x, want %#08x", path, math.Float32bits(got), math.Float32bits(want))
	}
}

func checkFloat64Win386(t *testing.T, path string, want, got float64) {
	t.Helper()
	if math.IsNaN(want) {
		if !math.IsNaN(got) {
			t.Fatalf("%s float64: got %v, want NaN", path, got)
		}
		return
	}
	if math.Float64bits(got) != math.Float64bits(want) {
		t.Fatalf("%s float64: got %#016x, want %#016x", path, math.Float64bits(got), math.Float64bits(want))
	}
}

func TestWindows386CallbackConcurrencyAndReentrancy(t *testing.T) {
	library := openWin386ABILibrary(t)

	var callOne func(uintptr, uint32) uint32
	purego.RegisterLibFunc(&callOne, library, "win386_call_stdcall_u32")
	var calls atomic.Uint64
	callback := purego.NewCallback(func(value uint32) uint32 {
		calls.Add(1)
		return value ^ 0xa5a5a5a5
	})
	var failed atomic.Bool
	var wg sync.WaitGroup
	for goroutine := range 8 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for iteration := range 250 {
				value := uint32(goroutine<<16 | iteration)
				if got := callOne(callback, value); got != value^0xa5a5a5a5 {
					failed.Store(true)
					return
				}
			}
		}()
	}
	wg.Wait()
	if failed.Load() || calls.Load() != 2000 {
		t.Fatalf("concurrent callback calls: failed=%v calls=%d, want 2000", failed.Load(), calls.Load())
	}

	var callStdcall func(uintptr, uint32) uint64
	var stdcall uintptr
	purego.RegisterLibFunc(&callStdcall, library, "win386_call_stdcall_reentrant")
	stdcall = purego.NewCallback(func(depth uint32) uint64 {
		if depth == 0 {
			return 1
		}
		return uint64(depth) + callStdcall(stdcall, depth-1)
	})
	if got, want := callStdcall(stdcall, 12), uint64(79); got != want {
		t.Fatalf("stdcall reentrant callback: got %d, want %d", got, want)
	}

	var callCdecl func(uintptr, uint32) uint64
	var cdecl uintptr
	purego.RegisterLibFunc(&callCdecl, library, "win386_call_cdecl_reentrant")
	cdecl = purego.NewCallback(func(_ purego.CDecl, depth uint32) uint64 {
		if depth == 0 {
			return 1
		}
		return uint64(depth) + callCdecl(cdecl, depth-1)
	})
	if got, want := callCdecl(cdecl, 12), uint64(79); got != want {
		t.Fatalf("cdecl reentrant callback: got %d, want %d", got, want)
	}
}

func TestWindows386CallbackFromNativeThread(t *testing.T) {
	library := openWin386ABILibrary(t)

	var badArgs atomic.Bool
	callback := purego.NewCallback(func(value uint64, f32 float32, f64 float64) uint64 {
		if value != win386Uint64 || f32 != 1.25 || f64 != -2.5 {
			badArgs.Store(true)
			return 0
		}
		return value
	})
	var callOnThread func(uintptr) uint64
	purego.RegisterLibFunc(&callOnThread, library, "win386_call_stdcall_on_thread")
	if got := callOnThread(callback); got != win386Uint64 || badArgs.Load() {
		t.Fatalf("native-thread callback: got %#x badArgs=%v", got, badArgs.Load())
	}
}

func TestWindows386CallbackTableBeyond32(t *testing.T) {
	library := openWin386ABILibrary(t)
	var last uintptr
	for i := range 40 {
		add := uint32(i)
		last = purego.NewCallback(func(value uint32) uint32 { return value + add })
	}
	var call func(uintptr, uint32) uint32
	purego.RegisterLibFunc(&call, library, "win386_call_stdcall_u32")
	if got := call(last, 100); got != 139 {
		t.Fatalf("callback table entry beyond 32: got %d, want 139", got)
	}
}
