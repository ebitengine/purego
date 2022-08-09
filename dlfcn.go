// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2022 The Ebitengine Authors

//go:build darwin
// +build darwin

package purego

import (
	"reflect"
	"runtime"
	"unsafe"

	"github.com/ebitengine/purego/internal/strings"
)

const RTLD_GLOBAL = 0x8

const RTLD_DEFAULT = ^uintptr(1)

func Dlopen(path string, mode int) uintptr {
	bs := strings.CString(path)
	ret, _, _ := SyscallN(dlopenABI0, uintptr(unsafe.Pointer(bs)), uintptr(mode), 0)
	runtime.KeepAlive(bs)
	return ret
}

func Dlsym(handle uintptr, name string) uintptr {
	bs := strings.CString(name)
	ret, _, _ := SyscallN(dlsymABI0, handle, uintptr(unsafe.Pointer(bs)), 0)
	runtime.KeepAlive(bs)
	return ret
}

func Dlerror() string {
	// msg is only valid until the next call to Dlerror
	// which is why it gets copied into a Go string
	msg, _, _ := SyscallN(dlerrorABI0)
	if msg == 0 {
		return ""
	}
	var length int
	for {
		// use unsafe.Add once we reach 1.17
		if *(*byte)(unsafe.Pointer(msg + uintptr(length))) == '\x00' {
			break
		}
		length++
	}
	// use unsafe.Slice once we reach 1.17
	s := make([]byte, length)
	copy(s, *(*[]byte)(unsafe.Pointer(&reflect.SliceHeader{Data: msg, Len: length, Cap: length})))
	return string(s)
}

func Dlclose(handle uintptr) bool {
	ret, _, _ := SyscallN(dlcloseABI0, handle)
	return ret != 0
}

// these functions exist in dlfcn_stubs.s and are calling C functions linked to in dlfcn_GOOS.go
var dlopenABI0 uintptr
var dlsymABI0 uintptr
var dlcloseABI0 uintptr
var dlerrorABI0 uintptr
