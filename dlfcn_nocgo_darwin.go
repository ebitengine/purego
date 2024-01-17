// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024 The Ebitengine Authors

//go:build !cgo

package purego

//go:cgo_import_dynamic purego_dlopen dlopen "/usr/lib/libSystem.B.dylib"
//go:cgo_import_dynamic purego_dlsym dlsym "/usr/lib/libSystem.B.dylib"
//go:cgo_import_dynamic purego_dlerror dlerror "/usr/lib/libSystem.B.dylib"
//go:cgo_import_dynamic purego_dlclose dlclose "/usr/lib/libSystem.B.dylib"
