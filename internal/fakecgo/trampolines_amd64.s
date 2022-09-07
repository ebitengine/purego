// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2022 The Ebitengine Authors

//go:build darwin
// +build darwin

/*
trampoline for emulating required C functions for cgo in go (see cgo.go)
(we convert cdecl calling convention to go and vice-versa)

Since we're called from go and call into C we can cheat a bit with the calling conventions:
 - in go all the registers are caller saved
 - in C we have a couple of callee saved registers

=> we can use BX, R12, R13, R14, R15 instead of the stack

C Calling convention cdecl used here (we only need integer args):
1. arg: DI
2. arg: SI
3. arg: DX
4. arg: CX
5. arg: R8
6. arg: R9
We don't need floats with these functions -> AX=0
return value will be in AX
*/
#include "textflag.h"
#include "go_asm.h"

// these trampolines map the gcc ABI to Go ABI and then calls into the Go equivalent functions.

// TODO: put <> to make these private
TEXT x_cgo_init_trampoline(SB), NOSPLIT, $16
	MOVQ DI, 0(SP)
	MOVQ SI, 8(SP)
	CALL ·x_cgo_init(SB)
	RET

TEXT x_cgo_thread_start_trampoline(SB), NOSPLIT, $8
	MOVQ DI, 0(SP)
	CALL ·x_cgo_thread_start(SB)
	RET

TEXT x_cgo_setenv_trampoline(SB), NOSPLIT, $8
	MOVQ DI, 0(SP)
	CALL ·x_cgo_setenv(SB)
	RET

TEXT x_cgo_unsetenv_trampoline(SB), NOSPLIT, $8
	MOVQ DI, 0(SP)
	CALL ·x_cgo_unsetenv(SB)
	RET

TEXT x_cgo_notify_runtime_init_done_trampoline(SB), NOSPLIT, $0
	CALL ·x_cgo_notify_runtime_init_done(SB)
	RET

// func setg_trampoline(setg uintptr, g uintptr)
TEXT ·setg_trampoline(SB), NOSPLIT, $0-16
	MOVQ G+8(FP), DI
	MOVQ setg+0(FP), AX
	CALL AX
	RET

TEXT threadentry_trampoline(SB), NOSPLIT, $16
	MOVQ DI, 0(SP)
	CALL ·threadentry(SB)
	MOVQ 8(SP), AX
	RET

TEXT ·setenv(SB), NOSPLIT, $0-0
	MOVQ name+0(FP), DI
	MOVQ value+8(FP), SI
	MOVL overwrite+16(FP), DX
	CALL libc_setenv(SB)
	MOVL AX, ret+24(FP)
	RET

TEXT ·unsetenv(SB), NOSPLIT, $0-0
	MOVQ name+0(FP), DI
	CALL libc_unsetenv(SB)
	MOVL AX, ret+8(FP)
	RET

TEXT ·malloc(SB), NOSPLIT, $0-0
	MOVQ size+0(FP), DI
	CALL libc_malloc(SB)
	MOVQ AX, ret+8(FP)
	RET

TEXT ·free(SB), NOSPLIT, $0-0
	MOVQ ptr+0(FP), DI
	CALL libc_free(SB)
	RET

TEXT ·pthread_attr_init(SB), NOSPLIT, $0-12
	MOVQ attr+0(FP), DI
	CALL libc_pthread_attr_init(SB)
	MOVL AX, ret+8(FP)
	RET

TEXT ·pthread_detach(SB), NOSPLIT, $0-12
	MOVQ thread+0(FP), DI
	CALL libc_pthread_detach(SB)
	MOVL AX, ret+8(FP)
	RET

TEXT ·pthread_create(SB), NOSPLIT, $0-36
	MOVQ thread+0(FP), DI
	MOVQ attr+8(FP), SI
	MOVQ start+16(FP), DX
	MOVQ arg+24(FP), CX
	CALL libc_pthread_create(SB)
	MOVL AX, ret+32(FP)
	RET

TEXT ·pthread_attr_destroy(SB), NOSPLIT, $0-0
	MOVQ attr+0(FP), DI
	CALL libc_pthread_attr_destroy(SB)
	MOVL AX, ret+8(FP)
	RET

TEXT ·pthread_attr_getstacksize(SB), NOSPLIT, $0-0
	MOVQ attr+0(FP), DI
	MOVQ stacksize+8(FP), SI
	CALL libc_pthread_attr_getstacksize(SB)
	MOVL AX, ret+16(FP)
	RET

TEXT ·pthread_sigmask(SB), NOSPLIT, $0-0
	MOVL how+0(FP), DI
	MOVQ ign+8(FP), SI
	MOVQ oset+16(FP), DX
	CALL libc_pthread_sigmask(SB)
	MOVL AX, ret+24(FP)
	RET

TEXT ·abort(SB), NOSPLIT, $0-0
	CALL libc_abort(SB)
	RET

TEXT ·sigfillset(SB), NOSPLIT, $0-12
	MOVQ set+0(FP), DI
	CALL libc_sigfillset(SB)
	MOVL AX, ret+8(FP)
	RET

TEXT ·nanosleep(SB), NOSPLIT, $0-20
	MOVQ ts+0(FP), DI
	MOVQ rem+8(FP), SI
	CALL libc_nanosleep(SB)
	MOVL AX, ret+16(FP)
	RET
