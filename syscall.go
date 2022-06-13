// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2022 The Ebiten Authors

//go:build darwin || windows
// +build darwin windows

package purego

const maxArgs = 9

// f matches argument numbers to a Syscall implementation that can take that many
var f = map[int]func(fn, a1, a2, a3, a4, a5, a6, a7, a8, a9 uintptr) (r1, r2, err uintptr){
	0: syscall_syscallX3,
	1: syscall_syscallX3,
	2: syscall_syscallX3,
	3: syscall_syscallX3,
	4: syscall_syscallX6,
	5: syscall_syscallX6,
	6: syscall_syscallX6,
	7: syscall_syscall9X,
	8: syscall_syscall9X,
	9: syscall_syscall9X,
}

// SyscallN takes fn, a C function pointer and a list of arguments as uintptr.
// There is an internal maximum number of arguments that SyscallN can take. It panics
// when the maximum is exceeded. It returns the result and the libc error code if there is one.
//
// NOTE: SyscallN does not properly call functions that have both integer and float parameters.
// See discussion comment https://github.com/ebiten/purego/pull/1#issuecomment-1128057607
// for an explanation of why that is.
//
// On amd64, if there are more than 8 floats the 9th and so on will be placed incorrectly on the
// stack.
//go:nosplit
func SyscallN(fn uintptr, args ...uintptr) (r1, r2, err uintptr) {
	if len(args) > maxArgs {
		panic("too many arguments to SyscallN")
	}
	// add padding so there is no out-of-bounds slicing
	var tmp [maxArgs]uintptr
	copy(tmp[:], args)
	return f[len(args)](fn, tmp[0], tmp[1], tmp[2], tmp[3], tmp[4], tmp[5], tmp[6], tmp[7], tmp[8])
}
