// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2022 The Ebitengine Authors

#include "textflag.h"
#include "internal/abi/abi_amd64.h"
#include "go_asm.h"
#include "funcdata.h"

// syscall9X calls a function in libc on behalf of the syscall package.
// syscall9X takes a pointer to a struct like:
// struct {
//	fn    uintptr
//	a1    uintptr
//	a2    uintptr
//	a3    uintptr
//	a4    uintptr
//	a5    uintptr
//	a6    uintptr
//	a7    uintptr
//	a8    uintptr
//	a9    uintptr
//	r1    uintptr
//	r2    uintptr
//	err   uintptr
// }
// syscall9X must be called on the g0 stack with the
// C calling convention (use libcCall).
GLOBL ·syscall9XABI0(SB), NOPTR|RODATA, $8
DATA ·syscall9XABI0(SB)/8, $syscall9X(SB)
TEXT syscall9X(SB), NOSPLIT, $0
	PUSHQ BP
	MOVQ  SP, BP
	SUBQ  $32, SP
	MOVQ  DI, 24(BP)               // save the pointer
	MOVQ  syscall9Args_fn(DI), R10 // fn
	MOVQ  syscall9Args_a2(DI), SI  // a2
	MOVQ  syscall9Args_a3(DI), DX  // a3
	MOVQ  syscall9Args_a4(DI), CX  // a4
	MOVQ  syscall9Args_a5(DI), R8  // a5
	MOVQ  syscall9Args_a6(DI), R9  // a6
	MOVQ  syscall9Args_a7(DI), R11 // a7
	MOVQ  syscall9Args_a8(DI), R12 // a8
	MOVQ  syscall9Args_a9(DI), R13 // a9
	MOVQ  syscall9Args_a1(DI), DI  // a1

	// these may be float arguments
	// so we put them also where C expects floats
	MOVQ DI, X0  // a1
	MOVQ SI, X1  // a2
	MOVQ DX, X2  // a3
	MOVQ CX, X3  // a4
	MOVQ R8, X4  // a5
	MOVQ R9, X5  // a6
	MOVQ R11, X6 // a7
	MOVQ R12, X7 // a8

	// push the remaining paramters onto the stack
	MOVQ R11, 0(SP)  // push a7
	MOVQ R12, 8(SP)  // push a8
	MOVQ R13, 16(SP) // push a9
	XORL AX, AX      // vararg: say "no float args"

	CALL R10

	MOVQ 24(BP), DI              // get the pointer back
	MOVQ AX, syscall9Args_r1(DI) // r1
	MOVQ DX, syscall9Args_r2(DI) // r2

	XORL AX, AX  // no error (it's ignored anyway)
	ADDQ $32, SP
	MOVQ BP, SP
	POPQ BP
	RET

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

	MOVQ 0(SP), R10 // get the return SP so that we can align register args with stack args

	// make space for first six arguments below the frame
	ADJSP $6*8, SP
	MOVQ  DI, 8(SP)
	MOVQ  SI, 16(SP)
	MOVQ  DX, 24(SP)
	MOVQ  CX, 32(SP)
	MOVQ  R8, 40(SP)
	MOVQ  R9, 48(SP)
	LEAQ  8(SP), R8  // R8 = address of args vector

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
	// $32 to make enough room for the arguments to runtime.cgocallback
	// and on Go 1.15 to take a pointer to the pointer to callbackArgs
	SUBQ $(32+callbackArgs__size), SP
	MOVQ AX, (32+callbackArgs_index)(SP)  // callback index
	MOVQ R8, (32+callbackArgs_args)(SP)   // address of args vector
	MOVQ $0, (32+callbackArgs_result)(SP) // result
	LEAQ 32(SP), AX                       // take the address of callbackArgs

	MOVQ AX, 24(SP) // get a **callbackArgs for go 1.15
	LEAQ 24(SP), AX // ^^^^

	// Call cgocallback, which will call callbackWrap(frame).
	MOVQ $8, 16(SP)                      // context is 0 or argsize (including return value) for Go 1.15
	MOVQ AX, 8(SP)                       // frame (address of callbackArgs)
	MOVQ $callbackWrapInternal<>(SB), BX
	MOVQ BX, 0(SP)                       // PC of function value to call (callbackWrap)
	CALL runtime·cgocallback(SB)         // runtime.cgocallback(fn, frame, ctxt uintptr)

	// Get callback result.
	MOVQ (32+callbackArgs_result)(SP), AX
	ADDQ $(32+callbackArgs__size), SP     // remove callbackArgs struct

	POP_REGS_HOST_TO_ABI0()

	MOVQ 0(SP), R10 // get the SP back

	ADJSP $-6*8, SP // remove arguments

	MOVQ R10, 0(SP)

	RET
