// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 The Ebitengine Authors

//go:build go1.28

package cbind

import "C"

import "unsafe"

// The dlfcn functions live in libSystem on darwin and ios.

//cgo:binding C.dlopen "/usr/lib/libSystem.B.dylib"
func dlopen(path *byte, mode int32) unsafe.Pointer

//cgo:binding C.dlsym "/usr/lib/libSystem.B.dylib"
func dlsym(handle unsafe.Pointer, name *byte) unsafe.Pointer

//cgo:binding C.dlerror "/usr/lib/libSystem.B.dylib"
func dlerror() *byte

//cgo:binding C.dlclose "/usr/lib/libSystem.B.dylib"
func dlclose(handle unsafe.Pointer) int32
