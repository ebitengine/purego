// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2022 The Ebitengine Authors

package purego

import (
	"syscall"
)

//go:nosplit
func syscall_syscall9X(fn, a1, a2, a3, a4, a5, a6, a7, a8, a9 uintptr) (r1, r2, err uintptr) {
	r1, r2, errno := syscall.Syscall9(fn, 9, a1, a2, a3, a4, a5, a6, a7, a8, a9)
	return r1, r2, uintptr(errno)
}

func NewCallback(fn interface{}) uintptr {
	return syscall.NewCallback(fn)
}
