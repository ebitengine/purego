// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024 The Ebitengine Authors

//go:build freebsd || linux

package cgo

/*
 #cgo LDFLAGS: -ldl

// Some distributions have RTLD_DEFAULT hidden behind this flag
#ifndef _GNU_SOURCE
#define _GNU_SOURCE
#endif
#include <dlfcn.h>
#include <stdlib.h>
*/
import "C"

import (
	"errors"
	"unsafe"
)

var (
	RTLD_DEFAULT = int(uintptr(C.RTLD_DEFAULT))
	RTLD_LAZY    = int(uintptr(C.RTLD_LAZY))
	RTLD_NOW     = int(uintptr(C.RTLD_NOW))
	RTLD_LOCAL   = int(uintptr(C.RTLD_LOCAL))
	RTLD_GLOBAL  = int(uintptr(C.RTLD_GLOBAL))
)

func Dlopen(filename string, flag int) (uintptr, error) {
	cfilename := C.CString(filename)
	defer C.free(unsafe.Pointer(cfilename))
	handle := C.dlopen(cfilename, C.int(flag))
	if handle == nil {
		return 0, errors.New(C.GoString(C.dlerror()))
	}
	return uintptr(handle), nil
}

func Dlsym(handle uintptr, symbol string) (uintptr, error) {
	csymbol := C.CString(symbol)
	defer C.free(unsafe.Pointer(csymbol))
	symbolAddr := C.dlsym(*(*unsafe.Pointer)(unsafe.Pointer(&handle)), csymbol)
	if symbolAddr == nil {
		return 0, errors.New(C.GoString(C.dlerror()))
	}
	return uintptr(symbolAddr), nil
}

func Dlclose(handle uintptr) error {
	result := C.dlclose(*(*unsafe.Pointer)(unsafe.Pointer(&handle)))
	if result != 0 {
		return errors.New(C.GoString(C.dlerror()))
	}
	return nil
}

// all that is needed is to assign each dl function because then its
// symbol will then be made available to the linker and linked to inside dlfcn.go
var (
	_ = C.dlopen
	_ = C.dlsym
	_ = C.dlerror
	_ = C.dlclose
)
