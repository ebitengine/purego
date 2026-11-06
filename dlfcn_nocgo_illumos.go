// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 The Ebitengine Authors

//go:build !cgo

package purego

// The dlfcn functions are part of libc on illumos.

//go:cgo_import_dynamic purego_dlopen dlopen "libc.so"
//go:cgo_import_dynamic purego_dlsym dlsym "libc.so"
//go:cgo_import_dynamic purego_dlerror dlerror "libc.so"
//go:cgo_import_dynamic purego_dlclose dlclose "libc.so"
