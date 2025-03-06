// Code generated by 'go generate' with gen.go. DO NOT EDIT.

// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2022 The Ebitengine Authors

//go:build !cgo && (darwin || freebsd || linux || netbsd)

#include "textflag.h"

// these stubs are here because it is not possible to go:linkname directly the C functions on darwin arm64

TEXT _malloc(SB), NOSPLIT|NOFRAME, $0-0
	JMP purego_malloc(SB)
	RET

TEXT _free(SB), NOSPLIT|NOFRAME, $0-0
	JMP purego_free(SB)
	RET

TEXT _setenv(SB), NOSPLIT|NOFRAME, $0-0
	JMP purego_setenv(SB)
	RET

TEXT _unsetenv(SB), NOSPLIT|NOFRAME, $0-0
	JMP purego_unsetenv(SB)
	RET

TEXT _sigfillset(SB), NOSPLIT|NOFRAME, $0-0
	JMP purego_sigfillset(SB)
	RET

TEXT _nanosleep(SB), NOSPLIT|NOFRAME, $0-0
	JMP purego_nanosleep(SB)
	RET

TEXT _abort(SB), NOSPLIT|NOFRAME, $0-0
	JMP purego_abort(SB)
	RET

TEXT _pthread_attr_init(SB), NOSPLIT|NOFRAME, $0-0
	JMP purego_pthread_attr_init(SB)
	RET

TEXT _pthread_create(SB), NOSPLIT|NOFRAME, $0-0
	JMP purego_pthread_create(SB)
	RET

TEXT _pthread_detach(SB), NOSPLIT|NOFRAME, $0-0
	JMP purego_pthread_detach(SB)
	RET

TEXT _pthread_sigmask(SB), NOSPLIT|NOFRAME, $0-0
	JMP purego_pthread_sigmask(SB)
	RET

TEXT _pthread_self(SB), NOSPLIT|NOFRAME, $0-0
	JMP purego_pthread_self(SB)
	RET

TEXT _pthread_get_stacksize_np(SB), NOSPLIT|NOFRAME, $0-0
	JMP purego_pthread_get_stacksize_np(SB)
	RET

TEXT _pthread_attr_getstacksize(SB), NOSPLIT|NOFRAME, $0-0
	JMP purego_pthread_attr_getstacksize(SB)
	RET

TEXT _pthread_attr_setstacksize(SB), NOSPLIT|NOFRAME, $0-0
	JMP purego_pthread_attr_setstacksize(SB)
	RET

TEXT _pthread_attr_destroy(SB), NOSPLIT|NOFRAME, $0-0
	JMP purego_pthread_attr_destroy(SB)
	RET

TEXT _pthread_mutex_lock(SB), NOSPLIT|NOFRAME, $0-0
	JMP purego_pthread_mutex_lock(SB)
	RET

TEXT _pthread_mutex_unlock(SB), NOSPLIT|NOFRAME, $0-0
	JMP purego_pthread_mutex_unlock(SB)
	RET

TEXT _pthread_cond_broadcast(SB), NOSPLIT|NOFRAME, $0-0
	JMP purego_pthread_cond_broadcast(SB)
	RET

TEXT _pthread_setspecific(SB), NOSPLIT|NOFRAME, $0-0
	JMP purego_pthread_setspecific(SB)
	RET
