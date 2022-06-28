// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2022 The Ebiten Authors

//go:build darwin
// +build darwin

package purego

import (
	"unsafe"
)

//go:linkname runtime_libcCall runtime.cgocall
func runtime_libcCall(fn uintptr, arg unsafe.Pointer) int32 // from runtime/cgocall.go
