// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024 The Ebitengine Authors

//go:build darwin || freebsd || linux

package load

import "github.com/ebitengine/purego"

func OpenLibrary(name string) (uintptr, error) {
	return purego.Dlopen(name, purego.RTLD_NOW|purego.RTLD_GLOBAL)
}

func OpenSymbol(lib uintptr, name string) (uintptr, error) {
	return purego.Dlsym(lib, name)
}
