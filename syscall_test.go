// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2023 The Ebitengine Authors

package purego_test

import (
	"os"
	"runtime"
	"testing"
	"unsafe"

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
	if runtime.GOOS != "darwin" {
		t.Skip("platform does not support returning errno from syscall")
	}

	libc, err := load.OpenLibrary("/usr/lib/libSystem.B.dylib")
	if err != nil {
		t.Fatal(err)
	}

	openSym, err := load.OpenSymbol(libc, "open")
	if err != nil {
		t.Fatal(err)
	}

	r1, _, errno := purego.SyscallN(openSym, uintptr(unsafe.Pointer(&[]byte("_file_that_does_not_exist_\x00")[0])), uintptr(os.O_RDWR))
	if int32(r1) != -1 {
		t.Errorf("open returned %d, wanted -1", r1)
	}

	var strerror func(int32) string
	purego.RegisterLibFunc(&strerror, libc, "strerror")

	const expected = "No such file or directory"
	got := strerror(int32(errno))
	if got != expected {
		t.Errorf("strerror returned %q, wanted \"%s\"", got, expected)
	}
}
