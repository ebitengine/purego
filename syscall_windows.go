// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2022 The Ebitengine Authors

package purego

import (
	"reflect"
	"syscall"
	"unsafe"
)

var syscallXABI0 uintptr

func syscall_syscallN(fn uintptr, args ...uintptr) (r1, r2, err uintptr) {
	r1, r2, errno := syscall.SyscallN(fn, args...)
	return r1, r2, uintptr(errno)
}

// NewCallback converts a Go function to a function pointer conforming to the stdcall calling convention.
// This is useful when interoperating with Windows code requiring callbacks. On 386 it supports scalar
// integer, pointer, bool, float32, and float64 arguments and results, including 64-bit values occupying two
// stack slots. On other Windows architectures, the standard library's callback restrictions apply.
// Only a limited number of callbacks may be created in a single Go process, and any memory allocated for
// these callbacks is never released. Although this function is similar to the Unix version it may act
// differently.
func NewCallback(fn any) uintptr {
	isCDecl := false
	ty := reflect.TypeOf(fn)
	if ty == nil || ty.Kind() != reflect.Func {
		panic("purego: the type must be a function")
	}
	for i := 0; i < ty.NumIn(); i++ {
		in := ty.In(i)
		if !in.AssignableTo(reflect.TypeOf(CDecl{})) {
			continue
		}
		if i != 0 {
			panic("purego: CDecl must be the first argument")
		}
		isCDecl = true
	}
	return newCallback(fn, isCDecl)
}

func loadSymbol(handle uintptr, name string) (uintptr, error) {
	return syscall.GetProcAddress(syscall.Handle(handle), name)
}

// callbackMaxFrame is a stub definition.
const callbackMaxFrame = 0

func callbackArgFromStack(argsBase unsafe.Pointer, stackSlot int, stackByteOffset *uintptr, inType reflect.Type) reflect.Value {
	panic("purego: callbackArgFromStack should not be called on windows")
}
