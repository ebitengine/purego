// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2022 The Ebitengine Authors

package unknown

// this file moved to package unknown since Cgo and assembly files can't be in the same package

/*
 #cgo LDFLAGS: -ldl

#include <stdint.h>
#include <dlfcn.h>
#include <errno.h>
#include <assert.h>

typedef struct syscall9Args {
	uintptr_t fn;
	uintptr_t a1, a2, a3, a4, a5, a6, a7, a8, a9;
	uintptr_t f1, f2, f3, f4, f5, f6, f7, f8;
	uintptr_t r1, r2, err;
} syscall9Args;

uintptr_t syscall9(struct syscall9Args *args) {
	assert((args->f1|args->f2|args->f3|args->f4|args->f5|args->f6|args->f7|args->f8) == 0);
	uintptr_t (*func_name)(uintptr_t a1, uintptr_t a2, uintptr_t a3, uintptr_t a4, uintptr_t a5, uintptr_t a6, uintptr_t a7, uintptr_t a8, uintptr_t a9);
	*(void**)(&func_name) = (void*)(args->fn);
	uintptr_t r1 =  func_name(args->a1,args->a2,args->a3,args->a4,args->a5,args->a6,args->a7,args->a8,args->a9);
	args->err = errno;
	return r1;
}

*/
import "C"
import "unsafe"

//go:linkname internal_syscall9XABI0 purego.syscall9XABI0
var internal_syscall9XABI0 = unsafe.Pointer(C.syscall9)

//go:nosplit
func Syscall9X(fn, a1, a2, a3, a4, a5, a6, a7, a8, a9 uintptr) (r1, r2, err uintptr) {
	args := C.syscall9Args{C.uintptr_t(fn), C.uintptr_t(a1), C.uintptr_t(a2), C.uintptr_t(a3),
		C.uintptr_t(a4), C.uintptr_t(a5), C.uintptr_t(a6),
		C.uintptr_t(a7), C.uintptr_t(a8), C.uintptr_t(a9), 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	r1 = uintptr(C.syscall9(&args))
	return r1, 0, uintptr(args.err)
}
