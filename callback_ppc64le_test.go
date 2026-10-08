// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 The Ebitengine Authors

//go:build linux

package purego_test

import (
	"testing"
	"unsafe"

	"github.com/ebitengine/purego"
)

func TestCallbackUsesCorrectTOC(t *testing.T) {
	libc, err := purego.Dlopen("libc.so.6", purego.RTLD_NOW|purego.RTLD_GLOBAL)
	if err != nil {
		t.Fatal(err)
	}

	var iterateProgramHeaders func(uintptr, unsafe.Pointer) int32
	purego.RegisterLibFunc(&iterateProgramHeaders, libc, "dl_iterate_phdr")

	called := false
	callback := func(info unsafe.Pointer, size uintptr, data unsafe.Pointer) int32 {
		called = true
		return 0
	}
	_ = purego.NewCallback(callback)
	if result := iterateProgramHeaders(purego.NewCallback(callback), nil); result != 0 {
		t.Errorf("dl_iterate_phdr returned %d, want 0", result)
	}
	if !called {
		t.Error("callback was not called")
	}
}
