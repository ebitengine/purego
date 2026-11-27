// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 The Ebitengine Authors

//go:build darwin || freebsd || linux || netbsd || windows

package purego

import "runtime"

// Syscall0 is a fixed-arity variant of [SyscallN] for zero arguments.
// It avoids the heap allocation of a variadic slice when called across module
// boundaries. See [SyscallN] for the full documentation and caveats.
//
//go:uintptrescapes
func Syscall0(fn uintptr) (r1, r2, err uintptr) {
	if fn == 0 {
		panic("purego: fn is nil")
	}
	if runtime.GOOS == "windows" {
		return syscall_syscallN(fn)
	}
	var tmp [maxArgs]uintptr
	var floats [maxArgs]uintptr
	s := syscall_SyscallN(fn, tmp[:], floats[:], 0)
	defer thePool.Put(s)
	return s.a1, s.a2, s.a3
}

// Syscall1 is a fixed-arity variant of [SyscallN] for one argument.
// It avoids the heap allocation of a variadic slice when called across module
// boundaries. See [SyscallN] for the full documentation and caveats.
//
//go:uintptrescapes
func Syscall1(fn, a1 uintptr) (r1, r2, err uintptr) {
	if fn == 0 {
		panic("purego: fn is nil")
	}
	if runtime.GOOS == "windows" {
		return syscall_syscallN(fn, a1)
	}
	var tmp [maxArgs]uintptr
	tmp[0] = a1
	var floats [maxArgs]uintptr
	floats[0] = a1
	s := syscall_SyscallN(fn, tmp[:], floats[:], 0)
	defer thePool.Put(s)
	return s.a1, s.a2, s.a3
}

// Syscall2 is a fixed-arity variant of [SyscallN] for two arguments.
// It avoids the heap allocation of a variadic slice when called across module
// boundaries. See [SyscallN] for the full documentation and caveats.
//
//go:uintptrescapes
func Syscall2(fn, a1, a2 uintptr) (r1, r2, err uintptr) {
	if fn == 0 {
		panic("purego: fn is nil")
	}
	if runtime.GOOS == "windows" {
		return syscall_syscallN(fn, a1, a2)
	}
	var tmp [maxArgs]uintptr
	tmp[0], tmp[1] = a1, a2
	var floats [maxArgs]uintptr
	floats[0], floats[1] = a1, a2
	s := syscall_SyscallN(fn, tmp[:], floats[:], 0)
	defer thePool.Put(s)
	return s.a1, s.a2, s.a3
}

// Syscall3 is a fixed-arity variant of [SyscallN] for three arguments.
// It avoids the heap allocation of a variadic slice when called across module
// boundaries. See [SyscallN] for the full documentation and caveats.
//
//go:uintptrescapes
func Syscall3(fn, a1, a2, a3 uintptr) (r1, r2, err uintptr) {
	if fn == 0 {
		panic("purego: fn is nil")
	}
	if runtime.GOOS == "windows" {
		return syscall_syscallN(fn, a1, a2, a3)
	}
	var tmp [maxArgs]uintptr
	tmp[0], tmp[1], tmp[2] = a1, a2, a3
	var floats [maxArgs]uintptr
	floats[0], floats[1], floats[2] = a1, a2, a3
	s := syscall_SyscallN(fn, tmp[:], floats[:], 0)
	defer thePool.Put(s)
	return s.a1, s.a2, s.a3
}

// Syscall4 is a fixed-arity variant of [SyscallN] for four arguments.
// It avoids the heap allocation of a variadic slice when called across module
// boundaries. See [SyscallN] for the full documentation and caveats.
//
//go:uintptrescapes
func Syscall4(fn, a1, a2, a3, a4 uintptr) (r1, r2, err uintptr) {
	if fn == 0 {
		panic("purego: fn is nil")
	}
	if runtime.GOOS == "windows" {
		return syscall_syscallN(fn, a1, a2, a3, a4)
	}
	var tmp [maxArgs]uintptr
	tmp[0], tmp[1], tmp[2], tmp[3] = a1, a2, a3, a4
	var floats [maxArgs]uintptr
	floats[0], floats[1], floats[2], floats[3] = a1, a2, a3, a4
	s := syscall_SyscallN(fn, tmp[:], floats[:], 0)
	defer thePool.Put(s)
	return s.a1, s.a2, s.a3
}

// Syscall5 is a fixed-arity variant of [SyscallN] for five arguments.
// It avoids the heap allocation of a variadic slice when called across module
// boundaries. See [SyscallN] for the full documentation and caveats.
//
//go:uintptrescapes
func Syscall5(fn, a1, a2, a3, a4, a5 uintptr) (r1, r2, err uintptr) {
	if fn == 0 {
		panic("purego: fn is nil")
	}
	if runtime.GOOS == "windows" {
		return syscall_syscallN(fn, a1, a2, a3, a4, a5)
	}
	var tmp [maxArgs]uintptr
	tmp[0], tmp[1], tmp[2], tmp[3], tmp[4] = a1, a2, a3, a4, a5
	var floats [maxArgs]uintptr
	floats[0], floats[1], floats[2], floats[3], floats[4] = a1, a2, a3, a4, a5
	s := syscall_SyscallN(fn, tmp[:], floats[:], 0)
	defer thePool.Put(s)
	return s.a1, s.a2, s.a3
}

// Syscall6 is a fixed-arity variant of [SyscallN] for six arguments.
// It avoids the heap allocation of a variadic slice when called across module
// boundaries. See [SyscallN] for the full documentation and caveats.
//
//go:uintptrescapes
func Syscall6(fn, a1, a2, a3, a4, a5, a6 uintptr) (r1, r2, err uintptr) {
	if fn == 0 {
		panic("purego: fn is nil")
	}
	if runtime.GOOS == "windows" {
		return syscall_syscallN(fn, a1, a2, a3, a4, a5, a6)
	}
	var tmp [maxArgs]uintptr
	tmp[0], tmp[1], tmp[2], tmp[3], tmp[4], tmp[5] = a1, a2, a3, a4, a5, a6
	var floats [maxArgs]uintptr
	floats[0], floats[1], floats[2], floats[3], floats[4], floats[5] = a1, a2, a3, a4, a5, a6
	s := syscall_SyscallN(fn, tmp[:], floats[:], 0)
	defer thePool.Put(s)
	return s.a1, s.a2, s.a3
}

// Syscall7 is a fixed-arity variant of [SyscallN] for seven arguments.
// It avoids the heap allocation of a variadic slice when called across module
// boundaries. See [SyscallN] for the full documentation and caveats.
//
//go:uintptrescapes
func Syscall7(fn, a1, a2, a3, a4, a5, a6, a7 uintptr) (r1, r2, err uintptr) {
	if fn == 0 {
		panic("purego: fn is nil")
	}
	if runtime.GOOS == "windows" {
		return syscall_syscallN(fn, a1, a2, a3, a4, a5, a6, a7)
	}
	var tmp [maxArgs]uintptr
	tmp[0], tmp[1], tmp[2], tmp[3], tmp[4], tmp[5], tmp[6] = a1, a2, a3, a4, a5, a6, a7
	var floats [maxArgs]uintptr
	floats[0], floats[1], floats[2], floats[3], floats[4], floats[5], floats[6] = a1, a2, a3, a4, a5, a6, a7
	s := syscall_SyscallN(fn, tmp[:], floats[:], 0)
	defer thePool.Put(s)
	return s.a1, s.a2, s.a3
}

// Syscall8 is a fixed-arity variant of [SyscallN] for eight arguments.
// It avoids the heap allocation of a variadic slice when called across module
// boundaries. See [SyscallN] for the full documentation and caveats.
//
//go:uintptrescapes
func Syscall8(fn, a1, a2, a3, a4, a5, a6, a7, a8 uintptr) (r1, r2, err uintptr) {
	if fn == 0 {
		panic("purego: fn is nil")
	}
	if runtime.GOOS == "windows" {
		return syscall_syscallN(fn, a1, a2, a3, a4, a5, a6, a7, a8)
	}
	var tmp [maxArgs]uintptr
	tmp[0], tmp[1], tmp[2], tmp[3], tmp[4], tmp[5], tmp[6], tmp[7] = a1, a2, a3, a4, a5, a6, a7, a8
	var floats [maxArgs]uintptr
	floats[0], floats[1], floats[2], floats[3], floats[4], floats[5], floats[6], floats[7] = a1, a2, a3, a4, a5, a6, a7, a8
	s := syscall_SyscallN(fn, tmp[:], floats[:], 0)
	defer thePool.Put(s)
	return s.a1, s.a2, s.a3
}

// Syscall9 is a fixed-arity variant of [SyscallN] for nine arguments.
// It avoids the heap allocation of a variadic slice when called across module
// boundaries. See [SyscallN] for the full documentation and caveats.
//
//go:uintptrescapes
func Syscall9(fn, a1, a2, a3, a4, a5, a6, a7, a8, a9 uintptr) (r1, r2, err uintptr) {
	if fn == 0 {
		panic("purego: fn is nil")
	}
	if runtime.GOOS == "windows" {
		return syscall_syscallN(fn, a1, a2, a3, a4, a5, a6, a7, a8, a9)
	}
	var tmp [maxArgs]uintptr
	tmp[0], tmp[1], tmp[2], tmp[3], tmp[4], tmp[5], tmp[6], tmp[7], tmp[8] = a1, a2, a3, a4, a5, a6, a7, a8, a9
	var floats [maxArgs]uintptr
	floats[0], floats[1], floats[2], floats[3], floats[4], floats[5], floats[6], floats[7], floats[8] = a1, a2, a3, a4, a5, a6, a7, a8, a9
	s := syscall_SyscallN(fn, tmp[:], floats[:], 0)
	defer thePool.Put(s)
	return s.a1, s.a2, s.a3
}

// Syscall10 is a fixed-arity variant of [SyscallN] for ten arguments.
// It avoids the heap allocation of a variadic slice when called across module
// boundaries. See [SyscallN] for the full documentation and caveats.
//
//go:uintptrescapes
func Syscall10(fn, a1, a2, a3, a4, a5, a6, a7, a8, a9, a10 uintptr) (r1, r2, err uintptr) {
	if fn == 0 {
		panic("purego: fn is nil")
	}
	if runtime.GOOS == "windows" {
		return syscall_syscallN(fn, a1, a2, a3, a4, a5, a6, a7, a8, a9, a10)
	}
	var tmp [maxArgs]uintptr
	tmp[0], tmp[1], tmp[2], tmp[3], tmp[4], tmp[5], tmp[6], tmp[7], tmp[8], tmp[9] = a1, a2, a3, a4, a5, a6, a7, a8, a9, a10
	var floats [maxArgs]uintptr
	floats[0], floats[1], floats[2], floats[3], floats[4], floats[5], floats[6], floats[7], floats[8], floats[9] = a1, a2, a3, a4, a5, a6, a7, a8, a9, a10
	s := syscall_SyscallN(fn, tmp[:], floats[:], 0)
	defer thePool.Put(s)
	return s.a1, s.a2, s.a3
}

// Syscall11 is a fixed-arity variant of [SyscallN] for eleven arguments.
// It avoids the heap allocation of a variadic slice when called across module
// boundaries. See [SyscallN] for the full documentation and caveats.
//
//go:uintptrescapes
func Syscall11(fn, a1, a2, a3, a4, a5, a6, a7, a8, a9, a10, a11 uintptr) (r1, r2, err uintptr) {
	if fn == 0 {
		panic("purego: fn is nil")
	}
	if runtime.GOOS == "windows" {
		return syscall_syscallN(fn, a1, a2, a3, a4, a5, a6, a7, a8, a9, a10, a11)
	}
	var tmp [maxArgs]uintptr
	tmp[0], tmp[1], tmp[2], tmp[3], tmp[4], tmp[5], tmp[6], tmp[7], tmp[8], tmp[9], tmp[10] = a1, a2, a3, a4, a5, a6, a7, a8, a9, a10, a11
	var floats [maxArgs]uintptr
	floats[0], floats[1], floats[2], floats[3], floats[4], floats[5], floats[6], floats[7], floats[8], floats[9], floats[10] = a1, a2, a3, a4, a5, a6, a7, a8, a9, a10, a11
	s := syscall_SyscallN(fn, tmp[:], floats[:], 0)
	defer thePool.Put(s)
	return s.a1, s.a2, s.a3
}

// Syscall12 is a fixed-arity variant of [SyscallN] for twelve arguments.
// It avoids the heap allocation of a variadic slice when called across module
// boundaries. See [SyscallN] for the full documentation and caveats.
//
//go:uintptrescapes
func Syscall12(fn, a1, a2, a3, a4, a5, a6, a7, a8, a9, a10, a11, a12 uintptr) (r1, r2, err uintptr) {
	if fn == 0 {
		panic("purego: fn is nil")
	}
	if runtime.GOOS == "windows" {
		return syscall_syscallN(fn, a1, a2, a3, a4, a5, a6, a7, a8, a9, a10, a11, a12)
	}
	var tmp [maxArgs]uintptr
	tmp[0], tmp[1], tmp[2], tmp[3], tmp[4], tmp[5], tmp[6], tmp[7], tmp[8], tmp[9], tmp[10], tmp[11] = a1, a2, a3, a4, a5, a6, a7, a8, a9, a10, a11, a12
	var floats [maxArgs]uintptr
	floats[0], floats[1], floats[2], floats[3], floats[4], floats[5], floats[6], floats[7], floats[8], floats[9], floats[10], floats[11] = a1, a2, a3, a4, a5, a6, a7, a8, a9, a10, a11, a12
	s := syscall_SyscallN(fn, tmp[:], floats[:], 0)
	defer thePool.Put(s)
	return s.a1, s.a2, s.a3
}

// Syscall13 is a fixed-arity variant of [SyscallN] for thirteen arguments.
// It avoids the heap allocation of a variadic slice when called across module
// boundaries. See [SyscallN] for the full documentation and caveats.
//
//go:uintptrescapes
func Syscall13(fn, a1, a2, a3, a4, a5, a6, a7, a8, a9, a10, a11, a12, a13 uintptr) (r1, r2, err uintptr) {
	if fn == 0 {
		panic("purego: fn is nil")
	}
	if runtime.GOOS == "windows" {
		return syscall_syscallN(fn, a1, a2, a3, a4, a5, a6, a7, a8, a9, a10, a11, a12, a13)
	}
	var tmp [maxArgs]uintptr
	tmp[0], tmp[1], tmp[2], tmp[3], tmp[4], tmp[5], tmp[6], tmp[7], tmp[8], tmp[9], tmp[10], tmp[11], tmp[12] = a1, a2, a3, a4, a5, a6, a7, a8, a9, a10, a11, a12, a13
	var floats [maxArgs]uintptr
	floats[0], floats[1], floats[2], floats[3], floats[4], floats[5], floats[6], floats[7], floats[8], floats[9], floats[10], floats[11], floats[12] = a1, a2, a3, a4, a5, a6, a7, a8, a9, a10, a11, a12, a13
	s := syscall_SyscallN(fn, tmp[:], floats[:], 0)
	defer thePool.Put(s)
	return s.a1, s.a2, s.a3
}

// Syscall14 is a fixed-arity variant of [SyscallN] for fourteen arguments.
// It avoids the heap allocation of a variadic slice when called across module
// boundaries. See [SyscallN] for the full documentation and caveats.
//
//go:uintptrescapes
func Syscall14(fn, a1, a2, a3, a4, a5, a6, a7, a8, a9, a10, a11, a12, a13, a14 uintptr) (r1, r2, err uintptr) {
	if fn == 0 {
		panic("purego: fn is nil")
	}
	if runtime.GOOS == "windows" {
		return syscall_syscallN(fn, a1, a2, a3, a4, a5, a6, a7, a8, a9, a10, a11, a12, a13, a14)
	}
	var tmp [maxArgs]uintptr
	tmp[0], tmp[1], tmp[2], tmp[3], tmp[4], tmp[5], tmp[6], tmp[7], tmp[8], tmp[9], tmp[10], tmp[11], tmp[12], tmp[13] = a1, a2, a3, a4, a5, a6, a7, a8, a9, a10, a11, a12, a13, a14
	var floats [maxArgs]uintptr
	floats[0], floats[1], floats[2], floats[3], floats[4], floats[5], floats[6], floats[7], floats[8], floats[9], floats[10], floats[11], floats[12], floats[13] = a1, a2, a3, a4, a5, a6, a7, a8, a9, a10, a11, a12, a13, a14
	s := syscall_SyscallN(fn, tmp[:], floats[:], 0)
	defer thePool.Put(s)
	return s.a1, s.a2, s.a3
}

// Syscall15 is a fixed-arity variant of [SyscallN] for fifteen arguments.
// It avoids the heap allocation of a variadic slice when called across module
// boundaries. See [SyscallN] for the full documentation and caveats.
//
//go:uintptrescapes
func Syscall15(fn, a1, a2, a3, a4, a5, a6, a7, a8, a9, a10, a11, a12, a13, a14, a15 uintptr) (r1, r2, err uintptr) {
	if fn == 0 {
		panic("purego: fn is nil")
	}
	if runtime.GOOS == "windows" {
		return syscall_syscallN(fn, a1, a2, a3, a4, a5, a6, a7, a8, a9, a10, a11, a12, a13, a14, a15)
	}
	var tmp [maxArgs]uintptr
	tmp[0], tmp[1], tmp[2], tmp[3], tmp[4], tmp[5], tmp[6], tmp[7], tmp[8], tmp[9], tmp[10], tmp[11], tmp[12], tmp[13], tmp[14] = a1, a2, a3, a4, a5, a6, a7, a8, a9, a10, a11, a12, a13, a14, a15
	var floats [maxArgs]uintptr
	floats[0], floats[1], floats[2], floats[3], floats[4], floats[5], floats[6], floats[7], floats[8], floats[9], floats[10], floats[11], floats[12], floats[13], floats[14] = a1, a2, a3, a4, a5, a6, a7, a8, a9, a10, a11, a12, a13, a14, a15
	s := syscall_SyscallN(fn, tmp[:], floats[:], 0)
	defer thePool.Put(s)
	return s.a1, s.a2, s.a3
}
