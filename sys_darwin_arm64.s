// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2022 The Ebitengine Authors

#include "textflag.h"
#include "go_asm.h"
#include "funcdata.h"
#include "internal/abi/abi_arm64.h"

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

	MOVD 8(RSP), R2              // pop structure pointer
	ADD  $16, RSP
	MOVD R0, syscall9Args_r1(R2) // save r1
	MOVD R1, syscall9Args_r2(R2) // save r2
	RET

// runtime·cgocallback expects a call to the ABIInternal function
// However, the tag <ABIInternal> is only available in the runtime :(
// This is a small wrapper function that copies both whatever is in the register
// and is on the stack and places both on the stack. It then calls callbackWrapPicker
// which will choose which parameter should be used depending on the version of Go.
// It then calls the real version of callbackWrap
TEXT callbackWrapInternal<>(SB), NOSPLIT, $0-0
	MOVD R0, 16(RSP)
	B    ·callbackWrapPicker(SB)
	RET

TEXT callbackasm1(SB), NOSPLIT, $208-0
	NO_LOCAL_POINTERS

	// On entry, the trampoline in zcallback_windows_arm64.s left
	// the callback index in R12 (which is volatile in the C ABI).

	// Save callback register arguments R0-R7.
	// We do this at the top of the frame so they're contiguous with stack arguments.
	// The 7*8 setting up R14 looks like a bug but is not: the eighth word
	// is the space the assembler reserved for our caller's frame pointer,
	// but we are not called from Go so that space is ours to use,
	// and we must to be contiguous with the stack arguments.
	MOVD $arg0-(7*8)(SP), R14
	STP  (R0, R1), (0*8)(R14)
	STP  (R2, R3), (2*8)(R14)
	STP  (R4, R5), (4*8)(R14)
	STP  (R6, R7), (6*8)(R14)

	// Create a struct callbackArgs on our stack.
	MOVD $cbargs-(18*8+callbackArgs__size)(SP), R13
	MOVD R12, callbackArgs_index(R13)               // callback index
	MOVD R14, R0
	MOVD R0, callbackArgs_args(R13)                 // address of args vector
	MOVD $0, R0
	MOVD R0, callbackArgs_result(R13)               // result

	// Move parameters into registers
	MOVD $callbackWrapInternal<>(SB), R0 // fn unsafe.Pointer
	MOVD R13, R1                         // frame (&callbackArgs{...})
	MOVD $0, R3                          // ctxt uintptr

	// We still need to save all callee save register as before, and then
	//  push 3 args for fn (R0, R1, R3), skipping R2.
	// Also note that at procedure entry in gc world, 8(RSP) will be the
	// first arg.
	SUB  $(8*24), RSP
	STP  (R0, R1), (8*1)(RSP)
	MOVD R3, (8*3)(RSP)

	// Push C callee-save registers R19-R28.
	// LR, FP already saved.
	SAVE_R19_TO_R28(8*4)
	SAVE_F8_TO_F15(8*14)
	STP (R29, R30), (8*22)(RSP)

	// Initialize Go ABI environment
	BL runtime·load_g(SB)
	BL runtime·cgocallback(SB)

	RESTORE_R19_TO_R28(8*4)
	RESTORE_F8_TO_F15(8*14)
	LDP (8*22)(RSP), (R29, R30)

	ADD $(8*24), RSP

	// Get callback result.
	MOVD $cbargs-(18*8+callbackArgs__size)(SP), R13
	MOVD callbackArgs_result(R13), R0

	RET
