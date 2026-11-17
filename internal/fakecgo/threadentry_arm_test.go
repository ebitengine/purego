// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 The Ebitengine Authors

//go:build !cgo && linux

package fakecgo

import (
	"testing"
	"unsafe"
)

// On 32-bit ELF the internal linker records a needed library only for a _ _
// import, not for the library named with each symbol in zsymbols_linux.go. A
// test binary of fakecgo alone then has no DT_NEEDED and fails to start, which
// is why threadentry_test.go skips arm. Naming libc here lets this one start.
//
//go:cgo_import_dynamic _ _ "libc.so.6"

// resultProbeArg marks the call made by TestThreadEntryTrampolinePreservesR4.
var resultProbeArg byte

// resultProbe stands in for threadentry during the test and, like it, returns
// nil, which ABI0 stores in its result slot. Threads the runtime starts in the
// meantime still get the real threadentry. They have no g yet, which is why
// resultProbe is nosplit and norace like threadentry.
//
//go:nosplit
//go:norace
func resultProbe(v unsafe.Pointer) unsafe.Pointer {
	if v != unsafe.Pointer(&resultProbeArg) {
		return threadentry(v)
	}
	return nil
}

// TestThreadEntryTrampolinePreservesR4 checks that threadentry_trampoline
// keeps the C caller's R4, which AAPCS makes callee-saved, out of the result
// slot of threadentry. threadentry is ABI0, so its argument is at 4(R13) and
// its result at 8(R13) of the trampoline's frame.
func TestThreadEntryTrampolinePreservesR4(t *testing.T) {
	const r4 uintptr = 0x89abcdef
	arg := uintptr(unsafe.Pointer(&resultProbeArg))

	old := threadentry_call
	threadentry_call = resultProbe
	defer func() { threadentry_call = old }()

	if got := callThreadEntryTrampoline(arg, r4); got != r4 {
		t.Errorf("R4 = %#x after threadentry_trampoline, want %#x", got, r4)
	}
}
