// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024 The Ebitengine Authors

package purego

import "github.com/ebitengine/purego/internal/cgo"

// Source for constants: https://android.googlesource.com/platform/bionic/+/refs/heads/main/libc/include/dlfcn.h

const (
	is64bits     = 1 << (^uintptr(0) >> 63) / 2
	is32bits     = 1 - is64bits
	RTLD_DEFAULT = is32bits * 0xffffffff
	RTLD_LAZY    = 0x00000001
	RTLD_NOW     = is64bits * 0x00000002
	RTLD_LOCAL   = 0x00000000
	RTLD_GLOBAL  = is64bits*0x00100 | is32bits*0x00000002
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
