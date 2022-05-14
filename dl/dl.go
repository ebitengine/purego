// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2022 The Ebiten Authors

//go:build darwin
// +build darwin

package dl

import (
	"github.com/ebiten/purego/syscall"
	"unsafe"
)

const RTLD_DEFAULT = ^uintptr(1)

func Sym(handle uintptr, name *byte) uintptr {
	ret, _, _ := syscall.SyscallX(dlsymABI0, handle, uintptr(unsafe.Pointer(name)), 0)
	return ret
}
