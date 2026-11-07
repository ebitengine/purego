// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 The Ebitengine Authors

package purego

import "unsafe"

//go:linkname __callbackasm1 callbackasm1
var __callbackasm1 byte

// CallbackasmAddr re-exports callbackasmAddr for external tests.
func CallbackasmAddr(i int) uintptr {
	return callbackasmAddr(i)
}

// Callbackasm1Addr returns the address that every callbackasm entry branches to.
func Callbackasm1Addr() uintptr {
	return uintptr(unsafe.Pointer(&__callbackasm1))
}
