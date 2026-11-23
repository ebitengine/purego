// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 The Ebitengine Authors

//go:build !cgo && !faketime

package purego

// Bionic has no versioned sonames, so the dlfcn functions are imported from "libdl.so" rather than "libdl.so.2".
// They are called through the stubs in dlfcn_stubs.s, as on Linux.

//go:cgo_import_dynamic purego_dlopen dlopen "libdl.so"
//go:cgo_import_dynamic purego_dlsym dlsym "libdl.so"
//go:cgo_import_dynamic purego_dlerror dlerror "libdl.so"
//go:cgo_import_dynamic purego_dlclose dlclose "libdl.so"
//go:cgo_import_dynamic _ _ "libdl.so"
