// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2023 The Ebitengine Authors

//go:build darwin || freebsd || (!cgo && linux && (amd64 || arm64))

package purego

import "unsafe"

var syscall9XABI0 uintptr

//go:nosplit
func syscall_syscall9X(fn, a1, a2, a3, a4, a5, a6, a7, a8, a9 uintptr) (r1, r2, err uintptr) {
	args := syscall9Args{
		fn, a1, a2, a3, a4, a5, a6, a7, a8, a9,
		a1, a2, a3, a4, a5, a6, a7, a8,
		r1, r2, err,
	}
	runtime_cgocall(syscall9XABI0, unsafe.Pointer(&args))
	return args.r1, args.r2, args.err
}

