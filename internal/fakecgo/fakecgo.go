// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2022 The Ebitengine Authors

//go:build !cgo && (darwin || freebsd || linux || netbsd)

package fakecgo

import "unsafe"

// setg_trampoline calls setg with the G provided
func setg_trampoline(setg uintptr, G uintptr)

// call5 takes fn the C function and 5 arguments and calls the function with those arguments
func call5(fn, a1, a2, a3, a4, a5 uintptr) uintptr

// argset matches runtime/cgocall.go:argset.
type argset struct {
	args   *uintptr
	retval uintptr
}

//go:nosplit
//go:norace
func (a *argset) arg(i int) unsafe.Pointer {
	// this indirection is to avoid go vet complaining about possible misuse of unsafe.Pointer
	return *(*unsafe.Pointer)(unsafe.Add(unsafe.Pointer(a.args), uintptr(i)*unsafe.Sizeof(uintptr(0))))
}
