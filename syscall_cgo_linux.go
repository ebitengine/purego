// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2022 The Ebitengine Authors

//go:build cgo && !(386 || amd64 || arm || arm64 || loong64 || ppc64le || riscv64 || s390x)

package purego

import (
	"github.com/ebitengine/purego/internal/cgo"
)

// syscallXABI0 is the C fallback; unlike the assembly trampolines it always
// writes errno back into the a3 field of syscallArgs. Keep capturesErrno in
// syscall_errno_saved.go in sync with this build constraint.
var syscallXABI0 = uintptr(cgo.SyscallXABI0)

func NewCallback(_ any) uintptr {
	panic("purego: NewCallback on Linux is only supported on 386/amd64/arm64/arm/loong64/ppc64le/riscv64/s390x")
}

func syscall_syscallN(fn uintptr, args ...uintptr) (r1, r2, err uintptr) {
	panic("purego: syscall_syscallN is only supported on windows")
}
