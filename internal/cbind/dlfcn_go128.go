// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 The Ebitengine Authors

//go:build go1.28 && (darwin || freebsd || illumos || linux || netbsd)

package cbind

// This file has no bindings of its own; importing "C" is what keeps it out of
// builds where the binding files are left out.
import "C"

import (
	"unsafe"

	"github.com/ebitengine/purego/internal/strings"
)

func init() {
	Dlopen = func(path string, mode int) uintptr {
		return uintptr(dlopen(strings.CString(path), int32(mode)))
	}
	Dlsym = func(handle uintptr, name string) uintptr {
		return uintptr(dlsym(handlePointer(handle), strings.CString(name)))
	}
	Dlerror = func() string {
		return strings.GoString(uintptr(unsafe.Pointer(dlerror())))
	}
	Dlclose = func(handle uintptr) bool {
		return dlclose(handlePointer(handle)) != 0
	}
}

// handlePointer turns a dlopen handle back into the void* the C functions take.
// Handles are C pointers (or pseudo-handles like RTLD_DEFAULT), never Go
// pointers, so the cgo pointer checks have nothing to say about them.
func handlePointer(h uintptr) unsafe.Pointer {
	return *(*unsafe.Pointer)(unsafe.Pointer(&h))
}
