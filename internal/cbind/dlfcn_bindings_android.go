// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 The Ebitengine Authors

//go:build go1.28

package cbind

import "C"

import "unsafe"

// Bionic keeps the dlfcn functions in libdl.so.

//cgo:binding C.dlopen "libdl.so"
func dlopen(path *byte, mode int32) unsafe.Pointer

//cgo:binding C.dlsym "libdl.so"
func dlsym(handle unsafe.Pointer, name *byte) unsafe.Pointer

//cgo:binding C.dlerror "libdl.so"
func dlerror() *byte

//cgo:binding C.dlclose "libdl.so"
func dlclose(handle unsafe.Pointer) int32
