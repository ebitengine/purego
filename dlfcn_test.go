// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2023 The Ebitengine Authors

//go:build darwin || freebsd || linux || netbsd

package purego_test

import (
	"os"
	"path/filepath"
	"syscall"
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

	if err := buildSharedLib(t, "CXX", libFileName, filepath.Join("testdata", "libdlnested", "nested_test.cpp")); err != nil {
		t.Fatal(err)
	}
	defer os.Remove(libFileName)

	lib, err := purego.Dlopen(libFileName, purego.RTLD_NOW|purego.RTLD_GLOBAL)
	if err != nil {
		t.Fatalf("Dlopen(%q) failed: %v", libFileName, err)
	}

	purego.Dlclose(lib)
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

func TestSyscallNErrnoIsNotAnInputArgument(t *testing.T) {
	if purego.CapturesErrno {
		t.Skip("this platform saves errno; see TestErrno")
	}

	openSym, err := purego.Dlsym(purego.RTLD_DEFAULT, "open")
	if err != nil {
		t.Fatal(err)
	}

	path, err := syscall.BytePtrFromString("_file_that_does_not_exist_")
	if err != nil {
		t.Fatal(err)
	}

	// The third argument is non-zero so that echoing it back as errno is visible.
	const third = uintptr(0o600)
	r1, _, errno := purego.SyscallN(openSym,
		uintptr(unsafe.Pointer(path)),
		uintptr(os.O_RDWR),
		third)
	if int32(r1) != -1 {
		t.Fatalf("open returned %d, wanted -1", r1)
	}
	if errno != 0 {
		t.Errorf("SyscallN returned %d as errno where errno is not captured, wanted 0", errno)
	}
}
