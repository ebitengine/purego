// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 The Ebitengine Authors

//go:build windows && 386

#include "textflag.h"

// callbackasm1 receives an index in CX from the callbackasm table. It calls a
// pointer-sized standard-library callback which decodes the original typed
// stack frame, then restores the native Windows/386 result registers.
TEXT callbackasm1(SB), NOSPLIT|NOFRAME, $0
	// Preserve the non-volatile Windows x86 registers.
	PUSHL BP
	PUSHL BX
	PUSHL SI
	PUSHL DI

	// Layout after this allocation:
	//   0..11   private bridge arguments
	//   12..35  win386CallbackResult
	//   36..51  saved DI, SI, BX, BP
	//   52       original return address
	//   56..     original callback arguments
	SUBL $36, SP
	MOVL CX, 0(SP)
	LEAL 56(SP), AX
	MOVL AX, 4(SP)
	LEAL 12(SP), AX
	MOVL AX, 8(SP)

	MOVL ·win386CallbackBridge(SB), AX
	CALL AX // stdcall bridge pops its three private arguments
	SUBL $12, SP // return to the base of our local frame

	// Move the original return address to the post-cleanup stack position.
	// This supports a variable stdcall pop while leaving EDX available for a
	// 64-bit result. For cdecl, stackPop is zero and this rewrites the same slot.
	MOVL 32(SP), CX
	MOVL 52(SP), DI
	LEAL 52(SP), SI
	ADDL CX, SI
	MOVL DI, 0(SI)

	MOVL 12(SP), AX
	MOVL 16(SP), DX
	MOVL 28(SP), DI
	CMPL DI, $1
	JE callback_float32
	CMPL DI, $2
	JE callback_float64
	JMP callback_result_ready

callback_float32:
	FMOVF 20(SP), F0
	JMP callback_result_ready

callback_float64:
	FMOVD 20(SP), F0

callback_result_ready:
	ADDL $36, SP
	POPL DI
	POPL SI
	POPL BX
	POPL BP
	ADDL CX, SP
	RET
