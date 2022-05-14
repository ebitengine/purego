// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2022 The Ebiten Authors

package dl

import (
	"github.com/ebiten/purego/syscall"
	"runtime"
	"unsafe"
)

const RTLD_DEFAULT = ^uintptr(1)

func Open(name string, mode int) uintptr {
	bs := []byte(name)
	ret, _, _ := syscall.SyscallN(dlopenABI0, uintptr(unsafe.Pointer(&bs[0])), uintptr(mode), 0)
	runtime.KeepAlive(bs)
	return ret
}

func Sym(handle uintptr, name string) uintptr {
	bs := []byte(name)
	ret, _, _ := syscall.SyscallN(dlsymABI0, handle, uintptr(unsafe.Pointer(&bs[0])), 0)
	runtime.KeepAlive(bs)
	return ret
}

//go:cgo_import_dynamic _dlopen dlopen "/usr/lib/libSystem.B.dylib"
//go:cgo_import_dynamic _dlerror dlerror "/usr/lib/libSystem.B.dylib"
//go:cgo_import_dynamic _dlclose dlclose "/usr/lib/libSystem.B.dylib"
//go:cgo_import_dynamic _dlsym dlsym "/usr/lib/libSystem.B.dylib"

var dlopenABI0 uintptr

func dlopen() // implemented in assembly

var dlsymABI0 uintptr

func dlsym() // implemented in assembly
