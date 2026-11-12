// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 The Ebitengine Authors

//go:build go1.28

package cbind

import "C"

import "unsafe"

// The dlfcn functions are part of libc on illumos.

//cgo:binding C.dlopen "libc.so"
func dlopen(path *byte, mode int32) unsafe.Pointer

//cgo:binding C.dlsym "libc.so"
func dlsym(handle unsafe.Pointer, name *byte) unsafe.Pointer

//cgo:binding C.dlerror "libc.so"
func dlerror() *byte

//cgo:binding C.dlclose "libc.so"
func dlclose(handle unsafe.Pointer) int32
