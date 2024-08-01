// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024 The Ebitengine Authors

//go:build freebsd || linux

package cgo

/*
 #cgo LDFLAGS: -ldl

#include <dlfcn.h>
*/
import "C"

import (
	"errors"
	"unsafe"
)

var (
	RTLD_DEFAULT = C.RTLD_DEFAULT
	RTLD_LAZY    = C.RTLD_LAZY
	RTLD_NOW     = C.RTLD_NOW
	RTLD_LOCAL   = C.RTLD_LOCAL
	RTLD_GLOBAL  = C.RTLD_GLOBAL
)

func Dlopen(filename string, flag int) (uintptr, error) {
	cfilename := C.CString(filename)
	defer C.free(unsafe.Pointer(cfilename))
	handle := C.dlopen(cfilename, C.int(flag))
	if handle == nil {
		return 0, errors.New(C.GoString(C.dlerror()))
	}
	return handle, nil
}

func Dlsym(handle uintptr, symbol string) (uintptr, error) {
	csymbol := C.CString(symbol)
	defer C.free(unsafe.Pointer(csymbol))
	symbolAddr := C.dlsym(handle, csymbol)
	if symbolAddr == nil {
		return 0, errors.New(C.GoString(C.dlerror()))
	}
	return symbolAddr, nil
}

func Dlclose(handle uintptr) error {
	result := C.dlclose(handle)
	if result != 0 {
		return errors.New(C.GoString(C.dlerror()))
	}
	return nil
}
