// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2022 The Ebitengine Authors

//go:build darwin
// +build darwin

#include "textflag.h"
#include "go_asm.h"

// these trampolines map the gcc ABI to Go ABI and then calls into the Go equivalent functions.

// TODO: put <> to make these private
TEXT x_cgo_init_trampoline(SB), NOSPLIT, $0-0
	MOVD R0, 8(RSP)
	MOVD R1, 16(RSP)
	CALL ·x_cgo_init(SB)
	RET

TEXT x_cgo_thread_start_trampoline(SB), NOSPLIT, $0-0
	MOVD R0, 8(RSP)
	CALL ·x_cgo_thread_start(SB)
	RET

TEXT x_cgo_setenv_trampoline(SB), NOSPLIT, $0-0
	MOVD R0, 8(RSP)
	CALL ·x_cgo_setenv(SB)
	RET

TEXT x_cgo_unsetenv_trampoline(SB), NOSPLIT, $0-0
	MOVD R0, 8(RSP)
	CALL ·x_cgo_unsetenv(SB)
	RET

TEXT x_cgo_notify_runtime_init_done_trampoline(SB), NOSPLIT, $0-0
	CALL ·x_cgo_notify_runtime_init_done(SB)
	RET

// func setg_trampoline(setg uintptr, g uintptr)
TEXT ·setg_trampoline(SB), NOSPLIT, $0-16
	MOVD G+8(FP), R0
	MOVD setg+0(FP), R1
	CALL R1
	RET

TEXT threadentry_trampoline(SB), NOSPLIT, $0-0
	MOVD R0, 8(RSP)
	CALL ·threadentry(SB)
	MOVD $0, R0           // TODO: get the return value from threadentry
	RET

TEXT ·setenv(SB), NOSPLIT, $0-0
	MOVD name+0(FP), R0
	MOVD value+8(FP), R1
	MOVD overwrite+16(FP), R2
	CALL libc_setenv(SB)
	MOVD R0, ret+24(FP)
	RET

TEXT ·unsetenv(SB), NOSPLIT, $0-0
	MOVD name+0(FP), R0
	CALL libc_unsetenv(SB)
	MOVD R0, ret+8(FP)
	RET

TEXT ·malloc(SB), NOSPLIT, $0-0
	MOVD size+0(FP), R0
	CALL libc_malloc(SB)
	MOVD R0, ret+8(FP)
	RET

TEXT ·free(SB), NOSPLIT, $0-0
	MOVD ptr+0(FP), R0
	CALL libc_free(SB)
	RET

TEXT ·pthread_attr_init(SB), NOSPLIT, $0-12
	MOVD attr+0(FP), R0
	CALL libc_pthread_attr_init(SB)
	MOVD R0, ret+8(FP)
	RET

TEXT ·pthread_detach(SB), NOSPLIT, $0-12
	MOVD thread+0(FP), R0
	CALL libc_pthread_detach(SB)
	MOVD R0, ret+8(FP)
	RET

TEXT ·pthread_create(SB), NOSPLIT, $0-36
	MOVD thread+0(FP), R0
	MOVD attr+8(FP), R1
	MOVD start+16(FP), R2
	MOVD arg+24(FP), R3
	CALL libc_pthread_create(SB)
	MOVD R0, ret+32(FP)
	RET

TEXT ·pthread_attr_destroy(SB), NOSPLIT, $0-0
	MOVD attr+0(FP), R0
	CALL libc_pthread_attr_destroy(SB)
	MOVD R0, ret+8(FP)
	RET

TEXT ·pthread_attr_getstacksize(SB), NOSPLIT, $0-0
	MOVD attr+0(FP), R0
	MOVD stacksize+8(FP), R1
	CALL libc_pthread_attr_getstacksize(SB)
	MOVD R0, ret+16(FP)
	RET

TEXT ·pthread_sigmask(SB), NOSPLIT, $0-0
	MOVD how+0(FP), R0
	MOVD ign+8(FP), R1
	MOVD oset+16(FP), R2
	CALL libc_pthread_sigmask(SB)
	MOVD R0, ret+24(FP)
	RET

TEXT ·abort(SB), NOSPLIT, $0-0
	CALL libc_abort(SB)
	RET

TEXT ·sigfillset(SB), NOSPLIT, $0-12
	MOVD set+0(FP), R0
	CALL libc_sigfillset(SB)
	MOVD R0, ret+8(FP)
	RET

TEXT ·nanosleep(SB), NOSPLIT, $0-20
	MOVD ts+0(FP), R0
	MOVD rem+8(FP), R1
	CALL libc_nanosleep(SB)
	MOVD R0, ret+16(FP)
	RET
