// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2022 The Ebitengine Authors

//go:build !go1.16
// +build !go1.16

#include "textflag.h"
#include "internal/abi/abi_amd64.h"
#include "go_asm.h"
#include "funcdata.h"

GLOBL ·cbctxts(SB), NOPTR, $8

TEXT callbackasm1(SB), NOSPLIT, $0
	// remove return address from stack, we are not returning there
	MOVQ 0(SP), AX
	ADDQ $8, SP

	MOVQ 0(SP), R12 // get the return SP so that we can align register args with stack args

	// Construct args vector for cgocallback().
	// The values are in registers.
	ADJSP $6*8, SP
	MOVQ  DI, 8(SP)
	MOVQ  SI, 16(SP)
	MOVQ  DX, 24(SP)
	MOVQ  CX, 32(SP)
	MOVQ  R8, 40(SP)
	MOVQ  R9, 48(SP)
	LEAQ  8(SP), R8  // R8 = address of args vector

	MOVQ R12, 0(SP) // push the stack pointer below registers

	// determine index into runtime·cbctxts table
	MOVQ $callbackasm(SB), DX
	SUBQ DX, AX
	MOVQ $0, DX
	MOVQ $5, CX               // divide by 5 because each call instruction in ·callbacks is 5 bytes long
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

	MOVQ -8(CX)(DX*1), AX // return value
	POPQ -8(CX)(DX*1)     // restore bytes just after the args

	MOVQ  0(SP), R12 // get the SP back
	ADJSP $-6*8, SP  // remove arguments
	MOVQ  R12, 0(SP)
	RET
