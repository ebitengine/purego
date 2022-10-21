// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2022 The Ebitengine Authors

package fakecgo

//go:cgo_import_dynamic purego_malloc malloc "libSystem.dylib"
//go:cgo_import_dynamic purego_free free "libSystem.dylib"
//go:cgo_import_dynamic purego_setenv setenv "libSystem.dylib"
//go:cgo_import_dynamic purego_unsetenv unsetenv "libSystem.dylib"
//go:cgo_import_dynamic purego_pthread_attr_init pthread_attr_init "libSystem.dylib"
//go:cgo_import_dynamic purego_pthread_create pthread_create "libSystem.dylib"
//go:cgo_import_dynamic purego_pthread_detach pthread_detach "libSystem.dylib"
//go:cgo_import_dynamic purego_pthread_attr_destroy pthread_attr_destroy "libSystem.dylib"
//go:cgo_import_dynamic purego_pthread_attr_getstacksize pthread_attr_getstacksize "libSystem.dylib"
//go:cgo_import_dynamic purego_pthread_sigmask pthread_sigmask "libSystem.dylib"
//go:cgo_import_dynamic purego_abort abort "libSystem.dylib"
//go:cgo_import_dynamic purego_sigfillset sigfillset "libSystem.dylib"
//go:cgo_import_dynamic purego_nanosleep nanosleep "libSystem.dylib"
