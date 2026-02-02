// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 The Ebitengine Authors

//go:build linux

#include "textflag.h"
#include "go_asm.h"
#include "funcdata.h"

// PPC64LE ELFv2 ABI callbackasm1 implementation
// On entry, R11 contains the callback index (set by callbackasm)
//
// ELFv2 stack frame layout requirements:
//   0(R1)   - back chain (pointer to caller's frame)
//   8(R1)   - CR save area (optional)
//  16(R1)   - LR save area (for callee to save caller's LR)
//  24(R1)   - TOC save area (if needed)
//  32(R1)+  - parameter save area / local variables
//
// Our frame (total 288 bytes, 16-byte aligned):
//  32(R1)   - saved R31 (Go assembler uses it)
//  40(R1)   - callbackArgs struct (24 bytes: index, args ptr, result)
//  64(R1)   - args array: floats (64) + ints (64) + stack args (56) = 184 bytes
// Total: 64 + 184 = 248, add 32 for fixed area = 280, round to 288

#define FRAME_SIZE     288
#define SAVE_R31       32
#define CB_ARGS        40
#define ARGS_ARRAY     64
#define FLOAT_OFF      0
#define INT_OFF        64
#define STACK_OFF      128

TEXT callbackasm1(SB), NOSPLIT|NOFRAME, $0
	NO_LOCAL_POINTERS

	// On entry, the trampoline in zcallback_ppc64le.s left
	// the callback index in R11.

	// Per ELFv2 ABI, save LR to caller's frame BEFORE allocating our frame
	MOVD LR, R0
	MOVD R0, 16(R1)

	// Allocate our stack frame (with back chain via MOVDU)
	MOVDU R1, -FRAME_SIZE(R1)

	// Save R31 - Go assembler uses it for MOVD from SB (like arm64's R27)
	MOVD R31, SAVE_R31(R1)

	// Save R11 (callback index) immediately - it's volatile and will be clobbered!
	// Store it in the callbackArgs struct's index field now.
	MOVD R11, (CB_ARGS+0)(R1)

	// Save callback arguments to args array.
	// Layout: floats first (F1-F8), then ints (R3-R10), then stack args
	FMOVD F1, (ARGS_ARRAY+FLOAT_OFF+0*8)(R1)
	FMOVD F2, (ARGS_ARRAY+FLOAT_OFF+1*8)(R1)
	FMOVD F3, (ARGS_ARRAY+FLOAT_OFF+2*8)(R1)
	FMOVD F4, (ARGS_ARRAY+FLOAT_OFF+3*8)(R1)
	FMOVD F5, (ARGS_ARRAY+FLOAT_OFF+4*8)(R1)
	FMOVD F6, (ARGS_ARRAY+FLOAT_OFF+5*8)(R1)
	FMOVD F7, (ARGS_ARRAY+FLOAT_OFF+6*8)(R1)
	FMOVD F8, (ARGS_ARRAY+FLOAT_OFF+7*8)(R1)

	MOVD R3, (ARGS_ARRAY+INT_OFF+0*8)(R1)
	MOVD R4, (ARGS_ARRAY+INT_OFF+1*8)(R1)
	MOVD R5, (ARGS_ARRAY+INT_OFF+2*8)(R1)
	MOVD R6, (ARGS_ARRAY+INT_OFF+3*8)(R1)
	MOVD R7, (ARGS_ARRAY+INT_OFF+4*8)(R1)
	MOVD R8, (ARGS_ARRAY+INT_OFF+5*8)(R1)
	MOVD R9, (ARGS_ARRAY+INT_OFF+6*8)(R1)
	MOVD R10, (ARGS_ARRAY+INT_OFF+7*8)(R1)

	// Copy stack arguments from caller's frame.
	// Caller's stack args start at caller_R1 + 96 (ELFv2 ABI).
	// Our R1 = caller_R1 - FRAME_SIZE, so args at R1 + FRAME_SIZE + 96
	MOVD (FRAME_SIZE+96)(R1), R12
	MOVD R12, (ARGS_ARRAY+STACK_OFF+0*8)(R1)
	MOVD (FRAME_SIZE+104)(R1), R12
	MOVD R12, (ARGS_ARRAY+STACK_OFF+1*8)(R1)
	MOVD (FRAME_SIZE+112)(R1), R12
	MOVD R12, (ARGS_ARRAY+STACK_OFF+2*8)(R1)
	MOVD (FRAME_SIZE+120)(R1), R12
	MOVD R12, (ARGS_ARRAY+STACK_OFF+3*8)(R1)
	MOVD (FRAME_SIZE+128)(R1), R12
	MOVD R12, (ARGS_ARRAY+STACK_OFF+4*8)(R1)
	MOVD (FRAME_SIZE+136)(R1), R12
	MOVD R12, (ARGS_ARRAY+STACK_OFF+5*8)(R1)
	MOVD (FRAME_SIZE+144)(R1), R12
	MOVD R12, (ARGS_ARRAY+STACK_OFF+6*8)(R1)

	// Finish setting up callbackArgs struct at CB_ARGS(R1)
	// struct { index uintptr; args unsafe.Pointer; result uintptr }
	// Note: index was already saved earlier (R11 is volatile)
	ADD  $ARGS_ARRAY, R1, R12
	MOVD R12, (CB_ARGS+8)(R1) // address of args vector
	MOVD $0, (CB_ARGS+16)(R1) // result = 0

	// Call crosscall2 with arguments in registers:
	// R3 = fn (from callbackWrap_call closure)
	// R4 = frame (address of callbackArgs)
	// R6 = ctxt (0)
	MOVD ·callbackWrap_call(SB), R3
	MOVD (R3), R3                   // dereference closure to get fn
	ADD  $CB_ARGS, R1, R4           // frame = &callbackArgs
	MOVD $0, R6                     // ctxt = 0

	BL crosscall2(SB)

	// Get callback result into R3
	MOVD (CB_ARGS+16)(R1), R3

	// Restore R31
	MOVD SAVE_R31(R1), R31

	// Deallocate frame
	ADD $FRAME_SIZE, R1

	// Restore LR from caller's frame (per ELFv2, it was saved at 16(old_R1))
	MOVD 16(R1), R0
	MOVD R0, LR

	RET
