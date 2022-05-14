// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2022 The Ebiten Authors

//go:build darwin
// +build darwin

package syscall

import "unsafe"

func SyscallX(fn, a1, a2, a3 uintptr) (r1, r2, err uintptr) {
	return syscall_syscallX(fn, a1, a2, a3)
}

func Syscall6X(fn, a1, a2, a3, a4, a5, a6 uintptr) (r1, r2, err uintptr) {
	return syscall_syscall6X(fn, a1, a2, a3, a4, a5, a6)
}

var syscall9XABI0 uintptr

func syscall9X() // implemented in assembly

func Syscall9X(fn, a1, a2, a3, a4, a5, a6, a7, a8, a9 uintptr) (r1, r2, err uintptr) {
	args := struct{ fn, a1, a2, a3, a4, a5, a6, a7, a8, a9, r1, r2, err uintptr }{
		fn, a1, a2, a3, a4, a5, a6, a7, a8, a9, r1, r2, err}
	runtime_entersyscall()
	runtime_libcCall(unsafe.Pointer(syscall9XABI0), unsafe.Pointer(&args))
	runtime_exitsyscall()
	return args.r1, args.r2, args.err
}

var syscallXFABI0 uintptr

func syscallXF() // implemented in assembly

func SyscallXF(fn, a1, a2, a3 uintptr, f1, f2, f3 float64) (r1, r2, err uintptr) {
	args := struct {
		fn, a1, a2, a3 uintptr
		f1, f2, f3     float64
		r1, r2, err    uintptr
	}{fn, a1, a2, a3, f1, f2, f3, r1, r2, err}
	runtime_entersyscall()
	runtime_libcCall(unsafe.Pointer(syscallXFABI0), unsafe.Pointer(&args))
	runtime_exitsyscall()
	return args.r1, args.r2, args.err
}
