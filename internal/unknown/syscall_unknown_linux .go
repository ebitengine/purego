// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2022 The Ebitengine Authors

package unknown

// this file moved to package unknown since Cgo and assembly files can't be in the same package

/*
 #cgo LDFLAGS: -ldl

#include <stdint.h>
#include <dlfcn.h>
uintptr_t syscall9(uintptr_t fn, uintptr_t a1, uintptr_t a2, uintptr_t a3, uintptr_t a4, uintptr_t a5, uintptr_t a6, uintptr_t a7, uintptr_t a8, uintptr_t a9) {
	uintptr_t (*func_name)(uintptr_t a1, uintptr_t a2, uintptr_t a3, uintptr_t a4, uintptr_t a5, uintptr_t a6, uintptr_t a7, uintptr_t a8, uintptr_t a9);
	*(void**)(&func_name) = (void*)(fn);
	return func_name(a1,a2,a3,a4,a5,a6,a7,a8,a9);
}

*/
import "C"

//go:nosplit
func Syscall9X(fn, a1, a2, a3, a4, a5, a6, a7, a8, a9 uintptr) (r1, r2, err uintptr) {
	r1 = uintptr(C.syscall9(C.uintptr_t(fn), C.uintptr_t(a1), C.uintptr_t(a2), C.uintptr_t(a3), C.uintptr_t(a4), C.uintptr_t(a5), C.uintptr_t(a6), C.uintptr_t(a7), C.uintptr_t(a8), C.uintptr_t(a9)))
	return r1, 0, 0
}
