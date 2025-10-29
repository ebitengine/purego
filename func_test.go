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
