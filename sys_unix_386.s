// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 The Ebitengine Authors

//go:build linux

#include "textflag.h"
#include "go_asm.h"
#include "funcdata.h"

// callbackasm1 is the second part of the callback trampoline.
// On entry:
//   - CX contains the callback index (set by callbackasm)
//   - 0(SP) contains the return address to C caller
//   - 4(SP), 8(SP), ... contain C arguments (cdecl convention)
//
// i386 cdecl calling convention:
// - All arguments passed on stack
// - Return value in EAX (and EDX for 64-bit)
// - Caller cleans the stack
// - Callee must preserve: EBX, ESI, EDI, EBP
TEXT callbackasm1(SB), NOSPLIT, $0
	// Save the return address
	MOVL 0(SP), AX

	// Allocate stack frame (must be done carefully to preserve args access)
	// Layout:
	//   0-15: saved callee-saved registers (BX, SI, DI, BP)
	//   16-19: saved callback index
	//   20-23: saved return address
	//   24-35: callbackArgs struct (12 bytes)
	//   36-163: copy of C arguments (128 bytes for 32 args)
	// Total: 164 bytes, round up to 176 for alignment
	SUBL $176, SP

	// Save callee-saved registers
	MOVL BX, 0(SP)
	MOVL SI, 4(SP)
	MOVL DI, 8(SP)
	MOVL BP, 12(SP)

	// Save callback index and return address
	MOVL CX, 16(SP)
	MOVL AX, 20(SP)

	// Copy C arguments from original stack location to our frame
	// Original args start at 176+4(SP) = 180(SP) (past our frame + original return addr)
	// Copy to our frame at 36(SP)
	// Copy 32 arguments (128 bytes)
	MOVL 180(SP), AX
	MOVL AX, 36(SP)
	MOVL 184(SP), AX
	MOVL AX, 40(SP)
	MOVL 188(SP), AX
	MOVL AX, 44(SP)
	MOVL 192(SP), AX
	MOVL AX, 48(SP)
	MOVL 196(SP), AX
	MOVL AX, 52(SP)
	MOVL 200(SP), AX
	MOVL AX, 56(SP)
	MOVL 204(SP), AX
	MOVL AX, 60(SP)
	MOVL 208(SP), AX
	MOVL AX, 64(SP)
	MOVL 212(SP), AX
	MOVL AX, 68(SP)
	MOVL 216(SP), AX
	MOVL AX, 72(SP)
	MOVL 220(SP), AX
	MOVL AX, 76(SP)
	MOVL 224(SP), AX
	MOVL AX, 80(SP)
	MOVL 228(SP), AX
	MOVL AX, 84(SP)
	MOVL 232(SP), AX
	MOVL AX, 88(SP)
	MOVL 236(SP), AX
	MOVL AX, 92(SP)
	MOVL 240(SP), AX
	MOVL AX, 96(SP)
	MOVL 244(SP), AX
	MOVL AX, 100(SP)
	MOVL 248(SP), AX
	MOVL AX, 104(SP)
	MOVL 252(SP), AX
	MOVL AX, 108(SP)
	MOVL 256(SP), AX
	MOVL AX, 112(SP)
	MOVL 260(SP), AX
	MOVL AX, 116(SP)
	MOVL 264(SP), AX
	MOVL AX, 120(SP)
	MOVL 268(SP), AX
	MOVL AX, 124(SP)
	MOVL 272(SP), AX
	MOVL AX, 128(SP)
	MOVL 276(SP), AX
	MOVL AX, 132(SP)
	MOVL 280(SP), AX
	MOVL AX, 136(SP)
	MOVL 284(SP), AX
	MOVL AX, 140(SP)
	MOVL 288(SP), AX
	MOVL AX, 144(SP)
	MOVL 292(SP), AX
	MOVL AX, 148(SP)
	MOVL 296(SP), AX
	MOVL AX, 152(SP)
	MOVL 300(SP), AX
	MOVL AX, 156(SP)

	// Set up callbackArgs struct at 24(SP)
	// struct callbackArgs {
	//     index  uintptr  // offset 0
	//     args   *byte    // offset 4
	//     result uintptr  // offset 8
	// }
	MOVL 16(SP), AX // callback index
	MOVL AX, 24(SP) // callbackArgs.index
	LEAL 36(SP), AX // pointer to copied arguments
	MOVL AX, 28(SP) // callbackArgs.args
	MOVL $0, 32(SP) // callbackArgs.result = 0

	// Call crosscall2(fn, frame, 0, ctxt)
	// crosscall2 expects arguments on stack:
	//   0(SP) = fn
	//   4(SP) = frame (pointer to callbackArgs)
	//   8(SP) = ignored (was n)
	//   12(SP) = ctxt
	SUBL $16, SP

	MOVL ·callbackWrap_call(SB), AX
	MOVL (AX), AX                   // fn = *callbackWrap_call
	MOVL AX, 0(SP)                  // fn
	LEAL (24+16)(SP), AX            // &callbackArgs (adjusted for SUB $16)
	MOVL AX, 4(SP)                  // frame
	MOVL $0, 8(SP)                  // 0
	MOVL $0, 12(SP)                 // ctxt

	CALL crosscall2(SB)

	ADDL $16, SP

	// Get result from callbackArgs.result
	MOVL 32(SP), AX

	// Restore callee-saved registers
	MOVL 0(SP), BX
	MOVL 4(SP), SI
	MOVL 8(SP), DI
	MOVL 12(SP), BP

	// Restore return address and clean up
	MOVL 20(SP), CX // get return address
	ADDL $176, SP   // remove our frame
	MOVL CX, 0(SP)  // put return address back

	RET
