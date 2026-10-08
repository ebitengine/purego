// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 The Ebitengine Authors

//go:build linux && ppc64le

package purego_test

import (
	"crypto/sha256"
	"path/filepath"
	"testing"

	"github.com/ebitengine/purego"
)

func TestCallbackPreservesVectorRegisters(t *testing.T) {
	libFileName := filepath.Join(t.TempDir(), "libcallback_vregs.so")
	if err := buildSharedLib(t, "CC", libFileName, filepath.Join("testdata", "callback_vregs_ppc64le.S")); err != nil {
		t.Fatal(err)
	}

	lib, err := purego.Dlopen(libFileName, purego.RTLD_NOW|purego.RTLD_LOCAL)
	if err != nil {
		t.Fatal(err)
	}

	var check func(uintptr) int
	purego.RegisterLibFunc(&check, lib, "check_callback_vregs")
	data := make([]byte, 4096)
	callback := purego.NewCallback(func() {
		for range 100 {
			sha256.Sum256(data)
		}
	})

	if got := check(callback); got != 1 {
		t.Fatal("callee-saved vector registers were clobbered")
	}
}
