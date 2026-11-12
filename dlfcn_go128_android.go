// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 The Ebitengine Authors

//go:build go1.28

package purego

import (
	"errors"

	"github.com/ebitengine/purego/internal/cbind"
)

// errNoBindings is returned when purego was built for Android with CGO_ENABLED=0
// on Go 1.28+. Bionic has no fallback trampolines in purego, so the dlfcn
// functions are only available through the cgo bindings in internal/cbind.
var errNoBindings = errors.New("purego: Android needs cgo bindings; build without CGO_ENABLED=0")

func Dlopen(path string, mode int) (uintptr, error) {
	if cbind.Dlopen == nil {
		return 0, errNoBindings
	}
	h := cbind.Dlopen(path, mode)
	if h == 0 {
		return 0, errors.New(cbind.Dlerror())
	}
	return h, nil
}

func Dlsym(handle uintptr, name string) (uintptr, error) {
	if cbind.Dlsym == nil {
		return 0, errNoBindings
	}
	p := cbind.Dlsym(handle, name)
	if p == 0 {
		return 0, errors.New(cbind.Dlerror())
	}
	return p, nil
}

func Dlclose(handle uintptr) error {
	if cbind.Dlclose == nil {
		return errNoBindings
	}
	if cbind.Dlclose(handle) {
		return errors.New(cbind.Dlerror())
	}
	return nil
}

func loadSymbol(handle uintptr, name string) (uintptr, error) {
	return Dlsym(handle, name)
}
