// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2023 The Ebitengine Authors

//go:build darwin || freebsd || linux

#include "textflag.h"
#include "abi_amd64.h"
#include "go_asm.h"
#include "funcdata.h"

TEXT callbackasm1(SB), NOSPLIT|NOFRAME, $0
	// remove return address from stack, we are not returning to callbackasm, but to its caller.
	MOVQ 0(SP), AX
	ADDQ $8, SP

	MOVQ 0(SP), R10 // get the return SP so that we can align register args with stack args

	// make space for first six int and 8 float arguments below the frame
	ADJSP $14*8, SP
	MOVSD X0, (1*8)(SP)
	MOVSD X1, (2*8)(SP)
	MOVSD X2, (3*8)(SP)
	MOVSD X3, (4*8)(SP)
	MOVSD X4, (5*8)(SP)
	MOVSD X5, (6*8)(SP)
	MOVSD X6, (7*8)(SP)
	MOVSD X7, (8*8)(SP)
	MOVQ  DI, (9*8)(SP)
	MOVQ  SI, (10*8)(SP)
	MOVQ  DX, (11*8)(SP)
	MOVQ  CX, (12*8)(SP)
	MOVQ  R8, (13*8)(SP)
	MOVQ  R9, (14*8)(SP)
	LEAQ  8(SP), R8      // R8 = address of args vector

	MOVQ R10, 0(SP) // push the stack pointer below registers

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
	MOVQ ·callbackWrap_call(SB), DI // Get the ABIInternal function pointer
	MOVQ (DI), DI                   // without <ABIInternal> by using a closure.
	MOVQ AX, SI                     // frame (address of callbackArgs)
	MOVQ $0, CX                     // context

	CALL crosscall2(SB) // runtime.cgocallback(fn, frame, ctxt uintptr)

	// Get callback result.
	MOVQ (24+callbackArgs_result)(SP), AX
	ADDQ $(24+callbackArgs__size), SP     // remove callbackArgs struct

	POP_REGS_HOST_TO_ABI0()

	MOVQ 0(SP), R10 // get the SP back

	ADJSP $-14*8, SP // remove arguments

	MOVQ R10, 0(SP)

	RET
