// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 The Ebitengine Authors

package purego_test

import (
	"math/bits"
	"testing"
	"unsafe"

	"github.com/ebitengine/purego"
)

func TestCallbackTrampolineEncoding(t *testing.T) {
	// Check every entry without consuming the process-wide callback slots.
	// Below ARMv7, a MOVW of a large immediate can insert a literal pool into
	// the table and make a computed callback address point at data instead.
	const maxFailures = 20
	failures := 0
	fail := func(format string, args ...any) {
		t.Helper()
		t.Errorf(format, args...)
		if failures++; failures >= maxFailures {
			t.Fatalf("stopping after %d failures", failures)
		}
	}
	target := purego.Callbackasm1Addr()
	for i := range purego.MaxCB {
		pc := purego.CallbackasmAddr(i)
		// pc identifies executable native memory, never a Go allocation.
		instructions := (*[3]uint32)(*(*unsafe.Pointer)(unsafe.Pointer(&pc)))
		mov, orr, branch := instructions[0], instructions[1], instructions[2]
		if mov&0xfffff000 != 0xe3a0c000 || orr&0xfffff000 != 0xe38cc000 {
			fail("callback %d: expected MOV/ORR immediates into R12, got %#08x/%#08x", i, mov, orr)
			continue
		}
		// ARM immediates are an 8-bit value rotated right by an even count.
		low := bits.RotateLeft32(mov&0xff, -2*int((mov>>8)&0xf))
		high := bits.RotateLeft32(orr&0xff, -2*int((orr>>8)&0xf))
		if index := low | high; index != uint32(i) {
			fail("callback %d: MOV/ORR load index %d", i, index)
		}
		if branch&0xff000000 != 0xea000000 {
			fail("callback %d: expected an unconditional branch at %#x, got %#08x", i, pc+8, branch)
			continue
		}
		// ARM branches use a signed 24-bit word offset from their own address
		// plus 8. The branch is at pc+8.
		destination := pc + 16 + uintptr(int32(branch<<8)>>6)
		if destination != target {
			fail("callback %d: branch target %#x, want callbackasm1 at %#x", i, destination, target)
		}
	}
}
