// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024 The Ebitengine Authors

package purego

import "github.com/ebitengine/purego/internal/cgo"

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
