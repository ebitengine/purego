// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2022 The Ebitengine Authors

package purego

import (
	"golang.org/x/sys/windows"
)

const (
	RTLD_DEFAULT = 0x00000 // Dlopen flags are not used on Windows.
)

// Dlopen examines the dynamic library or bundle file specified by path. If the file is compatible
// with the current process and has not already been loaded into the
// current process, it is loaded and linked. After being linked, if it contains
// any initializer functions, they are called, before Dlopen
// returns. It returns a handle that can be used with Dlsym and Dlclose.
// A second call to Dlopen with the same path will return the same handle, but the internal
// reference count for the handle will be incremented. Therefore, all
// Dlopen calls should be balanced with a Dlclose call.
func Dlopen(path string, _ int) (uintptr, error) {
	return openLibrary(path)
}

// Dlsym takes a "handle" of a dynamic library returned by Dlopen and the symbol name.
// It returns the address where that symbol is loaded into memory. If the symbol is not found,
// in the specified library or any of the libraries that were automatically loaded by Dlopen
// when that library was loaded, Dlsym returns zero.
func Dlsym(handle uintptr, name string) (uintptr, error) {
	return loadSymbol(handle, name)
}

// Dlclose decrements the reference count on the dynamic library handle.
// If the reference count drops to zero and no other loaded libraries
// use symbols in it, then the dynamic library is unloaded.
func Dlclose(handle uintptr) error {
	return windows.FreeLibrary(windows.Handle(handle))
}
