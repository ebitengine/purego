// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2022 The Ebitengine Authors

//go:build darwin
// +build darwin

package fakecgo

import (
	"syscall"
	"unsafe"
)

//go:nosplit
func x_cgo_notify_runtime_init_done() {
	// we don't support being called as a library
}

// _cgo_try_pthread_create retries pthread_create if it fails with
// EAGAIN.
//
//go:nosplit
func _cgo_try_pthread_create(thread *pthread_t, attr *pthread_attr_t, pfn unsafe.Pointer, arg *ThreadStart) int {
	var tries int
	var err int
	var ts timespec

	for tries = 0; tries < 20; tries++ {
		err = int(pthread_create(thread, attr, pfn, unsafe.Pointer(arg)))
		if err == 0 {
			pthread_detach(*thread)
			return 0
		}
		if err != int(syscall.EAGAIN) {
			return err
		}
		ts.tv_sec = 0
		ts.tv_nsec = (tries + 1) * 1000 * 1000 // Milliseconds.
		nanosleep(&ts, nil)
	}
	return int(syscall.EAGAIN)
}
