// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2023 The Ebitengine Authors

package purego_test

import (
	"os"
	"testing"
	"unsafe"

	"github.com/ebitengine/purego"
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
