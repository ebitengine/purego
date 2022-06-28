// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2022 The Ebiten Authors

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
	SUB  $16, RSP   // push structure pointer
	MOVD R0, 8(RSP)

	MOVD syscall9Args_fn(R0), R12 // fn
	MOVD syscall9Args_a2(R0), R1  // a2
	MOVD syscall9Args_a3(R0), R2  // a3
	MOVD syscall9Args_a4(R0), R3  // a4
	MOVD syscall9Args_a5(R0), R4  // a5
	MOVD syscall9Args_a6(R0), R5  // a6
	MOVD syscall9Args_a7(R0), R6  // a7
	MOVD syscall9Args_a8(R0), R7  // a8
	MOVD syscall9Args_a9(R0), R8  // a9
	MOVD syscall9Args_a1(R0), R0  // a1

	// these may be float arguments
	// so we put them also where C expects floats
	FMOVD R0, F0 // a1
	FMOVD R1, F1 // a2
	FMOVD R2, F2 // a3
	FMOVD R3, F3 // a4
	FMOVD R4, F4 // a5
	FMOVD R5, F5 // a6
	FMOVD R6, F6 // a7
	FMOVD R7, F7 // a8

	MOVD R8, (RSP) // push a9 onto stack

	BL (R12)

	MOVD 8(RSP), R2               // pop structure pointer
	ADD  $16, RSP
	MOVD R0, syscall9Args_r1(R2)  // save r1
	MOVD R1, syscall9Args_r2(R2)  // save r2
	CMP  $-1, R0
	BNE  ok
	SUB  $16, RSP                 // push structure pointer
	MOVD R2, (RSP)
	BL   libc_error(SB)
	MOVW (R0), R0
	MOVD (RSP), R2                // pop structure pointer
	ADD  $16, RSP
	MOVD R0, syscall9Args_err(R2) // save err

ok:
	RET
