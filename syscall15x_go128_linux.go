// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 The Ebitengine Authors

//go:build go1.28 && !(386 || amd64 || arm || arm64 || loong64 || ppc64le || riscv64 || s390x)

package purego

import (
	"unsafe"

	"github.com/ebitengine/purego/internal/cbind"
)

// On these architectures purego has no assembly trampoline. Before Go 1.28 it
// called through syscall15 in internal/cgo, which needs a C compiler. Now the call
// goes through a //cgo:funcptr binding (see internal/cbind), so it works in the
// default "cgo without a C toolchain" mode too. It still does not work with an
// explicit CGO_ENABLED=0, which leaves the bindings out.

func syscall15X(s *syscallArgs) {
	if cbind.Syscall15 == nil {
		panic("purego: calling C on this GOARCH needs cgo bindings; build without CGO_ENABLED=0")
	}
	cbind.Syscall15((*cbind.SyscallArgs)(unsafe.Pointer(s)))
}

func NewCallback(_ any) uintptr {
	panic("purego: NewCallback on Linux is only supported on 386/amd64/arm64/arm/loong64/ppc64le/riscv64/s390x")
}

func syscall_syscallN(fn uintptr, args ...uintptr) (r1, r2, err uintptr) {
	panic("purego: syscall_syscallN is only supported on windows")
}
