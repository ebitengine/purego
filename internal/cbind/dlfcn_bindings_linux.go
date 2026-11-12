// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 The Ebitengine Authors

//go:build go1.28 && !android

package cbind

import "C"

import "unsafe"

// libdl.so.2 still exists (as a stub) on glibc 2.34+, so the same DT_NEEDED works everywhere.

//cgo:binding C.dlopen "libdl.so.2"
func dlopen(path *byte, mode int32) unsafe.Pointer

//cgo:binding C.dlsym "libdl.so.2"
func dlsym(handle unsafe.Pointer, name *byte) unsafe.Pointer

//cgo:binding C.dlerror "libdl.so.2"
func dlerror() *byte

//cgo:binding C.dlclose "libdl.so.2"
func dlclose(handle unsafe.Pointer) int32
