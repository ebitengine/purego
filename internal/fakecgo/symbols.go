// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2022 The Ebitengine Authors

//go:build darwin
// +build darwin

package fakecgo

import "unsafe"

func setg_trampoline(setg uintptr, G uintptr)

//go:linkname memmove runtime.memmove
func memmove(to, from unsafe.Pointer, n uintptr)

func malloc(size uintptr) unsafe.Pointer

func free(ptr unsafe.Pointer)

func setenv(name *byte, value *byte, overwrite int32) int32

func unsetenv(name *byte) int32

// the no escape comments are needed inorder to ensure Go doesn't try
// and create a new object in the runtime when the runtime isn't available.
// After all, we are pretending to be C code.

//go:noescape
func pthread_attr_init(attr *pthread_attr_t) int32

//go:noescape
func pthread_create(thread *pthread_t, attr *pthread_attr_t, start, arg unsafe.Pointer) int32

func pthread_detach(thread pthread_t) int32

//go:noescape
func pthread_sigmask(how sighow, ign *sigset_t, oset *sigset_t) int32

//go:noescape
func pthread_attr_getstacksize(attr *pthread_attr_t, stacksize *size_t) int32

//go:noescape
func pthread_attr_destroy(attr *pthread_attr_t) int32

func abort()

//go:noescape
func sigfillset(set *sigset_t) int32

//go:noescape
func nanosleep(ts *timespec, rem *timespec) int32
