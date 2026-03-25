// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2023 The Ebitengine Authors

package purego_test

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"unsafe"

	"github.com/ebitengine/purego"
	"github.com/ebitengine/purego/internal/load"
)

func getSystemLibrary() (string, error) {
	switch runtime.GOOS {
	case "darwin":
		return "/usr/lib/libSystem.B.dylib", nil
	case "freebsd":
		return "libc.so.7", nil
	case "linux":
		return "libc.so.6", nil
	case "netbsd":
		return "libc.so", nil
	case "windows":
		return "ucrtbase.dll", nil
	default:
		return "", fmt.Errorf("GOOS=%s is not supported", runtime.GOOS)
	}
}

func TestRegisterFunc(t *testing.T) {
	library, err := getSystemLibrary()
	if err != nil {
		t.Fatalf("couldn't get system library: %s", err)
	}
	libc, err := load.OpenLibrary(library)
	if err != nil {
		t.Fatalf("failed to dlopen: %s", err)
	}
	var puts func(string)
	purego.RegisterLibFunc(&puts, libc, "puts")
	puts("Calling C from from Go without Cgo!")
}

func Test_qsort(t *testing.T) {
	if runtime.GOARCH != "arm" && runtime.GOARCH != "arm64" && runtime.GOARCH != "386" && runtime.GOARCH != "amd64" && runtime.GOARCH != "loong64" && runtime.GOARCH != "ppc64le" && runtime.GOARCH != "riscv64" && runtime.GOARCH != "s390x" {
		t.Skip("Platform doesn't support Floats")
		return
	}
	library, err := getSystemLibrary()
	if err != nil {
		t.Fatalf("couldn't get system library: %s", err)
	}
	libc, err := load.OpenLibrary(library)
	if err != nil {
		t.Fatalf("failed to dlopen: %s", err)
	}

	data := []int{88, 56, 100, 2, 25}
	sorted := []int{2, 25, 56, 88, 100}
	compare := func(_ purego.CDecl, a, b *int) int {
		return *a - *b
	}
	var qsort func(data []int, nitms uintptr, size uintptr, compar func(_ purego.CDecl, a, b *int) int)
	purego.RegisterLibFunc(&qsort, libc, "qsort")
	qsort(data, uintptr(len(data)), unsafe.Sizeof(int(0)), compare)
	for i := range data {
		if data[i] != sorted[i] {
			t.Errorf("got %d wanted %d at %d", data[i], sorted[i], i)
		}
	}
}

func TestRegisterFunc_Floats(t *testing.T) {
	if runtime.GOARCH != "arm" && runtime.GOARCH != "arm64" && runtime.GOARCH != "386" && runtime.GOARCH != "amd64" && runtime.GOARCH != "loong64" && runtime.GOARCH != "ppc64le" && runtime.GOARCH != "riscv64" && runtime.GOARCH != "s390x" {
		t.Skip("Platform doesn't support Floats")
		return
	}
	if runtime.GOOS == "windows" && runtime.GOARCH == "386" {
		t.Skip("need a 32bit gcc to run this test") // TODO: find 32bit gcc for test
	}
	library, err := getSystemLibrary()
	if err != nil {
		t.Fatalf("couldn't get system library: %s", err)
	}
	libc, err := load.OpenLibrary(library)
	if err != nil {
		t.Fatalf("failed to dlopen: %s", err)
	}
	{
		var strtof func(arg string) float32
		purego.RegisterLibFunc(&strtof, libc, "strtof")
		const (
			arg = "2"
		)
		got := strtof(arg)
		expected := float32(2)
		if got != expected {
			t.Errorf("strtof failed. got %f but wanted %f", got, expected)
		}
	}
	{
		var strtod func(arg string, ptr **byte) float64
		purego.RegisterLibFunc(&strtod, libc, "strtod")
		const (
			arg = "1"
		)
		got := strtod(arg, nil)
		expected := float64(1)
		if got != expected {
			t.Errorf("strtod failed. got %f but wanted %f", got, expected)
		}
	}
}

func TestRegisterLibFunc_Bool(t *testing.T) {
	if runtime.GOARCH != "arm" && runtime.GOARCH != "arm64" && runtime.GOARCH != "386" && runtime.GOARCH != "amd64" && runtime.GOARCH != "loong64" && runtime.GOARCH != "ppc64le" && runtime.GOARCH != "riscv64" && runtime.GOARCH != "s390x" {
		t.Skip("Platform doesn't support callbacks")
		return
	}
	// this callback recreates the state where the return register
	// contains other information but the least significant byte is false
	cbFalse := purego.NewCallback(func() uintptr {
		x := uint64(0x7F5948AE9A00)
		return uintptr(x)
	})
	var runFalse func() bool
	purego.RegisterFunc(&runFalse, cbFalse)
	expected := false
	if got := runFalse(); got != expected {
		t.Errorf("runFalse failed. got %t but wanted %t", got, expected)
	}
}

func TestRegisterFunc_FastPath(t *testing.T) {
	if runtime.GOOS == "windows" || (runtime.GOARCH != "amd64" && runtime.GOARCH != "arm64") {
		t.Skip("fast path only enabled on non-Windows amd64 and arm64")
	}

	lib := openBenchmarkLibrary(t)
	sym := openBenchmarkSymbol(t, lib, "sum5_c")

	var fn func(int64, int64, int64, int64, int64) int64
	purego.RegisterFunc(&fn, sym)

	if got, want := fn(1, 2, 3, 4, 5), int64(15); got != want {
		t.Fatalf("got %d, want %d", got, want)
	}

	allocs := testing.AllocsPerRun(1000, func() {
		if got := fn(1, 2, 3, 4, 5); got != 15 {
			panic(fmt.Sprintf("got %d, want 15", got))
		}
	})
	if allocs != 0 {
		t.Fatalf("allocs per run = %v, want 0", allocs)
	}
}

func TestRegisterFunc_FastPathFloat(t *testing.T) {
	if runtime.GOOS == "windows" || (runtime.GOARCH != "amd64" && runtime.GOARCH != "arm64") {
		t.Skip("fast path only enabled on non-Windows amd64 and arm64")
	}

	lib := openBenchmarkLibrary(t)

	t.Run("3ints_1float32", func(t *testing.T) {
		sym := openBenchmarkSymbol(t, lib, "weighted_sum3f_c")
		var fn func(int64, int64, int64, float32) int64
		purego.RegisterFunc(&fn, sym)
		if got, want := fn(1, 2, 3, 2.0), int64(12); got != want {
			t.Fatalf("got %d, want %d", got, want)
		}
		allocs := testing.AllocsPerRun(1000, func() {
			fn(1, 2, 3, 2.0)
		})
		if allocs != 0 {
			t.Fatalf("allocs per run = %v, want 0", allocs)
		}
	})

	t.Run("5ints_1float32", func(t *testing.T) {
		sym := openBenchmarkSymbol(t, lib, "weighted_sum5f_c")
		var fn func(int64, int64, int64, int64, int64, float32) int64
		purego.RegisterFunc(&fn, sym)
		if got, want := fn(1, 2, 3, 4, 5, 3.0), int64(45); got != want {
			t.Fatalf("got %d, want %d", got, want)
		}
		allocs := testing.AllocsPerRun(1000, func() {
			fn(1, 2, 3, 4, 5, 3.0)
		})
		if allocs != 0 {
			t.Fatalf("allocs per run = %v, want 0", allocs)
		}
	})

	t.Run("3ints_1float64", func(t *testing.T) {
		sym := openBenchmarkSymbol(t, lib, "weighted_sum3d_c")
		var fn func(int64, int64, int64, float64) int64
		purego.RegisterFunc(&fn, sym)
		if got, want := fn(1, 2, 3, 2.0), int64(12); got != want {
			t.Fatalf("got %d, want %d", got, want)
		}
		allocs := testing.AllocsPerRun(1000, func() {
			fn(1, 2, 3, 2.0)
		})
		if allocs != 0 {
			t.Fatalf("allocs per run = %v, want 0", allocs)
		}
	})
}

func TestRegisterFunc_FastPathInterleavedFloat(t *testing.T) {
	if runtime.GOOS == "windows" || (runtime.GOARCH != "amd64" && runtime.GOARCH != "arm64") {
		t.Skip("fast path only enabled on non-Windows amd64 and arm64")
	}

	lib := openBenchmarkLibrary(t)

	t.Run("int_float_int_int", func(t *testing.T) {
		sym := openBenchmarkSymbol(t, lib, "interleaved_if_c")
		var fn func(int64, float32, int64, int64) int64
		purego.RegisterFunc(&fn, sym)
		if got, want := fn(1, 2.0, 2, 3), int64(12); got != want {
			t.Fatalf("got %d, want %d", got, want)
		}
	})

	t.Run("int_float_int_float_int", func(t *testing.T) {
		sym := openBenchmarkSymbol(t, lib, "interleaved_2f_c")
		var fn func(int64, float32, int64, float32, int64) int64
		purego.RegisterFunc(&fn, sym)
		if got, want := fn(10, 2.0, 20, 3.0, 5), int64(85); got != want {
			t.Fatalf("got %d, want %d", got, want)
		}
	})
}

func TestRegisterFunc_FastPathInterleavedFloat32x1(t *testing.T) {
	if runtime.GOOS == "windows" || (runtime.GOARCH != "amd64" && runtime.GOARCH != "arm64") {
		t.Skip("fast path only enabled on non-Windows amd64 and arm64")
	}

	lib := openBenchmarkLibrary(t)

	t.Run("5args_float_at_3", func(t *testing.T) {
		sym := openBenchmarkSymbol(t, lib, "rmsnorm_shape_c")
		var fn func(uintptr, uintptr, uintptr, float32, uintptr) int32
		purego.RegisterFunc(&fn, sym)
		got := fn(10, 20, 30, 1.0, 40)
		want := int32(10 + 20 + 30 + 1 + 40)
		if got != want {
			t.Fatalf("got %d, want %d", got, want)
		}
		allocs := testing.AllocsPerRun(1000, func() {
			fn(10, 20, 30, 1.0, 40)
		})
		if allocs != 0 {
			t.Fatalf("allocs per run = %v, want 0", allocs)
		}
	})

	t.Run("9args_float_at_4", func(t *testing.T) {
		sym := openBenchmarkSymbol(t, lib, "sdpa_shape_c")
		var fn func(uintptr, uintptr, uintptr, uintptr, float32, uintptr, uintptr, uintptr, uintptr) int32
		purego.RegisterFunc(&fn, sym)
		got := fn(1, 2, 3, 4, 1.0, 5, 6, 7, 8)
		want := int32(1 + 2 + 3 + 4 + 1 + 5 + 6 + 7 + 8)
		if got != want {
			t.Fatalf("got %d, want %d", got, want)
		}
		allocs := testing.AllocsPerRun(1000, func() {
			fn(1, 2, 3, 4, 1.0, 5, 6, 7, 8)
		})
		if allocs != 0 {
			t.Fatalf("allocs per run = %v, want 0", allocs)
		}
	})

	t.Run("3args_float_at_0", func(t *testing.T) {
		sym := openBenchmarkSymbol(t, lib, "interleaved_3_f0_c")
		var fn func(float32, int64, int64) int64
		purego.RegisterFunc(&fn, sym)
		got := fn(2.0, 10, 20)
		want := int64(60)
		if got != want {
			t.Fatalf("got %d, want %d", got, want)
		}
		allocs := testing.AllocsPerRun(1000, func() {
			fn(2.0, 10, 20)
		})
		if allocs != 0 {
			t.Fatalf("allocs per run = %v, want 0", allocs)
		}
	})
}

func TestRegisterFunc_FastPathBoolReturn(t *testing.T) {
	if runtime.GOOS == "windows" || (runtime.GOARCH != "amd64" && runtime.GOARCH != "arm64") {
		t.Skip("fast path only enabled on non-Windows amd64 and arm64")
	}

	cb := purego.NewCallback(func(v uintptr) bool {
		return v == 42
	})

	var fn func(uintptr) bool
	purego.RegisterFunc(&fn, cb)

	if !fn(42) {
		t.Fatal("fn(42) = false, want true")
	}
	if fn(0) {
		t.Fatal("fn(0) = true, want false")
	}
}

func TestABI(t *testing.T) {
	if runtime.GOOS == "windows" && runtime.GOARCH == "386" {
		t.Skip("need a 32bit gcc to run this test") // TODO: find 32bit gcc for test
	}
	libFileName := filepath.Join(t.TempDir(), "abitest.so")
	t.Logf("Build %v", libFileName)

	if err := buildSharedLib("CC", libFileName, filepath.Join("testdata", "abitest", "abi_test.c")); err != nil {
		t.Fatal(err)
	}

	lib, err := load.OpenLibrary(libFileName)
	if err != nil {
		t.Fatalf("Dlopen(%q) failed: %v", libFileName, err)
	}
	defer func() {
		if err := load.CloseLibrary(lib); err != nil {
			t.Fatalf("failed to close library: %s", err)
		}
	}()
	{
		const cName = "stack_uint8_t"
		const expect = 2047
		var fn func(a, b, c, d, e, f, g, h uint32, i, j uint8, k uint32) uint32
		purego.RegisterLibFunc(&fn, lib, cName)
		res := fn(256, 512, 4, 8, 16, 32, 64, 128, 1, 2, 1024)
		if res != expect {
			t.Fatalf("%s: got %d, want %d", cName, res, expect)
		}
	}
	{
		const cName = "reg_uint8_t"
		const expect = 1027
		var fn func(a, b uint8, c uint32) uint32
		purego.RegisterLibFunc(&fn, lib, cName)
		res := fn(1, 2, 1024)
		if res != expect {
			t.Fatalf("%s: got %d, want %d", cName, res, expect)
		}
	}
	{
		const cName = "stack_string"
		const expect = 255
		var fn func(a, b, c, d, e, f, g, h uint32, i string) uint32
		purego.RegisterLibFunc(&fn, lib, cName)
		res := fn(1, 2, 4, 8, 16, 32, 64, 128, "test")
		if res != expect {
			t.Fatalf("%s: got %d, want %d", cName, res, expect)
		}
	}
	{
		const cName = "stack_8i32_3strings"
		var fn func(*byte, uintptr, int32, int32, int32, int32, int32, int32, int32, int32, string, string, string)
		purego.RegisterLibFunc(&fn, lib, cName)
		buf := make([]byte, 256)
		fn(&buf[0], uintptr(len(buf)), 1, 2, 3, 4, 5, 6, 7, 8, "foo", "bar", "baz")
		res := string(buf[:bytes.IndexByte(buf, 0)])
		const want = "1:2:3:4:5:6:7:8:foo:bar:baz"
		if res != want {
			t.Fatalf("%s: got %q, want %q", cName, res, want)
		}
	}
}

func TestABI_ArgumentPassing(t *testing.T) {
	if runtime.GOOS == "windows" && runtime.GOARCH == "386" {
		t.Skip("need a 32bit gcc to run this test") // TODO: find 32bit gcc for test
	}
	libFileName := filepath.Join(t.TempDir(), "abitest.so")
	if err := buildSharedLib("CC", libFileName, filepath.Join("testdata", "abitest", "abi_test.c")); err != nil {
		t.Fatal(err)
	}
	lib, err := load.OpenLibrary(libFileName)
	if err != nil {
		t.Fatalf("Failed to open library %q: %v", libFileName, err)
	}
	t.Cleanup(func() {
		if err := load.CloseLibrary(lib); err != nil {
			t.Errorf("Failed to close library: %v", err)
		}
	})

	tests := []struct {
		name string
		fn   any
		cFn  string
		call func(any) string
		want string
	}{
		{
			name: "10_int32_baseline",
			fn:   new(func(*byte, uintptr, int32, int32, int32, int32, int32, int32, int32, int32, int32, int32)),
			cFn:  "stack_10_int32",
			call: func(f any) string {
				buf := make([]byte, 256)
				(*f.(*func(*byte, uintptr, int32, int32, int32, int32, int32, int32, int32, int32, int32, int32)))(&buf[0], 256, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10)
				return string(buf[:bytes.IndexByte(buf, 0)])
			},
			want: "1:2:3:4:5:6:7:8:9:10",
		},
		{
			name: "11_int32",
			fn:   new(func(*byte, uintptr, int32, int32, int32, int32, int32, int32, int32, int32, int32, int32, int32)),
			cFn:  "stack_11_int32",
			call: func(f any) string {
				buf := make([]byte, 256)
				(*f.(*func(*byte, uintptr, int32, int32, int32, int32, int32, int32, int32, int32, int32, int32, int32)))(&buf[0], 256, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11)
				return string(buf[:bytes.IndexByte(buf, 0)])
			},
			want: "1:2:3:4:5:6:7:8:9:10:11",
		},
		{
			name: "10_float32",
			fn:   new(func(*byte, uintptr, float32, float32, float32, float32, float32, float32, float32, float32, float32, float32)),
			cFn:  "stack_10_float32",
			call: func(f any) string {
				buf := make([]byte, 256)
				(*f.(*func(*byte, uintptr, float32, float32, float32, float32, float32, float32, float32, float32, float32, float32)))(&buf[0], 256, 1.0, 2.0, 3.0, 4.0, 5.0, 6.0, 7.0, 8.0, 9.0, 10.0)
				return string(buf[:bytes.IndexByte(buf, 0)])
			},
			want: "1.0:2.0:3.0:4.0:5.0:6.0:7.0:8.0:9.0:10.0",
		},
		{
			name: "mixed_stack_strings",
			fn:   new(func(*byte, uintptr, int32, int32, int32, int32, int32, int32, int32, int32, string, bool, int32, string)),
			cFn:  "stack_mixed_stack_4args",
			call: func(f any) string {
				buf := make([]byte, 256)
				(*f.(*func(*byte, uintptr, int32, int32, int32, int32, int32, int32, int32, int32, string, bool, int32, string)))(&buf[0], 256, 1, 2, 3, 4, 5, 6, 7, 8, "foo", false, 99, "bar")
				return string(buf[:bytes.IndexByte(buf, 0)])
			},
			want: "1:2:3:4:5:6:7:8:foo:0:99:bar",
		},
		{
			name: "20_int32",
			fn:   new(func(*byte, uintptr, int32, int32, int32, int32, int32, int32, int32, int32, int32, int32, int32, int32, int32, int32, int32, int32, int32, int32, int32, int32)),
			cFn:  "stack_20_int32",
			call: func(f any) string {
				buf := make([]byte, 512)
				(*f.(*func(*byte, uintptr, int32, int32, int32, int32, int32, int32, int32, int32, int32, int32, int32, int32, int32, int32, int32, int32, int32, int32, int32, int32)))(&buf[0], 512, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20)
				return string(buf[:bytes.IndexByte(buf, 0)])
			},
			want: "1:2:3:4:5:6:7:8:9:10:11:12:13:14:15:16:17:18:19:20",
		},
		{
			name: "8int_hfa2_stack",
			fn:   new(func(*byte, uintptr, int32, int32, int32, int32, int32, int32, int32, int32, struct{ x, y float32 })),
			cFn:  "stack_8int_hfa2_stack",
			call: func(f any) string {
				buf := make([]byte, 256)
				(*f.(*func(*byte, uintptr, int32, int32, int32, int32, int32, int32, int32, int32, struct{ x, y float32 })))(&buf[0], 256, 1, 2, 3, 4, 5, 6, 7, 8, struct{ x, y float32 }{10.0, 20.0})
				return string(buf[:bytes.IndexByte(buf, 0)])
			},
			want: "1:2:3:4:5:6:7:8:10.0:20.0",
		},
		{
			name: "8int_2structs_stack",
			fn:   new(func(*byte, uintptr, int32, int32, int32, int32, int32, int32, int32, int32, struct{ x, y int32 }, struct{ x, y int32 })),
			cFn:  "stack_8int_2structs_stack",
			call: func(f any) string {
				buf := make([]byte, 256)
				(*f.(*func(*byte, uintptr, int32, int32, int32, int32, int32, int32, int32, int32, struct{ x, y int32 }, struct{ x, y int32 })))(&buf[0], 256, 1, 2, 3, 4, 5, 6, 7, 8, struct{ x, y int32 }{9, 10}, struct{ x, y int32 }{11, 12})
				return string(buf[:bytes.IndexByte(buf, 0)])
			},
			want: "1:2:3:4:5:6:7:8:9:10:11:12",
		},
		{
			name: "8float_hfa2_stack",
			fn:   new(func(*byte, uintptr, float32, float32, float32, float32, float32, float32, float32, float32, struct{ x, y float32 })),
			cFn:  "stack_8float_hfa2_stack",
			call: func(f any) string {
				buf := make([]byte, 256)
				(*f.(*func(*byte, uintptr, float32, float32, float32, float32, float32, float32, float32, float32, struct{ x, y float32 })))(&buf[0], 256, 1.0, 2.0, 3.0, 4.0, 5.0, 6.0, 7.0, 8.0, struct{ x, y float32 }{9.0, 10.0})
				return string(buf[:bytes.IndexByte(buf, 0)])
			},
			want: "1.0:2.0:3.0:4.0:5.0:6.0:7.0:8.0:9.0:10.0",
		},
		{
			name: "8int_hfa2_floatregs",
			fn:   new(func(*byte, uintptr, int32, int32, int32, int32, int32, int32, int32, int32, struct{ x, y float32 })),
			cFn:  "stack_8int_hfa2_floatregs",
			call: func(f any) string {
				buf := make([]byte, 256)
				(*f.(*func(*byte, uintptr, int32, int32, int32, int32, int32, int32, int32, int32, struct{ x, y float32 })))(&buf[0], 256, 1, 2, 3, 4, 5, 6, 7, 8, struct{ x, y float32 }{10.0, 20.0})
				return string(buf[:bytes.IndexByte(buf, 0)])
			},
			want: "1:2:3:4:5:6:7:8:10.0:20.0",
		},
		{
			name: "8int_int_struct_int",
			fn:   new(func(*byte, uintptr, int32, int32, int32, int32, int32, int32, int32, int32, int32, struct{ x, y int32 }, int32)),
			cFn:  "stack_8int_int_struct_int",
			call: func(f any) string {
				buf := make([]byte, 256)
				(*f.(*func(*byte, uintptr, int32, int32, int32, int32, int32, int32, int32, int32, int32, struct{ x, y int32 }, int32)))(&buf[0], 256, 1, 2, 3, 4, 5, 6, 7, 8, 9, struct{ x, y int32 }{10, 11}, 12)
				return string(buf[:bytes.IndexByte(buf, 0)])
			},
			want: "1:2:3:4:5:6:7:8:9:10:11:12",
		},
		{
			name: "8int_hfa4_stack",
			fn:   new(func(*byte, uintptr, int32, int32, int32, int32, int32, int32, int32, int32, struct{ x, y, z, w float32 })),
			cFn:  "stack_8int_hfa4_stack",
			call: func(f any) string {
				buf := make([]byte, 256)
				(*f.(*func(*byte, uintptr, int32, int32, int32, int32, int32, int32, int32, int32, struct{ x, y, z, w float32 })))(&buf[0], 256, 1, 2, 3, 4, 5, 6, 7, 8, struct{ x, y, z, w float32 }{10.0, 20.0, 30.0, 40.0})
				return string(buf[:bytes.IndexByte(buf, 0)])
			},
			want: "1:2:3:4:5:6:7:8:10.0:20.0:30.0:40.0",
		},
		{
			name: "8int_mixed_struct",
			fn: new(func(*byte, uintptr, int32, int32, int32, int32, int32, int32, int32, int32, struct {
				a int32
				b float32
			})),
			cFn: "stack_8int_mixed_struct",
			call: func(f any) string {
				buf := make([]byte, 256)
				(*f.(*func(*byte, uintptr, int32, int32, int32, int32, int32, int32, int32, int32, struct {
					a int32
					b float32
				})))(&buf[0], 256, 1, 2, 3, 4, 5, 6, 7, 8, struct {
					a int32
					b float32
				}{9, 10.0})
				return string(buf[:bytes.IndexByte(buf, 0)])
			},
			want: "1:2:3:4:5:6:7:8:9:10.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "20_int32" && (runtime.GOOS != "darwin" || runtime.GOARCH != "arm64") {
				t.Skip("20 int32 arguments only supported on Darwin ARM64 with smart stack checking")
			}
			if tt.name == "10_float32" && (runtime.GOARCH == "loong64" || runtime.GOARCH == "ppc64le" || runtime.GOARCH == "riscv64" || runtime.GOARCH == "s390x") {
				t.Skip("float32 stack arguments not yet supported on this platform")
			}
			// Struct tests require Darwin ARM64 or AMD64
			if strings.HasPrefix(tt.name, "8int_") && (runtime.GOOS != "darwin" || (runtime.GOARCH != "arm64" && runtime.GOARCH != "amd64")) {
				t.Skip("struct argument tests only supported on Darwin ARM64/AMD64")
			}
			if strings.HasPrefix(tt.name, "8float_") && (runtime.GOOS != "darwin" || (runtime.GOARCH != "arm64" && runtime.GOARCH != "amd64")) {
				t.Skip("struct argument tests only supported on Darwin ARM64/AMD64")
			}

			purego.RegisterLibFunc(tt.fn, lib, tt.cFn)
			got := tt.call(tt.fn)
			if got != tt.want {
				t.Errorf("%s\n  got:  %q\n  want: %q", tt.cFn, got, tt.want)
			}
		})
	}
}

func TestABI_TooManyArguments(t *testing.T) {
	if runtime.GOOS != "darwin" || runtime.GOARCH != "arm64" {
		t.Skip("This test is specific to Darwin ARM64")
	}

	libFileName := filepath.Join(t.TempDir(), "abitest.so")
	if err := buildSharedLib("CC", libFileName, filepath.Join("testdata", "abitest", "abi_test.c")); err != nil {
		t.Fatal(err)
	}
	lib, err := load.OpenLibrary(libFileName)
	if err != nil {
		t.Fatalf("Failed to open library %q: %v", libFileName, err)
	}
	t.Cleanup(func() {
		if err := load.CloseLibrary(lib); err != nil {
			t.Errorf("Failed to close library: %v", err)
		}
	})

	// Test that 35 int64 arguments (27 slots needed) exceeds the limit
	t.Run("35_int64_exceeds_limit", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Logf("Got expected panic: %v", r)
			} else {
				t.Errorf("Expected panic but didn't get one")
			}
		}()

		var fn func(*byte, uintptr, int64, int64, int64, int64, int64, int64, int64, int64, int64, int64, int64, int64, int64, int64, int64, int64, int64, int64, int64, int64, int64, int64, int64, int64, int64, int64, int64, int64, int64, int64, int64, int64, int64, int64, int64)
		purego.RegisterLibFunc(&fn, lib, "stack_35_int64_exceeds")
	})
}

func openBenchmarkLibrary(tb testing.TB) uintptr {
	tb.Helper()

	libFileName := filepath.Join(tb.TempDir(), "libbenchmark.so")
	if err := buildSharedLib("CC", libFileName, filepath.Join("testdata", "benchmarktest", "benchmark.c")); err != nil {
		tb.Fatalf("build benchmark library: %v", err)
	}
	tb.Cleanup(func() {
		_ = os.Remove(libFileName)
	})

	h, err := load.OpenLibrary(libFileName)
	if err != nil {
		tb.Fatalf("open benchmark library %q: %v", libFileName, err)
	}
	tb.Cleanup(func() {
		if err := load.CloseLibrary(h); err != nil {
			tb.Fatalf("close benchmark library: %v", err)
		}
	})
	return h
}

func openBenchmarkSymbol(tb testing.TB, lib uintptr, name string) uintptr {
	tb.Helper()

	sym, err := load.OpenSymbol(lib, name)
	if err != nil {
		tb.Fatalf("open benchmark symbol %q: %v", name, err)
	}
	return sym
}

func buildSharedLib(compilerEnv, libFile string, sources ...string) error {
	out, err := exec.Command("go", "env", compilerEnv).Output()
	if err != nil {
		return fmt.Errorf("go env %s error: %w", compilerEnv, err)
	}

	compiler := strings.TrimSpace(string(out))
	if compiler == "" {
		return errors.New("compiler not found")
	}

	args := []string{"-shared", "-Wall", "-Werror", "-fPIC", "-o", libFile}
	if runtime.GOARCH == "386" {
		args = append(args, "-m32")
	}
	// macOS arm64 can run amd64 tests through Rossetta.
	// Build the shared library based on the GOARCH and not
	// the default behavior of the compiler.
	if runtime.GOOS == "darwin" {
		var arch string
		switch runtime.GOARCH {
		case "arm64":
			arch = "arm64"
		case "amd64":
			arch = "x86_64"
		default:
			return fmt.Errorf("unknown macOS architecture %s", runtime.GOARCH)
		}
		args = append(args, "-arch", arch)
	}
	cmd := exec.Command(compiler, append(args, sources...)...)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("compile lib: %w\n%q\n%s", err, cmd, string(out))
	}

	return nil
}
