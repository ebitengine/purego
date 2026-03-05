// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 The Ebitengine Authors

//go:build linux

#include "textflag.h"
#include "go_asm.h"
#include "funcdata.h"

// PPC64LE ELFv2 ABI:
// - Integer args: R3-R10 (8 registers)
// - Float args: F1-F8 (8 registers)
// - Return: R3 (integer), F1 (float)
// - Stack pointer: R1
// - Link register: LR (special)
// - TOC pointer: R2 (must preserve)

// Stack layout for ELFv2 ABI (aligned to 16 bytes):
// From callee's perspective when we call BL (CTR):
//  0(R1)  - back chain (our old R1)
//  8(R1)  - CR save word (optional)
// 16(R1)  - LR save (optional, we save it)
// 24(R1)  - Reserved (compilers)
// 32(R1)  - Parameter save area start (8 * 8 = 64 bytes for R3-R10)
// 96(R1)  - First stack arg (a9) - this is where callee looks
// 104(R1) - Second stack arg (a10)
// 112-280 - Stack args a11-a32
// 288(R1) - TOC save (outside parameter save area)
// 296(R1) - saved args pointer
// Total: 304 bytes

#define STACK_SIZE 304
#define LR_SAVE    16
#define TOC_SAVE   288
#define ARGP_SAVE  296

GLOBL ·syscall15XABI0(SB), NOPTR|RODATA, $8
DATA ·syscall15XABI0(SB)/8, $syscall15X(SB)

TEXT syscall15X(SB), NOSPLIT, $0
	// Prologue: create stack frame
	// R3 contains the args pointer on entry
	MOVD R1, R12          // save old SP
	SUB  $STACK_SIZE, R1  // allocate stack frame
	MOVD R12, 0(R1)       // save back chain
	MOVD LR, R12
	MOVD R12, LR_SAVE(R1) // save LR
	MOVD R2, TOC_SAVE(R1) // save TOC

	// Save args pointer (in R3)
	MOVD R3, ARGP_SAVE(R1)

	// R11 := args pointer (syscall15Args*)
	MOVD R3, R11

	// Load float args into F1-F8
	FMOVD syscall15Args_f1(R11), F1
	FMOVD syscall15Args_f2(R11), F2
	FMOVD syscall15Args_f3(R11), F3
	FMOVD syscall15Args_f4(R11), F4
	FMOVD syscall15Args_f5(R11), F5
	FMOVD syscall15Args_f6(R11), F6
	FMOVD syscall15Args_f7(R11), F7
	FMOVD syscall15Args_f8(R11), F8

	// Load integer args into R3-R10
	MOVD syscall15Args_a1(R11), R3
	MOVD syscall15Args_a2(R11), R4
	MOVD syscall15Args_a3(R11), R5
	MOVD syscall15Args_a4(R11), R6
	MOVD syscall15Args_a5(R11), R7
	MOVD syscall15Args_a6(R11), R8
	MOVD syscall15Args_a7(R11), R9
	MOVD syscall15Args_a8(R11), R10

	// Spill a9-a32 onto the stack (stack parameters start at 96(R1))
	// Per ELFv2: parameter save area is 32-95, stack args start at 96
	MOVD ARGP_SAVE(R1), R11          // reload args pointer
	MOVD syscall15Args_a9(R11), R12
	MOVD R12, 96(R1)                 // a9 at 96(R1)
	MOVD syscall15Args_a10(R11), R12
	MOVD R12, 104(R1)                // a10 at 104(R1)
	MOVD syscall15Args_a11(R11), R12
	MOVD R12, 112(R1)                // a11 at 112(R1)
	MOVD syscall15Args_a12(R11), R12
	MOVD R12, 120(R1)                // a12 at 120(R1)
	MOVD syscall15Args_a13(R11), R12
	MOVD R12, 128(R1)                // a13 at 128(R1)
	MOVD syscall15Args_a14(R11), R12
	MOVD R12, 136(R1)                // a14 at 136(R1)
	MOVD syscall15Args_a15(R11), R12
	MOVD R12, 144(R1)                // a15 at 144(R1)
	MOVD syscall15Args_a16(R11), R12
	MOVD R12, 152(R1)                // a16 at 152(R1)
	MOVD syscall15Args_a17(R11), R12
	MOVD R12, 160(R1)                // a17 at 160(R1)
	MOVD syscall15Args_a18(R11), R12
	MOVD R12, 168(R1)                // a18 at 168(R1)
	MOVD syscall15Args_a19(R11), R12
	MOVD R12, 176(R1)                // a19 at 176(R1)
	MOVD syscall15Args_a20(R11), R12
	MOVD R12, 184(R1)                // a20 at 184(R1)
	MOVD syscall15Args_a21(R11), R12
	MOVD R12, 192(R1)                // a21 at 192(R1)
	MOVD syscall15Args_a22(R11), R12
	MOVD R12, 200(R1)                // a22 at 200(R1)
	MOVD syscall15Args_a23(R11), R12
	MOVD R12, 208(R1)                // a23 at 208(R1)
	MOVD syscall15Args_a24(R11), R12
	MOVD R12, 216(R1)                // a24 at 216(R1)
	MOVD syscall15Args_a25(R11), R12
	MOVD R12, 224(R1)                // a25 at 224(R1)
	MOVD syscall15Args_a26(R11), R12
	MOVD R12, 232(R1)                // a26 at 232(R1)
	MOVD syscall15Args_a27(R11), R12
	MOVD R12, 240(R1)                // a27 at 240(R1)
	MOVD syscall15Args_a28(R11), R12
	MOVD R12, 248(R1)                // a28 at 248(R1)
	MOVD syscall15Args_a29(R11), R12
	MOVD R12, 256(R1)                // a29 at 256(R1)
	MOVD syscall15Args_a30(R11), R12
	MOVD R12, 264(R1)                // a30 at 264(R1)
	MOVD syscall15Args_a31(R11), R12
	MOVD R12, 272(R1)                // a31 at 272(R1)
	MOVD syscall15Args_a32(R11), R12
	MOVD R12, 280(R1)                // a32 at 280(R1)

	// Call function: load fn and call
	MOVD syscall15Args_fn(R11), R12
	MOVD R12, CTR
	BL   (CTR)

	// Restore TOC after call
	MOVD TOC_SAVE(R1), R2

	// Restore args pointer for storing results
	MOVD ARGP_SAVE(R1), R11

	// Store integer results back (R3, R4)
	MOVD R3, syscall15Args_a1(R11)
	MOVD R4, syscall15Args_a2(R11)

	// Store float return values (F1-F4)
	FMOVD F1, syscall15Args_f1(R11)
	FMOVD F2, syscall15Args_f2(R11)
	FMOVD F3, syscall15Args_f3(R11)
	FMOVD F4, syscall15Args_f4(R11)

	// Epilogue: restore and return
	MOVD LR_SAVE(R1), R12
	MOVD R12, LR
	ADD  $STACK_SIZE, R1
	RET
