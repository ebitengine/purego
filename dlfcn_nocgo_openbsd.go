// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 The Ebitengine Authors

//go:build !cgo

package purego

// OpenBSD keeps dlopen and friends in libc -- there is no libdl -- and ships
// no unversioned libc.so FILE: /usr/lib holds libc.so.103.0 and nothing else.
// "libc.so" is still the right name to write here: the Go linker emits it as a
// DT_NEEDED and OpenBSD's ld.so resolves an unversioned soname to the highest
// version present. The Go runtime itself does exactly this in
// runtime/sys_openbsd.go, which is the proof that it works.

//go:cgo_import_dynamic purego_dlopen dlopen "libc.so"
//go:cgo_import_dynamic purego_dlsym dlsym "libc.so"
//go:cgo_import_dynamic purego_dlerror dlerror "libc.so"
//go:cgo_import_dynamic purego_dlclose dlclose "libc.so"
