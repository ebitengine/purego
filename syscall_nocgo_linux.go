// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2022 The Ebitengine Authors

//go:build !cgo && (amd64 || arm64)

package purego

import (
	"math"
	"unsafe"
)

var syscall9XABI0 uintptr

type syscall9Args struct {
	fn, a1, a2, a3, a4, a5, a6, a7, a8, a9 uintptr
	f1, f2, f3, f4, f5, f6, f7, f8         float64
	r1, r2, err                            uintptr
}

//go:nosplit
func syscall_syscall9X(fn, a1, a2, a3, a4, a5, a6, a7, a8, a9 uintptr) (r1, r2, err uintptr) {
	args := syscall9Args{fn, a1, a2, a3, a4, a5, a6, a7, a8, a9,
		math.Float64frombits(uint64(a1)), math.Float64frombits(uint64(a2)), math.Float64frombits(uint64(a3)),
		math.Float64frombits(uint64(a4)), math.Float64frombits(uint64(a5)), math.Float64frombits(uint64(a6)),
		math.Float64frombits(uint64(a7)), math.Float64frombits(uint64(a8)),
		r1, r2, err}
	runtime_cgocall(syscall9XABI0, unsafe.Pointer(&args))
	return args.r1, args.r2, args.err
}

func NewCallback(_ interface{}) uintptr {
	panic("purego: NewCallback not supported")
}
