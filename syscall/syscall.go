// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2022 The Ebiten Authors

//go:build darwin
// +build darwin

package syscall

import "unsafe"

const maxArgs = 9

func SyscallN(fn uintptr, args ...uintptr) (r1, r2, err uintptr) {
	if len(args) < maxArgs {
		// add padding so there is no out-of-bounds slicing
		var tmp = make([]uintptr, maxArgs)
		copy(tmp, args)
		args = tmp
	}
	switch len(args) {
	case 0, 1, 2, 3:
		return syscall_syscallX(fn, args[0], args[1], args[2])
	case 4, 5, 6:
		return syscall_syscall6X(fn, args[0], args[1], args[2], args[3], args[4], args[5])
	case 7, 8, 9:
		return syscall_syscall9X(fn, args[0], args[1], args[2], args[3], args[4], args[5], args[6], args[7], args[8])
	default:
		panic("too many arguments to SyscallN")
	}
}

func callc(fn uintptr, args unsafe.Pointer) {
	runtime_entersyscall()
	runtime_libcCall(unsafe.Pointer(fn), args)
	runtime_exitsyscall()
}

var syscallXABI0 uintptr

func syscallX() // implemented in assembly

func syscall_syscallX(fn, a1, a2, a3 uintptr) (r1, r2, err uintptr) {
	args := struct{ fn, a1, a2, a3, r1, r2, err uintptr }{
		fn, a1, a2, a3, r1, r2, err}
	callc(syscallXABI0, unsafe.Pointer(&args))
	return args.r1, args.r2, args.err
}

var syscall6XABI0 uintptr

func syscall6X() // implemented in assembly

func syscall_syscall6X(fn, a1, a2, a3, a4, a5, a6 uintptr) (r1, r2, err uintptr) {
	args := struct{ fn, a1, a2, a3, a4, a5, a6, r1, r2, err uintptr }{
		fn, a1, a2, a3, a4, a5, a6, r1, r2, err}
	callc(syscall6XABI0, unsafe.Pointer(&args))
	return args.r1, args.r2, args.err
}

var syscall9XABI0 uintptr

func syscall9X() // implemented in assembly

func syscall_syscall9X(fn, a1, a2, a3, a4, a5, a6, a7, a8, a9 uintptr) (r1, r2, err uintptr) {
	args := struct{ fn, a1, a2, a3, a4, a5, a6, a7, a8, a9, r1, r2, err uintptr }{
		fn, a1, a2, a3, a4, a5, a6, a7, a8, a9, r1, r2, err}
	callc(syscall9XABI0, unsafe.Pointer(&args))
	return args.r1, args.r2, args.err
}
