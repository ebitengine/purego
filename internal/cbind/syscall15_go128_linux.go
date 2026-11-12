// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 The Ebitengine Authors

//go:build go1.28 && !(386 || amd64 || arm || arm64 || loong64 || ppc64le || riscv64 || s390x)

package cbind

import "C"

import "runtime"

// This replaces syscall15 from internal/cgo (syscall_cgo_unix.go), which casts the
// target to a function taking 32 uintptr_t arguments and calls it. The same cast
// can be written as a //cgo:funcptr type, so cmd/cgo generates the indirect call
// in Go assembly and no C compiler is needed.
//
// purego_fn32 is not a real C typedef. The binding only names it, and nothing
// checks it unless someone runs `cgo -gen-binding` against a header declaring it.

//cgo:binding C.purego_fn32
type fn32 uintptr

//cgo:funcptr
func (fp fn32) Call(a1, a2, a3, a4, a5, a6, a7, a8, a9, a10, a11, a12, a13, a14, a15, a16,
	a17, a18, a19, a20, a21, a22, a23, a24, a25, a26, a27, a28, a29, a30, a31, a32 uintptr) uintptr

// __errno_location is the glibc and musl accessor for the calling thread's errno.
//
//cgo:binding C.__errno_location "_"
func errnoLocation() *int32

func init() {
	Syscall15 = syscall15
}

func syscall15(s *SyscallArgs) {
	// Like the C version, only integer arguments are supported here: there is no
	// way to say which of a C function's parameters are floats without knowing
	// its signature.
	for _, f := range s.F {
		if f != 0 {
			panic("purego: float arguments are not supported on " + runtime.GOARCH)
		}
	}
	// The C version read errno inside the same C call. Here the call and the errno
	// read are two separate cgo calls, so pin the goroutine to its thread so that
	// both run on the same one.
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	a := &s.A
	r1 := fn32(s.Fn).Call(a[0], a[1], a[2], a[3], a[4], a[5], a[6], a[7], a[8], a[9], a[10],
		a[11], a[12], a[13], a[14], a[15], a[16], a[17], a[18], a[19], a[20], a[21], a[22],
		a[23], a[24], a[25], a[26], a[27], a[28], a[29], a[30], a[31])
	errno := *errnoLocation()
	s.A[0] = r1
	s.A[2] = uintptr(errno)
}
