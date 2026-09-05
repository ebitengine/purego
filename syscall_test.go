// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2023 The Ebitengine Authors

package purego_test

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/ebitengine/purego"
	"github.com/ebitengine/purego/internal/load"
)

func TestOS(t *testing.T) {
	// set and unset an environment variable since this calls into fakecgo.
	err := os.Setenv("TESTING", "SOMETHING")
	if err != nil {
		t.Errorf("failed to Setenv: %s", err)
	}
	err = os.Unsetenv("TESTING")
	if err != nil {
		t.Errorf("failed to Unsetenv: %s", err)
	}
}

func TestErrno(t *testing.T) {
	if !purego.CapturesErrno {
		t.Skip("platform does not support returning errno from syscall")
	}

	// setErrno is called without arguments on purpose. SyscallN mirrors every
	// integer argument into the float slots and the C fallback in internal/cgo,
	// which is used on the Linux architectures that have no assembly trampoline,
	// asserts that those slots are zero. Calling open here would therefore abort
	// on those architectures before the errno check could run.
	libFileName := filepath.Join(t.TempDir(), "liberrnotest.so")
	if err := buildSharedLib(t, "CC", libFileName, filepath.Join("testdata", "liberrnotest", "errno_test.c")); err != nil {
		t.Fatal(err)
	}
	defer os.Remove(libFileName)

	lib, err := load.OpenLibrary(libFileName)
	if err != nil {
		t.Fatal(err)
	}

	setErrno, err := load.OpenSymbol(lib, "setErrno")
	if err != nil {
		t.Fatal(err)
	}

	r1, _, errno := purego.SyscallN(setErrno)
	if int32(r1) != -1 {
		t.Errorf("setErrno returned %d, wanted -1", r1)
	}

	libcName := "/usr/lib/libSystem.B.dylib"
	if runtime.GOOS == "linux" {
		// Reaching here on Linux means that this architecture has no assembly
		// trampoline and goes through the C fallback in internal/cgo, which does
		// report the libc errno.
		libcName = "libc.so.6"
	}
	libc, err := load.OpenLibrary(libcName)
	if err != nil {
		t.Fatal(err)
	}

	var strerror func(int32) string
	purego.RegisterLibFunc(&strerror, libc, "strerror")

	const expected = "No such file or directory"
	got := strerror(int32(errno))
	if got != expected {
		t.Errorf("strerror returned %q, wanted \"%s\"", got, expected)
	}
}
