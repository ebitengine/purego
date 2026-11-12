// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 The Ebitengine Authors

// Package cbind holds the parts of purego that Go 1.28's cgo bindings
// (golang/go#81450) can provide without a C compiler.
//
// The binding files in this package have `import "C"` and the go1.28 build
// constraint. The go command builds them in two cases: with CGO_ENABLED=1, and in
// the default "cgo without a C toolchain" mode. It leaves them out when CGO_ENABLED=0
// is set explicitly, and on older Go releases, which would try to compile the bodiless
// binding declarations as ordinary cgo and fail.
//
// There is no build tag that tells the rest of purego which case it is in: the `cgo`
// tag is false in bindings mode. So the binding files fill in the variables below
// from init, and callers check for nil and fall back to purego's own trampolines
// when they are unset.
package cbind

import "structs"

// The dlfcn functions. Set when the bindings are built; nil otherwise.
var (
	Dlopen  func(path string, mode int) uintptr
	Dlsym   func(handle uintptr, name string) uintptr
	Dlerror func() string
	// Dlclose reports true when dlclose failed, matching purego's fnDlclose.
	Dlclose func(handle uintptr) bool
)

// Syscall15 calls args.Fn with the integer arguments in args.A and stores the
// result in args.A[0] and errno in args.A[2]. It is set only on Linux
// architectures that purego has no assembly trampoline for, where it replaces the
// C-compiled syscall15 in internal/cgo.
var Syscall15 func(args *SyscallArgs)

// SyscallArgs has the same layout as purego's syscallArgs, so purego can convert a
// pointer to one into a pointer to the other.
type SyscallArgs struct {
	_ structs.HostLayout

	Fn uintptr
	A  [32]uintptr
	F  [8]uintptr
	R8 uintptr
}
