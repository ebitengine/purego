// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 The Ebitengine Authors

//go:build !cgo && linux

package fakecgo

// callThreadEntryTrampoline calls threadentry_trampoline like a C caller that
// passes arg in R3 and keeps r14 in R14 across the call. It returns R14 as the
// trampoline leaves it.
// It is only used by tests. Implemented in trampolines_ppc64le.s.
func callThreadEntryTrampoline(arg, r14 uintptr) uintptr
