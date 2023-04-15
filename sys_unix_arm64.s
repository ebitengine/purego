// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2023 The Ebitengine Authors

//go:build darwin || (!cgo && linux)

#include "textflag.h"
#include "go_asm.h"
#include "funcdata.h"
#include "internal/abi/abi_arm64.h"

TEXT callbackasm1(SB), NOSPLIT|NOFRAME, $0
	NO_LOCAL_POINTERS

	// On entry, the trampoline in zcallback_darwin_arm64.s left
	// the callback index in R12 (which is volatile in the C ABI).

	// Save callback register arguments R0-R7 and F0-F7.
	// We do this at the top of the frame so they're contiguous with stack arguments.
	SUB   $(16*8), RSP, R14
	FMOVD F0, (0*8)(R14)
	FMOVD F1, (1*8)(R14)
	FMOVD F2, (2*8)(R14)
	FMOVD F3, (3*8)(R14)
	FMOVD F4, (4*8)(R14)
	FMOVD F5, (5*8)(R14)
	FMOVD F6, (6*8)(R14)
	FMOVD F7, (7*8)(R14)
	STP   (R0, R1), (8*8)(R14)
	STP   (R2, R3), (10*8)(R14)
	STP   (R4, R5), (12*8)(R14)
	STP   (R6, R7), (14*8)(R14)

	// Adjust SP by frame size.
	// crosscall2 clobbers FP in the frame record so only save/restore SP.
	SUB  $(28*8), RSP
	MOVD R30, (RSP)

	// Create a struct callbackArgs on our stack.
	ADD  $(callbackArgs__size + 3*8), RSP, R13
	MOVD R12, callbackArgs_index(R13)          // callback index
	MOVD R14, R0
	MOVD R0, callbackArgs_args(R13)            // address of args vector
	MOVD $0, R0
	MOVD R0, callbackArgs_result(R13)          // result

	// Move parameters into registers
	// Get the ABIInternal function pointer
	// without <ABIInternal> by using a closure.
	MOVD ·callbackWrap_call(SB), R0
	MOVD (R0), R0                   // fn unsafe.Pointer
	MOVD R13, R1                    // frame (&callbackArgs{...})
	MOVD $0, R3                     // ctxt uintptr

	BL crosscall2(SB)

	// Get callback result.
	ADD  $(callbackArgs__size + 3*8), RSP, R13
	MOVD callbackArgs_result(R13), R0

	// Restore SP
	MOVD (RSP), R30
	ADD  $(28*8), RSP

	RET
