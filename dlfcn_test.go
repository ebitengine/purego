// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2023 The Ebitengine Authors

//go:build darwin || linux

package purego_test

import (
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
)

func getSystemLibrary() (string, error) {
	switch runtime.GOOS {
	case "darwin":
		return "/usr/lib/libSystem.B.dylib", nil
	case "linux":
		return "libc.so.6", nil
	default:
		return "", fmt.Errorf("GOOS=%s is not supported", runtime.GOOS)
	}
}

func TestRegisterFunc(t *testing.T) {
	library, err := getSystemLibrary()
	if err != nil {
		t.Errorf("couldn't get system library: %s", err)
	}
	libc, err := purego.Dlopen(library, purego.RTLD_NOW|purego.RTLD_GLOBAL)
	if err != nil {
		t.Errorf("failed to dlopen: %s", err)
	}
	var puts func(string)
	purego.RegisterLibFunc(&puts, libc, "puts")
	puts("Calling C from from Go without Cgo!")
}

func ExampleNewCallback() {
	if runtime.GOOS == "linux" {
		// TODO: enable once callbacks are working properly on Linux
		fmt.Println("1 2 3 4 5 6 7 8 9\n45")
		return
	}

	cb := purego.NewCallback(func(a1, a2, a3, a4, a5, a6, a7, a8, a9 int) int {
		fmt.Println(a1, a2, a3, a4, a5, a6, a7, a8, a9)
		return a1 + a2 + a3 + a4 + a5 + a6 + a7 + a8 + a9
	})

	var fn func(a1, a2, a3, a4, a5, a6, a7, a8, a9 int) int
	purego.RegisterFunc(&fn, cb)

	ret := fn(1, 2, 3, 4, 5, 6, 7, 8, 9)
	fmt.Println(ret)

	//Output: 1 2 3 4 5 6 7 8 9
	// 45
}

func Test_qsort(t *testing.T) {
	if runtime.GOOS == "linux" {
		// TODO: enable once callbacks are working properly on Linux
		t.SkipNow()
	}

	library, err := getSystemLibrary()
	if err != nil {
		t.Errorf("couldn't get system library: %s", err)
	}
	libc, err := purego.Dlopen(library, purego.RTLD_NOW|purego.RTLD_GLOBAL)
	if err != nil {
		t.Errorf("failed to dlopen: %s", err)
	}

	var data = []int{88, 56, 100, 2, 25}
	var sorted = []int{2, 25, 56, 88, 100}
	compare := func(a, b *int) int {
		return *a - *b
	}
	var qsort func(data []int, nitms uintptr, size uintptr, compar func(a, b *int) int)
	purego.RegisterLibFunc(&qsort, libc, "qsort")
	qsort(data, uintptr(len(data)), unsafe.Sizeof(int(0)), compare)
	for i := range data {
		if data[i] != sorted[i] {
			t.Errorf("got %d wanted %d at %d", data[i], sorted[i], i)
		}
	}
}

func TestNestedDlopenCall(t *testing.T) {
	libFileName := filepath.Join(t.TempDir(), "libdlnested.so")
	t.Logf("Build %v", libFileName)

	if err := buildSharedLib("CXX", libFileName, filepath.Join("libdlnested", "nested.cpp")); err != nil {
		t.Fatal(err)
	}
	defer os.Remove(libFileName)

	lib, err := purego.Dlopen(libFileName, purego.RTLD_NOW|purego.RTLD_GLOBAL)
	if err != nil {
		t.Fatalf("Dlopen(%q) failed: %v", libFileName, err)
	}

	purego.Dlclose(lib)
}

func buildSharedLib(compilerEnv, libFile string, sources ...string) error {
	out, err := exec.Command("go", "env", compilerEnv).Output()
	if err != nil {
		return fmt.Errorf("go env error: %w", err)
	}

	compiler := strings.TrimSpace(string(out))
	if compiler == "" {
		return errors.New("compiler not found")
	}

	args := []string{"-shared", "-Wall", "-Werror", "-o", libFile}

	// macOS arm64 can run amd64 tests through Rossetta.
	// Build the shared library based on the GOARCH and not
	// the default behavior of the compiler.
	var archFlag, arch string
	if runtime.GOOS == "darwin" {
		archFlag = "-arch"
		switch runtime.GOARCH {
		case "arm64":
			arch = "arm64"
		case "amd64":
			arch = "x86_64"
		default:
			// if GOARCH is unknown then build the shared library
			// with the compiler's normal target.
			archFlag = ""
			arch = ""
		}
		args = append(args, archFlag, arch)
	}
	cmd := exec.Command(compiler, append(args, sources...)...)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("compile lib: %w\n%q\n%s", err, cmd, string(out))
	}

	return nil
}
