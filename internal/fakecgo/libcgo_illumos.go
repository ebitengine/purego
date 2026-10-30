// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 The Ebitengine Authors

//go:build !cgo

package fakecgo

import (
	"structs"
	"unsafe"
)

// On illumos these are structs that libc writes into, so their size must
// match C exactly.
// Source: https://github.com/illumos/illumos-gate/blob/master/usr/src/uts/common/sys/types.h
type (
	pthread_mutex_t struct {
		_       structs.HostLayout
		flag1   uint16
		flag2   uint8
		ceiling uint8
		typ     uint16
		magic   uint16
		lock    uint64
		data    uint64
	}
	pthread_cond_t struct {
		_     structs.HostLayout
		flag  [4]uint8
		typ   uint16
		magic uint16
		data  uint64
	}
)

var (
	_ [24]byte = [unsafe.Sizeof(pthread_mutex_t{})]byte{}
	_ [16]byte = [unsafe.Sizeof(pthread_cond_t{})]byte{}
)

const (
	_MUTEX_MAGIC = 0x4d58 // "MX"
	_COND_MAGIC  = 0x4356 // "CV"
)

// Source: https://github.com/illumos/illumos-gate/blob/master/usr/src/head/pthread.h
var (
	PTHREAD_COND_INITIALIZER  = pthread_cond_t{magic: _COND_MAGIC}
	PTHREAD_MUTEX_INITIALIZER = pthread_mutex_t{magic: _MUTEX_MAGIC}
)

// Source: https://github.com/illumos/illumos-gate/blob/master/usr/src/uts/common/sys/iso/signal_iso.h
const (
	SIG_BLOCK   sighow = 1
	SIG_UNBLOCK sighow = 2
	SIG_SETMASK sighow = 3
)

type stack_t struct {
	/* not implemented */
}
