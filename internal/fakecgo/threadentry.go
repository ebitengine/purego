// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build !cgo && (darwin || freebsd || illumos || linux || netbsd)

package fakecgo

import "unsafe"

// threadentry_trampolineABI0 maps the C ABI to Go ABI then calls the Go function
//
//go:linkname x_threadentry_trampoline threadentry_trampoline
var x_threadentry_trampoline byte
var threadentry_trampolineABI0 = &x_threadentry_trampoline

// setg_func is the setg_gcc function passed to x_cgo_init.
var setg_func uintptr

// threadentry follows runtime/cgo/gcc_unix.c. threadentry_platform is the
// per-OS hook, like x_cgo_threadentry_platform there.
//
//go:nosplit
//go:norace
func threadentry(v unsafe.Pointer) unsafe.Pointer {
	ts := *(*ThreadStart)(v)
	free(v)

	threadentry_platform()

	setg_trampoline(setg_func, uintptr(unsafe.Pointer(ts.g)))

	callThreadEntryFn(ts.fn)

	return nil
}
