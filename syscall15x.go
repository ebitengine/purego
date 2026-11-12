// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 The Ebitengine Authors

//go:build !386 && !arm && !ppc64le && (darwin || freebsd || illumos || linux || netbsd || windows) && !(go1.28 && linux && !(amd64 || arm64 || loong64 || riscv64 || s390x))

package purego

import "unsafe"

// syscall15X runs the syscall15X trampoline on the system stack. On Go 1.28+ the
// Linux architectures without a purego trampoline use syscall15x_go128_linux.go
// instead.
func syscall15X(s *syscallArgs) {
	runtime_cgocall(syscallXABI0, unsafe.Pointer(s))
}
