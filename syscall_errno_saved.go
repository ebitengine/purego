// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 The Ebitengine Authors

//go:build darwin || (linux && cgo && !(386 || amd64 || arm || arm64 || loong64 || ppc64le || riscv64 || s390x))

package purego

// capturesErrno reports whether the trampoline behind syscallXABI0 writes
// errno back into the a3 field of syscallArgs.
//
// The assembly trampolines do that only on darwin - see the GOOS_darwin guards
// in sys_*.s - while the C fallback in internal/cgo always does. That fallback
// is what syscall_cgo_linux.go selects, on the Linux architectures that have no
// assembly trampoline and so can only be built with Cgo: mips, mipsle, mips64,
// mips64le and ppc64 today. Everywhere else the trampoline leaves the caller's
// third argument untouched, so reading a3 back would return an input argument
// rather than an error code.
//
// Keep this build constraint in sync with syscall_cgo_linux.go.
const capturesErrno = true
