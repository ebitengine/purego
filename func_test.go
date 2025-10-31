// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2023 The Ebitengine Authors

package purego_test

import (
	"errors"
	"fmt"
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
	if runtime.GOARCH != "arm64" && runtime.GOARCH != "amd64" && runtime.GOARCH != "loong64" {
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
	if runtime.GOARCH != "arm64" && runtime.GOARCH != "amd64" && runtime.GOARCH != "loong64" {
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
	if runtime.GOARCH != "arm64" && runtime.GOARCH != "amd64" && runtime.GOARCH != "loong64" {
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
		res := string(buf[:strings.IndexByte(string(buf), 0)])
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
		t.Fatalf("Dlopen(%q) failed: %v", libFileName, err)
	}
	t.Cleanup(func() {
		if err := load.CloseLibrary(lib); err != nil {
			t.Errorf("Failed to close library: %v", err)
		}
	})

	tests := []struct {
		name string
		fn   interface{}
		cFn  string
		call func(interface{}) string
		want string
	}{
		{
			name: "10_int32_baseline",
			fn:   new(func(*byte, uintptr, int32, int32, int32, int32, int32, int32, int32, int32, int32, int32)),
			cFn:  "stack_10_int32",
			call: func(f interface{}) string {
				buf := make([]byte, 256)
				(*f.(*func(*byte, uintptr, int32, int32, int32, int32, int32, int32, int32, int32, int32, int32)))(&buf[0], 256, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10)
				return string(buf[:strings.IndexByte(string(buf), 0)])
			},
			want: "1:2:3:4:5:6:7:8:9:10",
		},
		{
			name: "11_int32",
			fn:   new(func(*byte, uintptr, int32, int32, int32, int32, int32, int32, int32, int32, int32, int32, int32)),
			cFn:  "stack_11_int32",
			call: func(f interface{}) string {
				buf := make([]byte, 256)
				(*f.(*func(*byte, uintptr, int32, int32, int32, int32, int32, int32, int32, int32, int32, int32, int32)))(&buf[0], 256, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11)
				return string(buf[:strings.IndexByte(string(buf), 0)])
			},
			want: "1:2:3:4:5:6:7:8:9:10:11",
		},
		{
			name: "10_float32",
			fn:   new(func(*byte, uintptr, float32, float32, float32, float32, float32, float32, float32, float32, float32, float32)),
			cFn:  "stack_10_float32",
			call: func(f interface{}) string {
				buf := make([]byte, 256)
				(*f.(*func(*byte, uintptr, float32, float32, float32, float32, float32, float32, float32, float32, float32, float32)))(&buf[0], 256, 1.0, 2.0, 3.0, 4.0, 5.0, 6.0, 7.0, 8.0, 9.0, 10.0)
				return string(buf[:strings.IndexByte(string(buf), 0)])
			},
			want: "1.0:2.0:3.0:4.0:5.0:6.0:7.0:8.0:9.0:10.0",
		},
		{
			name: "mixed_stack_strings",
			fn:   new(func(*byte, uintptr, int32, int32, int32, int32, int32, int32, int32, int32, string, bool, int32, string)),
			cFn:  "stack_mixed_stack_4args",
			call: func(f interface{}) string {
				buf := make([]byte, 256)
				(*f.(*func(*byte, uintptr, int32, int32, int32, int32, int32, int32, int32, int32, string, bool, int32, string)))(&buf[0], 256, 1, 2, 3, 4, 5, 6, 7, 8, "foo", false, 99, "bar")
				return string(buf[:strings.IndexByte(string(buf), 0)])
			},
			want: "1:2:3:4:5:6:7:8:foo:0:99:bar",
		},
		{
			name: "20_int32",
			fn:   new(func(*byte, uintptr, int32, int32, int32, int32, int32, int32, int32, int32, int32, int32, int32, int32, int32, int32, int32, int32, int32, int32, int32, int32)),
			cFn:  "stack_20_int32",
			call: func(f interface{}) string {
				buf := make([]byte, 512)
				(*f.(*func(*byte, uintptr, int32, int32, int32, int32, int32, int32, int32, int32, int32, int32, int32, int32, int32, int32, int32, int32, int32, int32, int32, int32)))(&buf[0], 512, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20)
				return string(buf[:strings.IndexByte(string(buf), 0)])
			},
			want: "1:2:3:4:5:6:7:8:9:10:11:12:13:14:15:16:17:18:19:20",
		},
		{
			name: "8int_hfa2_stack",
			fn:   new(func(*byte, uintptr, int32, int32, int32, int32, int32, int32, int32, int32, struct{ x, y float32 })),
			cFn:  "stack_8int_hfa2_stack",
			call: func(f interface{}) string {
				buf := make([]byte, 256)
				(*f.(*func(*byte, uintptr, int32, int32, int32, int32, int32, int32, int32, int32, struct{ x, y float32 })))(&buf[0], 256, 1, 2, 3, 4, 5, 6, 7, 8, struct{ x, y float32 }{10.0, 20.0})
				return string(buf[:strings.IndexByte(string(buf), 0)])
			},
			want: "1:2:3:4:5:6:7:8:10.0:20.0",
		},
		{
			name: "8int_2structs_stack",
			fn:   new(func(*byte, uintptr, int32, int32, int32, int32, int32, int32, int32, int32, struct{ x, y int32 }, struct{ x, y int32 })),
			cFn:  "stack_8int_2structs_stack",
			call: func(f interface{}) string {
				buf := make([]byte, 256)
				(*f.(*func(*byte, uintptr, int32, int32, int32, int32, int32, int32, int32, int32, struct{ x, y int32 }, struct{ x, y int32 })))(&buf[0], 256, 1, 2, 3, 4, 5, 6, 7, 8, struct{ x, y int32 }{9, 10}, struct{ x, y int32 }{11, 12})
				return string(buf[:strings.IndexByte(string(buf), 0)])
			},
			want: "1:2:3:4:5:6:7:8:9:10:11:12",
		},
		{
			name: "8float_hfa2_stack",
			fn:   new(func(*byte, uintptr, float32, float32, float32, float32, float32, float32, float32, float32, struct{ x, y float32 })),
			cFn:  "stack_8float_hfa2_stack",
			call: func(f interface{}) string {
				buf := make([]byte, 256)
				(*f.(*func(*byte, uintptr, float32, float32, float32, float32, float32, float32, float32, float32, struct{ x, y float32 })))(&buf[0], 256, 1.0, 2.0, 3.0, 4.0, 5.0, 6.0, 7.0, 8.0, struct{ x, y float32 }{9.0, 10.0})
				return string(buf[:strings.IndexByte(string(buf), 0)])
			},
			want: "1.0:2.0:3.0:4.0:5.0:6.0:7.0:8.0:9.0:10.0",
		},
		{
			name: "8int_hfa2_floatregs",
			fn:   new(func(*byte, uintptr, int32, int32, int32, int32, int32, int32, int32, int32, struct{ x, y float32 })),
			cFn:  "stack_8int_hfa2_floatregs",
			call: func(f interface{}) string {
				buf := make([]byte, 256)
				(*f.(*func(*byte, uintptr, int32, int32, int32, int32, int32, int32, int32, int32, struct{ x, y float32 })))(&buf[0], 256, 1, 2, 3, 4, 5, 6, 7, 8, struct{ x, y float32 }{10.0, 20.0})
				return string(buf[:strings.IndexByte(string(buf), 0)])
			},
			want: "1:2:3:4:5:6:7:8:10.0:20.0",
		},
		{
			name: "8int_int_struct_int",
			fn:   new(func(*byte, uintptr, int32, int32, int32, int32, int32, int32, int32, int32, int32, struct{ x, y int32 }, int32)),
			cFn:  "stack_8int_int_struct_int",
			call: func(f interface{}) string {
				buf := make([]byte, 256)
				(*f.(*func(*byte, uintptr, int32, int32, int32, int32, int32, int32, int32, int32, int32, struct{ x, y int32 }, int32)))(&buf[0], 256, 1, 2, 3, 4, 5, 6, 7, 8, 9, struct{ x, y int32 }{10, 11}, 12)
				return string(buf[:strings.IndexByte(string(buf), 0)])
			},
			want: "1:2:3:4:5:6:7:8:9:10:11:12",
		},
		{
			name: "8int_hfa4_stack",
			fn:   new(func(*byte, uintptr, int32, int32, int32, int32, int32, int32, int32, int32, struct{ x, y, z, w float32 })),
			cFn:  "stack_8int_hfa4_stack",
			call: func(f interface{}) string {
				buf := make([]byte, 256)
				(*f.(*func(*byte, uintptr, int32, int32, int32, int32, int32, int32, int32, int32, struct{ x, y, z, w float32 })))(&buf[0], 256, 1, 2, 3, 4, 5, 6, 7, 8, struct{ x, y, z, w float32 }{10.0, 20.0, 30.0, 40.0})
				return string(buf[:strings.IndexByte(string(buf), 0)])
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
			call: func(f interface{}) string {
				buf := make([]byte, 256)
				(*f.(*func(*byte, uintptr, int32, int32, int32, int32, int32, int32, int32, int32, struct {
					a int32
					b float32
				})))(&buf[0], 256, 1, 2, 3, 4, 5, 6, 7, 8, struct {
					a int32
					b float32
				}{9, 10.0})
				return string(buf[:strings.IndexByte(string(buf), 0)])
			},
			want: "1:2:3:4:5:6:7:8:9:10.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "20_int32" && (runtime.GOOS != "darwin" || runtime.GOARCH != "arm64") {
				t.Skip("20 int32 arguments only supported on Darwin ARM64 with smart stack checking")
			}
			if tt.name == "10_float32" && (runtime.GOARCH == "386" || runtime.GOARCH == "arm" || runtime.GOARCH == "loong64") {
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
		t.Fatalf("Dlopen(%q) failed: %v", libFileName, err)
	}
	t.Cleanup(func() {
		if err := load.CloseLibrary(lib); err != nil {
			t.Errorf("Failed to close library: %v", err)
		}
	})

	// Test that 25 int64 arguments (17 slots needed) exceeds the limit
	t.Run("25_int64_exceeds_limit", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				msg := fmt.Sprint(r)
				if !strings.Contains(msg, "too many stack arguments") {
					t.Errorf("Expected detailed error message, got: %v", r)
				}
				t.Logf("Got expected panic with message: %v", r)
			} else {
				t.Errorf("Expected panic but didn't get one")
			}
		}()

		var fn func(*byte, uintptr, int64, int64, int64, int64, int64, int64, int64, int64, int64, int64, int64, int64, int64, int64, int64, int64, int64, int64, int64, int64, int64, int64, int64, int64, int64)
		purego.RegisterLibFunc(&fn, lib, "stack_25_int64_exceeds")
	})
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
