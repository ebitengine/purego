// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2023 The Ebitengine Authors

//go:build darwin || linux

package purego_test

import (
	"os"
	"path/filepath"
	"testing"
	"unsafe"

	"github.com/ebitengine/purego"
)

// TestCallGoFromSharedLib is a test that checks for stack corruption on darwin/arm64
// when C calls Go code from a non-Go thread in a dynamically loaded share library.
func TestCallGoFromSharedLib(t *testing.T) {
	libFileName := filepath.Join(t.TempDir(), "libcbtest.so")
	t.Logf("Build %v", libFileName)

	if err := buildSharedLib("CC", libFileName, filepath.Join("libcbtest", "callback.c")); err != nil {
		t.Fatal(err)
	}
	defer os.Remove(libFileName)

	lib, err := purego.Dlopen(libFileName, purego.RTLD_NOW|purego.RTLD_GLOBAL)
	if err != nil {
		t.Fatalf("Dlopen(%q) failed: %v", libFileName, err)
	}

	var callCallback func(p uintptr, s string) int
	purego.RegisterLibFunc(&callCallback, lib, "callCallback")

	goFunc := func(cstr *byte, n int) int {
		s := string(unsafe.Slice(cstr, n))
		t.Logf("FROM Go: %s\n", s)
		return 1
	}

	const want = 10101
	cb := purego.NewCallback(goFunc)
	for i := 0; i < 10; i++ {
		got := callCallback(cb, "a test string")
		if got != want {
			t.Fatalf("%d: callCallback() got %v want %v", i, got, want)
		}
	}
}
