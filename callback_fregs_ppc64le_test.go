// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 The Ebitengine Authors

//go:build linux

package purego_test

import (
	"crypto/sha256"
	"path/filepath"
	"testing"

	"github.com/ebitengine/purego"
)

func TestCallbackPreservesFloatRegisters(t *testing.T) {
	libFileName := filepath.Join(t.TempDir(), "libcallback_fregs.so")
	if err := buildSharedLib(t, "CC", libFileName, filepath.Join("testdata", "libcbtest", "callback_fregs_ppc64le.S")); err != nil {
		t.Fatal(err)
	}

	lib, err := purego.Dlopen(libFileName, purego.RTLD_NOW|purego.RTLD_LOCAL)
	if err != nil {
		t.Fatal(err)
	}

	var check func(uintptr) uint64
	purego.RegisterLibFunc(&check, lib, "check_callback_fregs")
	data := make([]byte, 4096)
	callback := purego.NewCallback(func() {
		for range 100 {
			sha256.Sum256(data)
		}
	})

	mask := check(callback)
	for i := range 18 {
		if mask&(1<<i) != 0 {
			t.Errorf("F%d was clobbered", 14+i)
		}
	}
}
