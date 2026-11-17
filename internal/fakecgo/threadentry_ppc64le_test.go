// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 The Ebitengine Authors

//go:build !cgo && linux

package fakecgo

import (
	"testing"
	"unsafe"
)

// spillProbeArg marks the call made by TestThreadEntryTrampolinePreservesR14.
var spillProbeArg byte

// spillProbeAbove is the word just above spillProbe's spill slot.
var spillProbeAbove uintptr

// spillProbe stands in for threadentry during the test. Taking the address of
// v makes the compiler store v in its spill slot. Threads the runtime starts
// in the meantime still get the real threadentry. They have no g yet, which is
// why spillProbe is nosplit and norace like threadentry.
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

// TestThreadEntryTrampolinePreservesR14 checks that threadentry_trampoline
// reserves the spill slot of threadentry's register argument. Without the
// reservation, the slot is where the trampoline saved R14, which the ELFv2 ABI
// makes callee-saved, for its C caller.
func TestThreadEntryTrampolinePreservesR14(t *testing.T) {
	const r14 uintptr = 0x0123456789abcdef
	arg := uintptr(unsafe.Pointer(&spillProbeArg))

	old := threadentry_call
	threadentry_call = spillProbe
	defer func() { threadentry_call = old }()

	spillProbeAbove = 0
	if got := callThreadEntryTrampoline(arg, r14); got != r14 {
		t.Fatalf("R14 = %#x after threadentry_trampoline, want %#x (the argument was %#x)", got, r14, arg)
	}
	// The saved R14 sits just above the reserved slot. If it isn't there, v is
	// not in that slot, because the compiler or the trampoline's layout changed,
	// and this test checks nothing.
	if spillProbeAbove != r14 {
		t.Errorf("word above the spill slot = %#x, want the saved R14 %#x: v is not in the reserved slot (the compiler or the trampoline's layout changed)", spillProbeAbove, r14)
	}
}
