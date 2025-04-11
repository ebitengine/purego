// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024 The Ebitengine Authors

//go:build darwin || freebsd || linux || netbsd

package load

import "github.com/ebitengine/purego"

func OpenLibrary(name string) (uintptr, error) {
	return purego.Dlopen(name, purego.RTLD_NOW|purego.RTLD_GLOBAL)
}

func CloseLibrary(handle uintptr) error {
	return purego.Dlclose(handle)
}

func OpenSymbol(lib uintptr, name string) (uintptr, error) {
	return purego.Dlsym(lib, name)
}
