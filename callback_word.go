// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 The Ebitengine Authors

//go:build darwin || freebsd || illumos || (linux && (386 || amd64 || arm || arm64 || loong64 || ppc64le || riscv64 || (s390x && (cgo || go1.27)))) || netbsd

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
		return func(a *callbackArgs) {
			a.result[0] = f()
		}
	case func(uintptr) uintptr:
		return func(a *callbackArgs) {
			a.result[0] = f(a.word(0))
		}
	case func(uintptr, uintptr) uintptr:
		return func(a *callbackArgs) {
			a.result[0] = f(a.word(0), a.word(1))
		}
	case func(uintptr, uintptr, uintptr) uintptr:
		return func(a *callbackArgs) {
			a.result[0] = f(a.word(0), a.word(1), a.word(2))
		}
	case func(uintptr, uintptr, uintptr, uintptr) uintptr:
		return func(a *callbackArgs) {
			a.result[0] = f(a.word(0), a.word(1), a.word(2), a.word(3))
		}
	case func(uintptr, uintptr, uintptr, uintptr, uintptr) uintptr:
		return func(a *callbackArgs) {
			a.result[0] = f(a.word(0), a.word(1), a.word(2), a.word(3), a.word(4))
		}
	case func(uintptr, uintptr, uintptr, uintptr, uintptr, uintptr) uintptr:
		return func(a *callbackArgs) {
			a.result[0] = f(a.word(0), a.word(1), a.word(2), a.word(3), a.word(4), a.word(5))
		}
	case func(uintptr, uintptr, uintptr, uintptr, uintptr, uintptr, uintptr) uintptr:
		return func(a *callbackArgs) {
			a.result[0] = f(a.word(0), a.word(1), a.word(2), a.word(3), a.word(4), a.word(5), a.word(6))
		}
	case func(uintptr, uintptr, uintptr, uintptr, uintptr, uintptr, uintptr, uintptr) uintptr:
		return func(a *callbackArgs) {
			a.result[0] = f(a.word(0), a.word(1), a.word(2), a.word(3), a.word(4), a.word(5), a.word(6), a.word(7))
		}
	case func(uintptr, uintptr, uintptr, uintptr, uintptr, uintptr, uintptr, uintptr, uintptr) uintptr:
		return func(a *callbackArgs) {
			a.result[0] = f(a.word(0), a.word(1), a.word(2), a.word(3), a.word(4), a.word(5), a.word(6), a.word(7), a.word(8))
		}
	default:
		return nil
	}
}

// word reads a pointer-sized integer argument from the saved C frame.
// Such arguments occupy one full slot even on big-endian targets and Apple's
// packed stack ABI. PPC64 and S390X keep stack arguments in a separate block.
func (a *callbackArgs) word(i int) uintptr {
	if i < numOfIntegerRegisters() {
		return *(*uintptr)(a.integerRegister(i))
	}
	if stack := a.stackFrame(); stack != nil {
		return *(*uintptr)(unsafe.Add(stack, uintptr(i-numOfIntegerRegisters())*ptrSize))
	}
	return *(*uintptr)(unsafe.Add(a.args, uintptr(numOfFloatRegisters()+i)*ptrSize))
}
