// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2023 The Ebitengine Authors

package purego_test

import (
	"fmt"
	"os"
	"runtime"
	"testing"

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
	libc := purego.Dlopen(library, purego.RTLD_NOW|purego.RTLD_GLOBAL)
	if err := purego.Dlerror(); err != "" {
		t.Errorf("failed to dlopen: %s", err)
	}
	var puts func(string)
	purego.RegisterLibFunc(&puts, libc, "puts")
	puts("Calling C from from Go without Cgo!")
}
