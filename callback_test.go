// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2023 The Ebitengine Authors

//go:build darwin

package purego_test

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"unsafe"

	"github.com/ebitengine/purego"
)

// TestCallGoFromSharedLib is a test that checks for stack corruption on darwin/arm64
// when C calls Go code from a non-Go thread in a dynamically loaded share library.
func TestCallGoFromSharedLib(t *testing.T) {
	libFileName := filepath.Join(t.TempDir(), "libcbtest.so")
	t.Logf("Build %v", libFileName)

	if err := buildSharedLib(libFileName, filepath.Join("libcbtest", "callback.c")); err != nil {
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

func buildSharedLib(libFile string, sources ...string) error {
	out, err := exec.Command("go", "env", "CC").Output()
	if err != nil {
		return fmt.Errorf("go env error: %w", err)
	}

	cc := strings.TrimSpace(string(out))
	if cc == "" {
		return errors.New("CC not found")
	}

	args := append([]string{"-shared", "-Wall", "-Werror", "-o", libFile}, sources...)
	cmd := exec.Command(cc, args...)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("compile lib: %w\n%q\n%s", err, cmd, string(out))
	}

	return nil
}

func TestNewCallback(t *testing.T) {
	// This tests the maximum number of arguments a function to NewCallback can take
	const (
		expectCbTotal    = -3
		expectedCbTotalF = float64(36)
	)
	var cbTotal int
	var cbTotalF float64
	imp := purego.NewCallback(func(a1, a2, a3, a4, a5, a6, a7, a8, a9 int,
		f1, f2, f3, f4, f5, f6, f7, f8 float64) {
		cbTotal = a1 + a2 + a3 + a4 + a5 + a6 + a7 + a8 + a9
		cbTotalF = f1 + f2 + f3 + f4 + f5 + f6 + f7 + f8
	})
	var fn func(a1, a2, a3, a4, a5, a6, a7, a8, a9 int,
		f1, f2, f3, f4, f5, f6, f7, f8 float64)
	purego.RegisterFunc(&fn, imp)
	fn(1, 2, -3, 4, -5, 6, -7, 8, -9,
		1, 2, 3, 4, 5, 6, 7, 8)

	if cbTotal != expectCbTotal {
		t.Fatalf("cbTotal not correct got %d but wanted %d", cbTotal, expectCbTotal)
	}
	if cbTotalF != expectedCbTotalF {
		t.Fatalf("cbTotal not correct got %f but wanted %f", cbTotalF, expectedCbTotalF)
	}
}
