// Copyright 2022 The Ebiten Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

#include "textflag.h"

// syscallXF calls a function in libc on behalf of the syscall package.
// syscallXF takes a pointer to a struct like:
// struct {
//	fn    uintptr
//	a1    uintptr
//	a2    uintptr
//	a3    uintptr
//	f1    float64
//	f2    float64
//	f3    float64
//	r1    uintptr
//	r2    uintptr
//	err   uintptr
// }
// syscallXF must be called on the g0 stack with the
// C calling convention (use libcCall).
GLOBL ·syscallXFABI0(SB), NOPTR|RODATA, $8
DATA ·syscallXFABI0(SB)/8, $·syscallXF(SB)
TEXT ·syscallXF(SB),NOSPLIT,$0
	PUSHQ	BP
	MOVQ	SP, BP
	SUBQ	$16, SP
	MOVQ	(0*8)(DI), R10 // fn
	MOVQ	(2*8)(DI), SI // a2
	MOVQ	(3*8)(DI), DX // a3
	MOVSD	(4*8)(DI), X0 // f1
	MOVSD	(5*8)(DI), X1 // f2
	MOVSD	(6*8)(DI), X2 // f3
	MOVQ	DI, (SP)
	MOVQ	(1*8)(DI), DI // a1
	XORL	AX, AX	      // vararg: say "no float args"

	CALL	R10

	MOVQ	(SP), DI
	MOVQ	AX, (7*8)(DI) // r1
	MOVQ	DX, (8*8)(DI) // r2

	// Standard libc functions return -1 on error
	// and set errno.
	CMPQ	AX, $-1
	JNE	ok

	// Get error code from libc.
	CALL	libc_error(SB)
	MOVLQSX	(AX), AX
	MOVQ	(SP), DI
	MOVQ	AX, (9*8)(DI) // err

ok:
	XORL	AX, AX        // no error (it's ignored anyway)
	MOVQ	BP, SP
	POPQ	BP
	RET

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
TEXT ·syscall9X(SB),NOSPLIT,$0
	PUSHQ	BP
	MOVQ	SP, BP
	SUBQ	$32, SP
	MOVQ	DI, -8(BP) // save the pointer
	MOVQ	(0*8)(DI), R10 // fn
	MOVQ	(2*8)(DI), SI  // a2
	MOVQ	(3*8)(DI), DX  // a3
	MOVQ	(4*8)(DI), CX  // a4
	MOVQ	(5*8)(DI), R8  // a5
	MOVQ	(6*8)(DI), R9  // a6
	MOVQ	(7*8)(DI), R11 // a7
	MOVQ	(8*8)(DI), R12 // a8
	MOVQ	(9*8)(DI), R13 // a9
	MOVQ	(1*8)(DI), DI  // a1
	MOVQ    R11, 0(SP)     // push a7
	MOVQ    R12, 8(SP)     // push a8
	MOVQ    R13, 16(SP)    // push a9
	XORL	AX, AX	       // vararg: say "no float args"

	CALL	R10

	MOVQ	-8(BP), DI     // get the pointer back
	MOVQ	AX, (10*8)(DI) // r1
	MOVQ	DX, (11*8)(DI) // r2

	// Standard libc functions return -1 on error
	// and set errno.
	CMPQ	AX, $-1
	JNE	ok

	// Get error code from libc.
	CALL	libc_error(SB)
	MOVLQSX	(AX), AX
	MOVQ	(SP), DI
	MOVQ	AX, (12*8)(DI) // err

ok:
	XORL	AX, AX        // no error (it's ignored anyway)
	ADDQ    $32, SP
	MOVQ	BP, SP
	POPQ	BP
	RET
