// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2022 The Ebitengine Authors

#include "textflag.h"
#include "go_asm.h"

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
	MOVQ  DI, 24(BP) // save the pointer

	MOVQ syscall9Args_f1(DI), X0 // f1
	MOVQ syscall9Args_f2(DI), X1 // f2
	MOVQ syscall9Args_f3(DI), X2 // f3
	MOVQ syscall9Args_f4(DI), X3 // f4
	MOVQ syscall9Args_f5(DI), X4 // f5
	MOVQ syscall9Args_f6(DI), X5 // f6
	MOVQ syscall9Args_f7(DI), X6 // f7
	MOVQ syscall9Args_f8(DI), X7 // f8

	MOVQ syscall9Args_fn(DI), R10 // fn
	MOVQ syscall9Args_a2(DI), SI  // a2
	MOVQ syscall9Args_a3(DI), DX  // a3
	MOVQ syscall9Args_a4(DI), CX  // a4
	MOVQ syscall9Args_a5(DI), R8  // a5
	MOVQ syscall9Args_a6(DI), R9  // a6
	MOVQ syscall9Args_a7(DI), R11 // a7
	MOVQ syscall9Args_a8(DI), R12 // a8
	MOVQ syscall9Args_a9(DI), R13 // a9
	MOVQ syscall9Args_a1(DI), DI  // a1

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
