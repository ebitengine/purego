// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 The Ebitengine Authors

//go:build !cgo

package fakecgo

import "structs"

// OpenBSD's pthread_mutex_t and pthread_cond_t are opaque POINTERS, and both
// PTHREAD_MUTEX_INITIALIZER and PTHREAD_COND_INITIALIZER are NULL --
// /usr/include/pthread.h:157-158 on 7.9. So uintptr(0) is the initialiser,
// which is the same shape NetBSD happens to have and not a copy of it.
type (
	pthread_cond_t  uintptr
	pthread_mutex_t uintptr
)

var (
	PTHREAD_COND_INITIALIZER  = pthread_cond_t(0)
	PTHREAD_MUTEX_INITIALIZER = pthread_mutex_t(0)
)

// Source: /usr/include/sys/signal.h on OpenBSD 7.9
//
//	typedef struct sigaltstack {
//	        void    *ss_sp;
//	        size_t   ss_size;
//	        int      ss_flags;
//	} stack_t;
type stack_t struct {
	_        structs.HostLayout
	ss_sp    uintptr
	ss_size  uintptr
	ss_flags int32
}

// Source: /usr/include/sys/signal.h -- SS_DISABLE 0x0004
const SS_DISABLE = 0x004
