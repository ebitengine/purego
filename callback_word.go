// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 The Ebitengine Authors

//go:build darwin || freebsd || (linux && (386 || amd64 || arm || arm64 || loong64 || ppc64le || riscv64 || (s390x && (cgo || go1.27)))) || netbsd

package purego

import "unsafe"

// compileWordCallback avoids reflect.Call for common callbacks whose arguments
// and result are uintptr. Select the typed call once at registration, leaving
// all other signatures on the existing reflection path. Do not reinterpret
// function types: pointers require garbage collector metadata, and narrow
// integers can have different ABI layouts.
func compileWordCallback(fn any) func(*callbackArgs) {
	switch f := fn.(type) {
	case func() uintptr:
		return func(a *callbackArgs) { a.result[0] = f() }
	case func(uintptr) uintptr:
		return func(a *callbackArgs) {
			v := a.words(1)
			a.result[0] = f(v[0])
		}
	case func(uintptr, uintptr) uintptr:
		return func(a *callbackArgs) {
			v := a.words(2)
			a.result[0] = f(v[0], v[1])
		}
	case func(uintptr, uintptr, uintptr) uintptr:
		return func(a *callbackArgs) {
			v := a.words(3)
			a.result[0] = f(v[0], v[1], v[2])
		}
	case func(uintptr, uintptr, uintptr, uintptr) uintptr:
		return func(a *callbackArgs) {
			v := a.words(4)
			a.result[0] = f(v[0], v[1], v[2], v[3])
		}
	case func(uintptr, uintptr, uintptr, uintptr, uintptr) uintptr:
		return func(a *callbackArgs) {
			v := a.words(5)
			a.result[0] = f(v[0], v[1], v[2], v[3], v[4])
		}
	case func(uintptr, uintptr, uintptr, uintptr, uintptr, uintptr) uintptr:
		return func(a *callbackArgs) {
			v := a.words(6)
			a.result[0] = f(v[0], v[1], v[2], v[3], v[4], v[5])
		}
	case func(uintptr, uintptr, uintptr, uintptr, uintptr, uintptr, uintptr) uintptr:
		return func(a *callbackArgs) {
			v := a.words(7)
			a.result[0] = f(v[0], v[1], v[2], v[3], v[4], v[5], v[6])
		}
	case func(uintptr, uintptr, uintptr, uintptr, uintptr, uintptr, uintptr, uintptr) uintptr:
		return func(a *callbackArgs) {
			v := a.words(8)
			a.result[0] = f(v[0], v[1], v[2], v[3], v[4], v[5], v[6], v[7])
		}
	case func(uintptr, uintptr, uintptr, uintptr, uintptr, uintptr, uintptr, uintptr, uintptr) uintptr:
		return func(a *callbackArgs) {
			v := a.words(9)
			a.result[0] = f(v[0], v[1], v[2], v[3], v[4], v[5], v[6], v[7], v[8])
		}
	default:
		return nil
	}
}

// words copies only pointer-sized integer arguments from the saved C frame.
// Such arguments occupy one full slot even on big-endian targets and Apple's
// packed stack ABI. PPC64 and S390X keep stack arguments in a separate block.
// Keep scratch space on this invocation's stack for concurrent/nested callbacks.
func (a *callbackArgs) words(count int) (values [9]uintptr) {
	frame := (*[callbackMaxFrame]uintptr)(a.args)
	stack := a.stackFrame()
	for i := range count {
		if i >= numOfIntegerRegisters() && stack != nil {
			values[i] = *(*uintptr)(unsafe.Add(stack, uintptr(i-numOfIntegerRegisters())*ptrSize))
		} else {
			values[i] = frame[numOfFloatRegisters()+i]
		}
	}
	return values
}
