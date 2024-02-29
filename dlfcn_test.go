// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2023 The Ebitengine Authors

//go:build darwin || freebsd || linux

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

func TestSimpleDlsym(t *testing.T) {
	if _, err := purego.Dlsym(purego.RTLD_DEFAULT, "dlsym"); err != nil {
		t.Errorf("Dlsym with RTLD_DEFAULT failed: %v", err)
	}
}

func TestNestedDlopenCall(t *testing.T) {
	libFileName := filepath.Join(t.TempDir(), "libdlnested.so")
	t.Logf("Build %v", libFileName)

	if err := buildSharedLib("CXX", libFileName, filepath.Join("testdata", "libdlnested", "nested_test.cpp")); err != nil {
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
		return fmt.Errorf("go env %s error: %w", compilerEnv, err)
	}

	compiler := strings.TrimSpace(string(out))
	if compiler == "" {
		return errors.New("compiler not found")
	}

	var args []string
	if runtime.GOOS == "freebsd" {
		args = []string{"-shared", "-Wall", "-Werror", "-fPIC", "-o", libFile}
	} else {
		args = []string{"-shared", "-Wall", "-Werror", "-o", libFile}
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

func TestSyscallN(t *testing.T) {
	var dlsym uintptr
	var err error
	if dlsym, err = purego.Dlsym(purego.RTLD_DEFAULT, "dlsym"); err != nil {
		t.Errorf("Dlsym with RTLD_DEFAULT failed: %v", err)
	}
	r1, _, err2 := purego.SyscallN(dlsym, purego.RTLD_DEFAULT, uintptr(unsafe.Pointer(&[]byte("dlsym\x00")[0])))
	if dlsym != r1 {
		t.Fatalf("SyscallN didn't return the same result as purego.Dlsym: %d", err2)
	}
}
