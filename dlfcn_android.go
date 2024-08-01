// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2022 The Ebitengine Authors

package purego

import "github.com/ebitengine/purego/internal/cgo"

var (
	RTLD_DEFAULT = cgo.RTLD_DEFAULT
	RTLD_LAZY    = cgo.RTLD_LAZY
	RTLD_NOW     = cgo.RTLD_NOW
	RTLD_LOCAL   = cgo.RTLD_LOCAL
	RTLD_GLOBAL  = cgo.RTLD_GLOBAL
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
