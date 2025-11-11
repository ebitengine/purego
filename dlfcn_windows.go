// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2022 The Ebitengine Authors

package purego

import (
	"golang.org/x/sys/windows"
	"path/filepath"
)

const (
	RTLD_DEFAULT = 0x00000 // Pseudo-handle for dlsym so search for any loaded symbol
	RTLD_LAZY    = 0x00001 // Relocations are performed at an implementation-dependent time.
	RTLD_NOW     = 0x00002 // Relocations are performed when the object is loaded.
	RTLD_LOCAL   = 0x00000 // All symbols are not made available for relocation processing by other modules.
	RTLD_GLOBAL  = 0x00100 // All symbols are available for relocation processing of other modules.
)

// Dlopen examines the dynamic library or bundle file specified by path.
//
// The mode was ignored on Windows.
func Dlopen(path string, mode int) (uintptr, error) {
	if len(path) == 0 {
		return getSelfModuleHandle(true), nil
	}
	handle, err := windows.LoadLibrary(filepath.ToSlash(path))
	if err != nil {
		return 0, err
	}
	return uintptr(handle), nil
}

// Dlsym takes a "handle" of a dynamic library returned by Dlopen and the symbol name.
func Dlsym(handle uintptr, name string) (uintptr, error) {
	if handle == RTLD_DEFAULT {
		handle = getSelfModuleHandle(false)
	}
	return windows.GetProcAddress(windows.Handle(handle), name)
}

// Dlclose decrements the reference count on the dynamic library handle.
func Dlclose(handle uintptr) error {
	return windows.FreeLibrary(windows.Handle(handle))
}

func getSelfModuleHandle(incrementRefCount bool) uintptr {
	var handle windows.Handle
	if incrementRefCount {
		windows.GetModuleHandleEx(0, nil, &handle)
	} else {
		windows.GetModuleHandleEx(windows.GET_MODULE_HANDLE_EX_FLAG_UNCHANGED_REFCOUNT, nil, &handle)
	}
	return uintptr(handle)
}
