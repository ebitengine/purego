// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2022 The Ebitengine Authors

package purego

// Source for constants: https://opensource.apple.com/source/dyld/dyld-360.14/include/dlfcn.h.auto.html

/* The MODE argument to `dlopen' contains one of the following: */
const (
	RTLD_LAZY   = 0x1 // Relocations are performed at an implementation-dependent time.
	RTLD_NOW    = 0x2 // Relocations are performed when the object is loaded.
	RTLD_LOCAL  = 0x4 // All symbols are not made available for relocation processing by other modules.
	RTLD_GLOBAL = 0x8 // All symbols are available for relocation processing of other modules.
)

/*
 * Special handle arguments for dlsym().
 */
const (
	RTLD_NEXT      = uintptr(0x8000000000000001) /* Search subsequent objects. */
	RTLD_DEFAULT   = uintptr(0x8000000000000002) /* Use default search algorithm. */
	RTLD_SELF      = uintptr(0x8000000000000003) /* Search this and subsequent objects (Mac OS X 10.5 and later) */
	RTLD_MAIN_ONLY = uintptr(0x8000000000000005) /* Search main executable only (Mac OS X 10.5 and later) */
)

//go:cgo_import_dynamic purego_dlopen dlopen "/usr/lib/libSystem.B.dylib"
//go:cgo_import_dynamic purego_dlsym dlsym "/usr/lib/libSystem.B.dylib"
//go:cgo_import_dynamic purego_dlerror dlerror "/usr/lib/libSystem.B.dylib"
//go:cgo_import_dynamic purego_dlclose dlclose "/usr/lib/libSystem.B.dylib"
