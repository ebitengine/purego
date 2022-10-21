// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2022 The Ebiten Authors

package purego

//go:cgo_import_dynamic purego_dlopen dlopen "libdl.so.2"
//go:cgo_import_dynamic purego_dlsym dlsym "libdl.so.2"
//go:cgo_import_dynamic purego_dlerror dlerror "libdl.so.2"
//go:cgo_import_dynamic purego__dlclose dlclose "libdl.so.2"

// on amd64 we don't need the following line - on 386 we do...
// anyway - with those lines the output is better (but doesn't matter) - without it on amd64 we get multiple DT_NEEDED with "libc.so.6" etc

//go:cgo_import_dynamic _ _ "libdl.so.2"
