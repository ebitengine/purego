// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 The Ebitengine Authors

//go:build linux

package purego

import (
	"math/bits"
	"testing"
	"unsafe"
)

func TestCallbackTrampolineAddresses(t *testing.T) {
	// Check every entry without consuming the process-wide callback slots.
	// On ARMv5, a MOVW of a large immediate can insert a literal pool into
	// the table and make a computed callback address point at data instead.
	stride := callbackasmAddr(1) - callbackasmAddr(0)
	if stride != 12 {
		t.Fatalf("callback stride %d, want 12", stride)
	}
	var target uintptr
	for i := range maxCB {
		pc := callbackasmAddr(i)
		// pc identifies executable native memory, never a Go allocation.
		instructions := (*[3]uint32)(*(*unsafe.Pointer)(unsafe.Pointer(&pc)))
		mov, orr := instructions[0], instructions[1]
		if mov&0xfffff000 != 0xe3a0c000 || orr&0xfffff000 != 0xe38cc000 {
			t.Fatalf("callback %d: expected MOV/ORR immediates into R12, got %#08x/%#08x", i, mov, orr)
		}
		// ARM immediates are an 8-bit value rotated right by an even count.
		low := bits.RotateLeft32(mov&0xff, -2*int((mov>>8)&0xf))
		high := bits.RotateLeft32(orr&0xff, -2*int((orr>>8)&0xf))
		if index := low | high; index != uint32(i) {
			t.Fatalf("callback %d: MOV/ORR load index %d", i, index)
		}
		pc += stride - 4
		instruction := instructions[2]
		if instruction&0xff000000 != 0xea000000 {
			t.Fatalf("callback %d: expected an unconditional branch at %#x, got %#08x", i, pc, instruction)
		}
		// ARM branches use a signed 24-bit word offset from PC+8.
		destination := uintptr(int64(pc) + 8 + int64(int32(instruction<<8)>>6))
		if i == 0 {
			target = destination
		} else if destination != target {
			t.Fatalf("callback %d: branch target %#x, want %#x", i, destination, target)
		}
	}
}
