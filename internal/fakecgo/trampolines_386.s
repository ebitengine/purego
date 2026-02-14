// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 The Ebitengine Authors

//go:build !cgo && (freebsd || linux)

#include "textflag.h"
#include "go_asm.h"

// These trampolines map the gcc ABI to Go ABI0 and then call into the Go equivalent functions.
// On i386, both GCC and Go use stack-based calling conventions.
//
// When C calls a function, the stack looks like:
//   0(SP) = return address
//   4(SP) = arg1
//   8(SP) = arg2
//   ...
//
// When we declare a Go function with frame size $N-0, Go's prologue
// effectively does SUB $N, SP, so the C arguments shift up by N bytes:
//   N+0(SP) = return address
//   N+4(SP) = arg1
//   N+8(SP) = arg2
//
// Go ABI0 on 386 expects arguments starting at 0(FP) which equals N+4(SP)
// after the prologue (where N is the local frame size).

// func setg_trampoline(setg uintptr, g uintptr)
// This is called from Go, so args are at normal FP positions
TEXT ·setg_trampoline(SB), NOSPLIT, $4-8
	MOVL g+4(FP), AX
	MOVL setg+0(FP), BX

	// setg expects g in 0(SP)
	MOVL AX, 0(SP)
	CALL BX
	RET

TEXT threadentry_trampoline(SB), NOSPLIT, $4-0
	MOVL 8(SP), AX                 // first C arg
	MOVL AX, 0(SP)                 // Go arg 1
	MOVL ·threadentry_call(SB), CX
	MOVL (CX), CX
	CALL CX
	RET

TEXT ·call5(SB), NOSPLIT, $20-28
	MOVL fn+0(FP), AX
	MOVL a1+4(FP), BX
	MOVL a2+8(FP), CX
	MOVL a3+12(FP), DX
	MOVL a4+16(FP), SI
	MOVL a5+20(FP), DI

	// Place arguments on local stack frame for C calling convention
	MOVL BX, 0(SP)     // a1
	MOVL CX, 4(SP)     // a2
	MOVL DX, 8(SP)     // a3
	MOVL SI, 12(SP)    // a4
	MOVL DI, 16(SP)    // a5
	CALL AX
	MOVL AX, r1+24(FP)
	RET
