// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2022 The Ebiten Authors

package purego

// Source for constants: https://codebrowser.dev/glibc/glibc/bits/dlfcn.h.html

const (
	RTLD_LAZY   = 0x00001 // Relocations are performed at an implementation-dependent time.
	RTLD_NOW    = 0x00002 // Relocations are performed when the object is loaded.
	RTLD_LOCAL  = 0x00000 // All symbols are not made available for relocation processing by other modules.
	RTLD_GLOBAL = 0x00100 // All symbols are available for relocation processing of other modules.
)

//go:cgo_import_dynamic purego_dlopen dlopen "libdl.so.2"
//go:cgo_import_dynamic purego_dlsym dlsym "libdl.so.2"
//go:cgo_import_dynamic purego_dlerror dlerror "libdl.so.2"
//go:cgo_import_dynamic purego_dlclose dlclose "libdl.so.2"

// on amd64 we don't need the following line - on 386 we do...
// anyway - with those lines the output is better (but doesn't matter) - without it on amd64 we get multiple DT_NEEDED with "libc.so.6" etc

//go:cgo_import_dynamic _ _ "libdl.so.2"
