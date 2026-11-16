// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 The Ebitengine Authors

//go:build !cgo && amd64 && (darwin || freebsd || illumos || linux || netbsd)

package fakecgo

// callThreadEntryFn calls fn (runtime.mstart) while saving and restoring the
// frame pointer and other callee-saved registers around the call. mstart
// returns with BP clobbered, so without this shim the caller's frame-pointer
// epilogue would fault. Implemented in trampolines_amd64.s.
//
//go:nosplit
//go:norace
func callThreadEntryFn(fn uintptr)

// callThreadEntryTrampoline calls threadentry_trampoline like a C caller that
// passes arg in DI and keeps bx in BX across the call. It returns BX as the
// trampoline leaves it.
// It is only used by tests. Implemented in trampolines_amd64.s.
func callThreadEntryTrampoline(arg, bx uintptr) uintptr
