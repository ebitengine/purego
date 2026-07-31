// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 The Ebitengine Authors

//go:build linux || windows

#include "textflag.h"
#include "go_asm.h"
#include "funcdata.h"

#define STACK_SIZE 160
#define PTR_ADDRESS (STACK_SIZE - 4)

// syscallX calls a function in libc on behalf of the syscall package.
// syscallX takes a pointer to a struct like:
// struct {
//	fn    uintptr
//	a1    uintptr
//  ...
//	a32   uintptr
//	f1    uintptr
//	...
//	f16   uintptr
//	floatReturn uintptr
// }
// syscallX must be called on the g0 stack with the
// C calling convention (use libcCall).
//
// On the i386 System V and Windows ABIs, all arguments are passed on the stack.
// Return value is in EAX (and EDX for 64-bit values).
GLOBL ·syscallXABI0(SB), NOPTR|RODATA, $4
DATA ·syscallXABI0(SB)/4, $syscallX(SB)
TEXT syscallX(SB), NOSPLIT|NOFRAME, $0-0
	// Called via C calling convention: argument pointer at 4(SP)
	// NOT via Go calling convention
	// On i386, the first argument is at 4(SP) after CALL pushes return address
	MOVL 4(SP), AX // get pointer to syscallArgs

	// Save callee-saved registers
	PUSHL BP
	PUSHL BX
	PUSHL SI
	PUSHL DI

	MOVL AX, BX // save args pointer in BX

	// Allocate stack space for C function arguments
	// i386: all 32 args on stack = 32 * 4 = 128 bytes
	// Plus 16 bytes for alignment and local storage
	SUBL $STACK_SIZE, SP
	MOVL SP, DI // save the stack pointer; the target may use cdecl or stdcall

	// Load function pointer
	MOVL syscallArgs_fn(BX), AX
	MOVL AX, (PTR_ADDRESS-4)(SP)  // save fn pointer

	// Push all integer arguments onto the stack (a1-a32)
	// i386 ABIs pass arguments right-to-left, but we're
	// setting up the stack from low to high addresses
	MOVL syscallArgs_a1(BX), AX
	MOVL AX, 0(SP)
	MOVL syscallArgs_a2(BX), AX
	MOVL AX, 4(SP)
	MOVL syscallArgs_a3(BX), AX
	MOVL AX, 8(SP)
	MOVL syscallArgs_a4(BX), AX
	MOVL AX, 12(SP)
	MOVL syscallArgs_a5(BX), AX
	MOVL AX, 16(SP)
	MOVL syscallArgs_a6(BX), AX
	MOVL AX, 20(SP)
	MOVL syscallArgs_a7(BX), AX
	MOVL AX, 24(SP)
	MOVL syscallArgs_a8(BX), AX
	MOVL AX, 28(SP)
	MOVL syscallArgs_a9(BX), AX
	MOVL AX, 32(SP)
	MOVL syscallArgs_a10(BX), AX
	MOVL AX, 36(SP)
	MOVL syscallArgs_a11(BX), AX
	MOVL AX, 40(SP)
	MOVL syscallArgs_a12(BX), AX
	MOVL AX, 44(SP)
	MOVL syscallArgs_a13(BX), AX
	MOVL AX, 48(SP)
	MOVL syscallArgs_a14(BX), AX
	MOVL AX, 52(SP)
	MOVL syscallArgs_a15(BX), AX
	MOVL AX, 56(SP)
	MOVL syscallArgs_a16(BX), AX
	MOVL AX, 60(SP)
	MOVL syscallArgs_a17(BX), AX
	MOVL AX, 64(SP)
	MOVL syscallArgs_a18(BX), AX
	MOVL AX, 68(SP)
	MOVL syscallArgs_a19(BX), AX
	MOVL AX, 72(SP)
	MOVL syscallArgs_a20(BX), AX
	MOVL AX, 76(SP)
	MOVL syscallArgs_a21(BX), AX
	MOVL AX, 80(SP)
	MOVL syscallArgs_a22(BX), AX
	MOVL AX, 84(SP)
	MOVL syscallArgs_a23(BX), AX
	MOVL AX, 88(SP)
	MOVL syscallArgs_a24(BX), AX
	MOVL AX, 92(SP)
	MOVL syscallArgs_a25(BX), AX
	MOVL AX, 96(SP)
	MOVL syscallArgs_a26(BX), AX
	MOVL AX, 100(SP)
	MOVL syscallArgs_a27(BX), AX
	MOVL AX, 104(SP)
	MOVL syscallArgs_a28(BX), AX
	MOVL AX, 108(SP)
	MOVL syscallArgs_a29(BX), AX
	MOVL AX, 112(SP)
	MOVL syscallArgs_a30(BX), AX
	MOVL AX, 116(SP)
	MOVL syscallArgs_a31(BX), AX
	MOVL AX, 120(SP)
	MOVL syscallArgs_a32(BX), AX
	MOVL AX, 124(SP)

	// Call the C function
	MOVL (PTR_ADDRESS-4)(SP), AX
	CLD
	CALL AX

	// BX and DI are callee-saved in both the i386 SysV and Windows ABIs.
	// Restore SP explicitly because a stdcall target may have popped its args.
	MOVL AX, syscallArgs_a1(BX) // return value r1
	MOVL DX, syscallArgs_a2(BX) // return value r2 (for 64-bit returns)

	// Save an x87 return only when RegisterFunc requested one. Mode 1 is a
	// scalar float result and therefore requires ST(0). Mode 2 is a one-field
	// float struct: MinGW returns it in ST(0), while MSVC uses EAX/EDX, so first
	// use FXAM to distinguish an x87 result from an empty register stack.
	CMPL syscallArgs_floatReturn(BX), $0
	JE no_float_return
	CMPL syscallArgs_floatReturn(BX), $2
	JNE save_float_return
	FXAM
	FSTSW AX
	ANDL $0x4500, AX // C3, C2, C0
	CMPL AX, $0x4100 // empty x87 register
	JNE save_float_return
	MOVL $0, syscallArgs_floatReturn(BX)
	JMP no_float_return

save_float_return:
	FMOVDP F0, syscallArgs_f1(BX)
	MOVL $1, syscallArgs_floatReturn(BX)

no_float_return:

	// Clean up stack
	MOVL DI, SP
	ADDL $STACK_SIZE, SP

	// Restore callee-saved registers
	POPL DI
	POPL SI
	POPL BX
	POPL BP

	RET
