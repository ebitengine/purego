// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2022 The Ebitengine Authors

//go:build go1.16
// +build go1.16

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

TEXT callbackasm1(SB), NOSPLIT, $0
	// remove return address from stack, we are not returning to callbackasm, but to its caller.
	MOVQ 0(SP), AX
	ADDQ $8, SP

	// make space for first six arguments below the frame
	// TODO: check to make sure that arguments 7 and above are right after these six
	ADJSP $6*8, SP
	MOVQ  DI, 0(SP)
	MOVQ  SI, 8(SP)
	MOVQ  DX, 16(SP)
	MOVQ  CX, 24(SP)
	MOVQ  R8, 32(SP)
	MOVQ  R9, 40(SP)
	LEAQ  (SP), R8   // R8 = address of args vector

	// determine index into runtime·cbs table
	MOVQ $callbackasm(SB), DX
	SUBQ DX, AX
	MOVQ $0, DX
	MOVQ $5, CX               // divide by 5 because each call instruction in ·callbacks is 5 bytes long
	DIVL CX
	SUBQ $1, AX               // subtract 1 because return PC is to the next slot

	// Switch from the host ABI to the Go ABI.
	PUSH_REGS_HOST_TO_ABI0()

	// Create a struct callbackArgs on our stack to be passed as
	// the "frame" to cgocallback and on to callbackWrap.
	// $24 to make enough room for the arguments to runtime.cgocallback
	SUBQ $(24+callbackArgs__size), SP
	MOVQ AX, (24+callbackArgs_index)(SP)  // callback index
	MOVQ R8, (24+callbackArgs_args)(SP)   // address of args vector
	MOVQ $0, (24+callbackArgs_result)(SP) // result
	LEAQ 24(SP), AX                       // take the address of callbackArgs

	// Call cgocallback, which will call callbackWrap(frame).
	MOVQ $0, 16(SP)                      // context
	MOVQ AX, 8(SP)                       // frame (address of callbackArgs)
	MOVQ $callbackWrapInternal<>(SB), BX
	MOVQ BX, 0(SP)                       // PC of function value to call (callbackWrap)
	CALL runtime·cgocallback(SB)         // runtime.cgocallback(fn, frame, ctxt uintptr)

	// Get callback result.
	MOVQ (24+callbackArgs_result)(SP), AX
	ADDQ $(24+callbackArgs__size), SP     // remove callbackArgs struct

	POP_REGS_HOST_TO_ABI0()

	ADJSP $-6*8, SP // remove arguments

	RET
