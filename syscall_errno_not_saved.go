// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 The Ebitengine Authors

//go:build !darwin && !(linux && cgo && !(386 || amd64 || arm || arm64 || loong64 || ppc64le || riscv64 || s390x))

package purego

// capturesErrno reports whether the trampoline behind syscallXABI0 writes
// errno back into the a3 field of syscallArgs.
//
// The assembly trampolines leave the caller's third argument untouched, so
// reading a3 back would return an input argument instead of an error code.
//
// Keep this build constraint in sync with syscall_errno_saved.go.
const capturesErrno = false
