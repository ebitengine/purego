// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024 The Ebitengine Authors

package purego

import "github.com/ebitengine/purego/internal/cgo"

// Source for constants: https://github.com/NetBSD/src/blob/trunk/include/dlfcn.h

const (
	intSize      = 32 << (^uint(0) >> 63) // 32 or 64
	RTLD_DEFAULT = 1<<intSize - 2         // Pseudo-handle for dlsym so search for any loaded symbol
	RTLD_LAZY    = 0x00000001             // Relocations are performed at an implementation-dependent time.
	RTLD_NOW     = 0x00000002             // Relocations are performed when the object is loaded.
	RTLD_LOCAL   = 0x00000000             // All symbols are not made available for relocation processing by other modules.
	RTLD_GLOBAL  = 0x00000100             // All symbols are available for relocation processing of other modules.
)

func Dlopen(path string, mode int) (uintptr, error) {
	return cgo.Dlopen(path, mode)
}

func Dlsym(handle uintptr, name string) (uintptr, error) {
	return cgo.Dlsym(handle, name)
}

func Dlclose(handle uintptr) error {
	return cgo.Dlclose(handle)
}

func loadSymbol(handle uintptr, name string) (uintptr, error) {
	return Dlsym(handle, name)
}
