// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2022 The Ebitengine Authors

//go:build darwin || freebsd || linux

#include "textflag.h"
#include "go_asm.h"
#include "funcdata.h"
#include "abi_arm64.h"

#define STACK_SIZE 64
#define PTR_ADDRESS (STACK_SIZE - 8)

// syscall15X calls a function in libc on behalf of the syscall package.
// syscall15X takes a pointer to a struct like:
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
//	a10    uintptr
//	a11    uintptr
//	a12    uintptr
//	a13    uintptr
//	a14    uintptr
//	a15    uintptr
//	r1    uintptr
//	r2    uintptr
//	err   uintptr
// }
// syscall15X must be called on the g0 stack with the
// C calling convention (use libcCall).
GLOBL ·syscall15XABI0(SB), NOPTR|RODATA, $8
DATA ·syscall15XABI0(SB)/8, $syscall15X(SB)
TEXT syscall15X(SB), NOSPLIT, $0
	SUB  $STACK_SIZE, RSP     // push structure pointer
	MOVD R0, PTR_ADDRESS(RSP)
	MOVD R0, R9

	FMOVD syscall15Args_f1(R9), F0 // f1
	FMOVD syscall15Args_f2(R9), F1 // f2
	FMOVD syscall15Args_f3(R9), F2 // f3
	FMOVD syscall15Args_f4(R9), F3 // f4
	FMOVD syscall15Args_f5(R9), F4 // f5
	FMOVD syscall15Args_f6(R9), F5 // f6
	FMOVD syscall15Args_f7(R9), F6 // f7
	FMOVD syscall15Args_f8(R9), F7 // f8

	MOVD syscall15Args_a1(R9), R0 // a1
	MOVD syscall15Args_a2(R9), R1 // a2
	MOVD syscall15Args_a3(R9), R2 // a3
	MOVD syscall15Args_a4(R9), R3 // a4
	MOVD syscall15Args_a5(R9), R4 // a5
	MOVD syscall15Args_a6(R9), R5 // a6
	MOVD syscall15Args_a7(R9), R6 // a7
	MOVD syscall15Args_a8(R9), R7 // a8

	MOVD syscall15Args_a9(R9), R10
	MOVD R10, 0(RSP)                // push a9 onto stack
	MOVD syscall15Args_a10(R9), R10
	MOVD R10, 8(RSP)                // push a10 onto stack
	MOVD syscall15Args_a11(R9), R10
	MOVD R10, 16(RSP)               // push a11 onto stack
	MOVD syscall15Args_a12(R9), R10
	MOVD R10, 24(RSP)               // push a12 onto stack
	MOVD syscall15Args_a13(R9), R10
	MOVD R10, 32(RSP)               // push a13 onto stack
	MOVD syscall15Args_a14(R9), R10
	MOVD R10, 40(RSP)               // push a14 onto stack
	MOVD syscall15Args_a15(R9), R10
	MOVD R10, 48(RSP)               // push a15 onto stack

	MOVD syscall15Args_fn(R9), R10 // fn
	BL   (R10)

	MOVD  PTR_ADDRESS(RSP), R2     // pop structure pointer
	ADD   $STACK_SIZE, RSP
	MOVD  R0, syscall15Args_r1(R2) // save r1
	FMOVD F0, syscall15Args_r2(R2) // save r2
	RET

TEXT callbackasm1(SB), NOSPLIT|NOFRAME, $0
	NO_LOCAL_POINTERS

	// On entry, the trampoline in zcallback_darwin_arm64.s left
	// the callback index in R12 (which is volatile in the C ABI).

	// Save callback register arguments R0-R7 and F0-F7.
	// We do this at the top of the frame so they're contiguous with stack arguments.
	SUB   $(16*8), RSP, R14
	FSTPD (F0, F1), (0*8)(R14)
	FSTPD (F2, F3), (2*8)(R14)
	FSTPD (F4, F5), (4*8)(R14)
	FSTPD (F6, F7), (6*8)(R14)
	STP   (R0, R1), (8*8)(R14)
	STP   (R2, R3), (10*8)(R14)
	STP   (R4, R5), (12*8)(R14)
	STP   (R6, R7), (14*8)(R14)

	// Adjust SP by frame size.
	SUB $(26*8), RSP

	// It is important to save R27 because the go assembler
	// uses it for move instructions for a variable.
	// This line:
	// MOVD ·callbackWrap_call(SB), R0
	// Creates the instructions:
	// ADRP 14335(PC), R27
	// MOVD 388(27), R0
	// R27 is a callee saved register so we are responsible
	// for ensuring its value doesn't change. So save it and
	// restore it at the end of this function.
	// R30 is the link register. crosscall2 doesn't save it
	// so it's saved here.
	STP (R27, R30), 0(RSP)

	// Create a struct callbackArgs on our stack.
	MOVD $(callbackArgs__size)(RSP), R13
	MOVD R12, callbackArgs_index(R13)    // callback index
	MOVD R14, callbackArgs_args(R13)     // address of args vector
	MOVD ZR, callbackArgs_result(R13)    // result

	// Move parameters into registers
	// Get the ABIInternal function pointer
	// without <ABIInternal> by using a closure.
	MOVD ·callbackWrap_call(SB), R0
	MOVD (R0), R0                   // fn unsafe.Pointer
	MOVD R13, R1                    // frame (&callbackArgs{...})
	MOVD $0, R3                     // ctxt uintptr

	BL crosscall2(SB)

	// Get callback result.
	MOVD $(callbackArgs__size)(RSP), R13
	MOVD callbackArgs_result(R13), R0

	// Restore LR and R27
	LDP 0(RSP), (R27, R30)
	ADD $(26*8), RSP

	RET

