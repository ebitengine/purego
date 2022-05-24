// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2022 The Ebiten Authors

//go:build darwin
// +build darwin

package purego

import "unsafe"

const maxArgs = 9

// SyscallN takes fn, a C function pointer and a list of arguments as uintptr.
// There is an internal maximum number of arguments that SyscallN can take. It panics
// when the maximum is exceeded. It returns the result and the libc error code if there is one.
//
// NOTE: SyscallN does not properly call functions that have both integer and float parameters.
// See discussion comment https://github.com/ebiten/purego/pull/1#issuecomment-1128057607
// for an explanation of why that is.
func SyscallN(fn uintptr, args ...uintptr) (r1, r2, err uintptr) {
	if len(args) > maxArgs {
		panic("too many arguments to SyscallN")
	}
	if len(args) <= 3 {
		var tmp = make([]uintptr, maxArgs)
		copy(tmp, args)
		syscall_syscallX(fn, tmp[0], tmp[1], tmp[2])
	}
	if len(args) < maxArgs {
		// add padding so there is no out-of-bounds slicing
		var tmp = make([]uintptr, maxArgs)
		copy(tmp, args)
		args = tmp
	}
	return syscall_syscall9X(fn, args[0], args[1], args[2], args[3], args[4], args[5], args[6], args[7], args[8])
}

func callc(fn uintptr, args unsafe.Pointer) {
	runtime_entersyscall()
	runtime_libcCall(unsafe.Pointer(fn), args)
	runtime_exitsyscall()
}

var syscall9XABI0 uintptr

func syscall9X() // implemented in assembly

func syscall_syscall9X(fn, a1, a2, a3, a4, a5, a6, a7, a8, a9 uintptr) (r1, r2, err uintptr) {
	args := struct{ fn, a1, a2, a3, a4, a5, a6, a7, a8, a9, r1, r2, err uintptr }{
		fn, a1, a2, a3, a4, a5, a6, a7, a8, a9, r1, r2, err}
	callc(syscall9XABI0, unsafe.Pointer(&args))
	return args.r1, args.r2, args.err
}
