// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 The Ebitengine Authors

//go:build !cgo && amd64 && (darwin || freebsd || illumos || linux || netbsd)

package fakecgo

import (
	"testing"
	"unsafe"
)

// spillProbeArg marks the call made by TestThreadEntryTrampolinePreservesBX.
var spillProbeArg byte

// spillProbeAbove is the word just above spillProbe's spill slot.
var spillProbeAbove uintptr

// spillProbe stands in for threadentry during the test. Taking the address of
// v makes even an optimized build store v in its spill slot, as threadentry
// does when built with -N. Threads the runtime starts in the meantime still
// get the real threadentry. They have no g yet, which is why spillProbe is
// nosplit and norace like threadentry.
//
//go:nosplit
//go:norace
func spillProbe(v unsafe.Pointer) unsafe.Pointer {
	if v != unsafe.Pointer(&spillProbeArg) {
		return threadentry(v)
	}
	readAboveSlot(&v)
	return nil
}

// readAboveSlot records the word just above *p in spillProbeAbove. It must be
// nosplit: otherwise it could be preempted for a stack scan, which cannot
// unwind through threadentry_trampoline and would crash the test binary. It is
// noinline, which makes spillProbe store v before the call.
//
//go:nosplit
//go:norace
//go:noinline
func readAboveSlot(p *unsafe.Pointer) {
	spillProbeAbove = *(*uintptr)(unsafe.Add(unsafe.Pointer(p), unsafe.Sizeof(*p)))
}

// TestThreadEntryTrampolinePreservesBX checks that threadentry_trampoline
// reserves the spill slot of threadentry's register argument. Without the
// reservation, the slot is where the trampoline saved BX for its C caller, and on darwin/amd64
// _pthread_start reads the thread's pthread_t from BX after threadentry
// returns.
func TestThreadEntryTrampolinePreservesBX(t *testing.T) {
	const bx uintptr = 0x0123456789abcdef
	arg := uintptr(unsafe.Pointer(&spillProbeArg))

	old := threadentry_call
	threadentry_call = spillProbe
	defer func() { threadentry_call = old }()

	spillProbeAbove = 0
	if got := callThreadEntryTrampoline(arg, bx); got != bx {
		t.Fatalf("BX = %#x after threadentry_trampoline, want %#x (the argument was %#x)", got, bx, arg)
	}
	// The saved BX sits just above the reserved slot. If it isn't there, v is
	// not in that slot, because the compiler or the trampoline's layout changed,
	// and this test checks nothing.
	if spillProbeAbove != bx {
		t.Fatalf("word above the spill slot = %#x, want the saved BX %#x: v is not in the reserved slot (the compiler or the trampoline's layout changed)", spillProbeAbove, bx)
	}
}
