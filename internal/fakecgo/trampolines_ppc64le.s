// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 The Ebitengine Authors

//go:build !cgo && linux

#include "textflag.h"
#include "go_asm.h"
#include "abi_ppc64x.h"

// These trampolines map the C ABI to Go ABI and call into the Go equivalent functions.

TEXT x_cgo_init_trampoline(SB), NOSPLIT, $0-0
	MOVD ·x_cgo_init_call(SB), R12
	MOVD (R12), R12
	MOVD R12, CTR
	CALL CTR
	RET

TEXT x_cgo_thread_start_trampoline(SB), NOSPLIT, $0-0
	MOVD ·x_cgo_thread_start_call(SB), R12
	MOVD (R12), R12
	MOVD R12, CTR
	CALL CTR
	RET

TEXT x_cgo_setenv_trampoline(SB), NOSPLIT, $0-0
	MOVD ·x_cgo_setenv_call(SB), R12
	MOVD (R12), R12
	MOVD R12, CTR
	CALL CTR
	RET

TEXT x_cgo_unsetenv_trampoline(SB), NOSPLIT, $0-0
	MOVD ·x_cgo_unsetenv_call(SB), R12
	MOVD (R12), R12
	MOVD R12, CTR
	CALL CTR
	RET

TEXT x_cgo_notify_runtime_init_done_trampoline(SB), NOSPLIT, $0-0
	CALL ·x_cgo_notify_runtime_init_done(SB)
	RET

TEXT x_cgo_bindm_trampoline(SB), NOSPLIT, $0-0
	CALL ·x_cgo_bindm(SB)
	RET

// func setg_trampoline(setg uintptr, g uintptr)
TEXT ·setg_trampoline(SB), NOSPLIT, $16-16
	MOVD R31, 8(R1) // save R31

	MOVD setg+0(FP), R12
	MOVD newg+8(FP), R3

	MOVD R3, 16(R1) // save newg before call

	MOVD R12, CTR
	CALL CTR

	// Assign g directly instead of calling runtime·load_g
	// setg_gcc has already stored newg into TLS; put it in the g register too.
	MOVD 16(R1), g

	MOVD 8(R1), R31
	XOR  R0, R0, R0
	RET

TEXT threadentry_trampoline(SB), NOSPLIT|NOFRAME, $0-0
	// Called from C (pthread_create). Must save all C callee-saved registers.
	// Uses NOFRAME for proper ELFv2 backchain via MOVDU.
	MOVD LR, R0
	MOVD R0, 16(R1)
	MOVW CR, R0
	MOVD R0, 8(R1)

	// threadentry is ABIInternal, so the caller reserves the spill slot for
	// its register argument at 32(R1), after the fixed frame. The saved
	// registers start at 40(R1), and the frame is rounded up to 16 bytes.
	MOVDU R1, -336(R1)

	SAVE_GPR(40)
	SAVE_FPR(40+SAVE_GPR_SIZE)

	MOVD $0, R0

	MOVD ·threadentry_call(SB), R12
	MOVD (R12), R12
	MOVD R12, CTR
	CALL CTR

	RESTORE_FPR(40+SAVE_GPR_SIZE)
	RESTORE_GPR(40)

	ADD $336, R1

	MOVD 16(R1), R0
	MOVD R0, LR
	MOVD 8(R1), R0
	MOVW R0, CR
	RET

TEXT ·call5(SB), NOSPLIT|NOFRAME, $0-56
	MOVD LR, R20
	MOVW CR, R21

	// Load arguments from Go stack into C argument registers
	// Go placed args at 32(R1), 40(R1), etc.
	MOVD 32(R1), R12 // fn
	MOVD 40(R1), R3  // a1 → first C arg
	MOVD 48(R1), R4  // a2 → second C arg
	MOVD 56(R1), R5  // a3 → third C arg
	MOVD 64(R1), R6  // a4 → fourth C arg
	MOVD 72(R1), R7  // a5 → fifth C arg

	// Go frames are only 8-byte aligned, but ELFv2 requires R1 to be 16-byte
	// aligned at a call. Keep the original R1 in R22, a non-volatile register.
	// The back chain is stored before R1 moves so that 0(R1) is always valid.
	MOVD  R1, R22
	RLDCR $0, R1, $~15, R10
	MOVD  R22, -32(R10)
	SUB   $32, R10, R1

	MOVD R12, CTR
	CALL CTR

	// Deallocate frame
	MOVD R22, R1

	// Store return value
	MOVD R3, 80(R1)

	// Clear R0 before returning to Go code.
	// Go uses R0 as a constant 0 register for things like "std r0,X(r1)"
	// to zero stack locations. C functions may leave garbage in R0.
	XOR R0, R0, R0

	// Restore LR/CR from non-volatile registers
	MOVD R20, LR
	MOVW R21, CR
	RET
