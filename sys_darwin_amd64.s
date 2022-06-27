// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2022 The Ebiten Authors

#include "textflag.h"

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
DATA ·syscall9XABI0(SB)/8, $·syscall9X(SB)
TEXT ·syscall9X(SB), NOSPLIT, $0
	PUSHQ BP
	MOVQ  SP, BP
	SUBQ  $32, SP
	MOVQ  DI, 24(BP)     // save the pointer
	MOVQ  (0*8)(DI), R10 // fn
	MOVQ  (2*8)(DI), SI  // a2
	MOVQ  (3*8)(DI), DX  // a3
	MOVQ  (4*8)(DI), CX  // a4
	MOVQ  (5*8)(DI), R8  // a5
	MOVQ  (6*8)(DI), R9  // a6
	MOVQ  (7*8)(DI), R11 // a7
	MOVQ  (8*8)(DI), R12 // a8
	MOVQ  (9*8)(DI), R13 // a9
	MOVQ  (1*8)(DI), DI  // a1

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

	MOVQ 24(BP), DI     // get the pointer back
	MOVQ AX, (10*8)(DI) // r1
	MOVQ DX, (11*8)(DI) // r2

	// Standard libc functions return -1 on error
	// and set errno.
	CMPQ AX, $-1
	JNE  ok

	// Get error code from libc.
	CALL    libc_error(SB)
	MOVLQSX (AX), AX
	MOVQ    (SP), DI
	MOVQ    AX, (12*8)(DI) // err

ok:
	XORL AX, AX  // no error (it's ignored anyway)
	ADDQ $32, SP
	MOVQ BP, SP
	POPQ BP
	RET
