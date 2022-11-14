// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2022 The Ebitengine Authors

//go:build cgo

package purego

import "github.com/ebitengine/purego/internal/unknown"

// this is only here to make the assembly files happy :)
type syscall9Args struct{ fn, a1, a2, a3, a4, a5, a6, a7, a8, a9, r1, r2, err uintptr }

//go:nosplit
func syscall_syscall9X(fn, a1, a2, a3, a4, a5, a6, a7, a8, a9 uintptr) (r1, r2, err uintptr) {
	return unknown.Syscall9X(fn, a1, a2, a3, a4, a5, a6, a7, a8, a9)
}

func NewCallback(_ interface{}) uintptr {
	panic("purego: NewCallback not supported")
}
