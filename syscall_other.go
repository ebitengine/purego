// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 The Ebitengine Authors

//go:build linux && !amd64 && !arm64 && !loong64

package purego

var syscall15XABI0 uintptr

func syscall_syscall15X(fn, a1, a2, a3, a4, a5, a6, a7, a8, a9, a10, a11, a12, a13, a14, a15 uintptr) (r1, r2, err uintptr) {
	panic("purego: syscall_syscall15X not supported on this platform")
}

func NewCallback(_ any) uintptr {
	panic("purego: NewCallback not supported on this platform")
}
