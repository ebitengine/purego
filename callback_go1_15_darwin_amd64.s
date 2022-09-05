// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2022 The Ebitengine Authors

//go:build !go1.16
// +build !go1.16

#include "textflag.h"
#include "internal/abi/abi_amd64.h"
#include "go_asm.h"
#include "funcdata.h"

// runtime·cgocallback expects a call to the ABIInternal function
// However, the tag <ABIInternal> is only available in the runtime :(
// This is a small wrapper function that copies both whatever is in the register
// and is on the stack and places both on the stack. It then calls callbackWrapPicker
// which will choose which parameter should be used depending on the version of Go.
// It then calls the real version of callbackWrap
TEXT callbackWrapInternal<>(SB), NOSPLIT, $0-0
	MOVQ AX, arg+16(SP)
	JMP  ·callbackWrapPicker(SB)
	RET

GLOBL ·cbctxts(SB), NOPTR, $8

TEXT callbackasm1(SB), NOSPLIT, $0
	// remove return address from stack, we are not returning there
	MOVQ 0(SP), AX
	ADDQ $8, SP

	// Construct args vector for cgocallback().
	// The values are in registers.
	ADJSP $6*8, SP
	MOVQ  DI, 0(SP)
	MOVQ  SI, 8(SP)
	MOVQ  DX, 16(SP)
	MOVQ  CX, 24(SP)
	MOVQ  R8, 32(SP)
	MOVQ  R9, 40(SP)

	// determine index into runtime·cbctxts table
	MOVQ $callbackasm(SB), DX
	SUBQ DX, AX
	MOVQ $0, DX
	MOVQ $5, CX               // divide by 5 because each call instruction in runtime·callbacks is 5 bytes long
	DIVL CX

	// find correspondent runtime·cbctxts table entry
	MOVQ ·cbctxts(SB), CX
	MOVQ -8(CX)(AX*8), AX

	// extract callback context
	MOVQ callbackcontext_argsize(AX), DX
	MOVQ callbackcontext_gobody(AX), AX

	// preserve whatever's at the memory location that
	// the callback will use to store the return value
	LEAQ  8(SP), CX   // args vector, skip return address
	PUSHQ 0(CX)(DX*1) // store 8 bytes from just after the args array
	ADDQ  $8, DX      // extend argsize by size of return value

	// Switch from the host ABI to the Go ABI.
	PUSH_REGS_HOST_TO_ABI0()

	// prepare call stack.  use SUBQ to hide from stack frame checks
	// cgocallback(Go func, void *frame, uintptr framesize)
	SUBQ $24, SP
	MOVQ DX, 16(SP)              // argsize (including return value)
	MOVQ CX, 8(SP)               // callback parameters
	MOVQ AX, 0(SP)               // address of target Go function
	CLD
	CALL runtime·cgocallback(SB)
	MOVQ 0(SP), AX
	MOVQ 8(SP), CX
	MOVQ 16(SP), DX
	ADDQ $24, SP

	POP_REGS_HOST_TO_ABI0()

	MOVQ  -8(CX)(DX*1), AX // return value
	POPQ  -8(CX)(DX*1)     // restore bytes just after the args
	ADJSP $-6*8, SP        // remove arguments
	RET
