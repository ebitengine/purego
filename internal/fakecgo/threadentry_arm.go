// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 The Ebitengine Authors

//go:build !cgo && (freebsd || linux)

package fakecgo

// callThreadEntryTrampoline calls threadentry_trampoline like a C caller that
// passes arg in R0 and keeps r4 in R4 across the call. It returns R4 as the
// trampoline leaves it.
// It is only used by tests. Implemented in trampolines_arm.s.
func callThreadEntryTrampoline(arg, r4 uintptr) uintptr
