// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2022 The Ebitengine Authors

//go:build darwin
// +build darwin

// Package objc is a low-level pure Go objective-c runtime. This package is easy to use incorrectly, so it is best
// to use a wrapper that provides the functionality you need in a safer way.
package objc

import (
	"fmt"
	"reflect"
	"runtime"
	"unsafe"

	"github.com/ebitengine/purego"
	"github.com/ebitengine/purego/internal/strings"
)

//TODO: support try/catch?
//https://stackoverflow.com/questions/7062599/example-of-how-objective-cs-try-catch-implementation-is-executed-at-runtime

var (
	objc = purego.Dlopen("/usr/lib/libobjc.A.dylib", purego.RTLD_GLOBAL)

	objc_msgSend           = purego.Dlsym(objc, "objc_msgSend")
	objc_msgSendSuper2     = purego.Dlsym(objc, "objc_msgSendSuper2")
	objc_getClass          = purego.Dlsym(objc, "objc_getClass")
	objc_allocateClassPair = purego.Dlsym(objc, "objc_allocateClassPair")
	objc_registerClassPair = purego.Dlsym(objc, "objc_registerClassPair")
	sel_registerName       = purego.Dlsym(objc, "sel_registerName")
	class_getSuperclass    = purego.Dlsym(objc, "class_getSuperclass")
	class_addMethod        = purego.Dlsym(objc, "class_addMethod")
	object_getClass        = purego.Dlsym(objc, "object_getClass")
)

type Id uintptr

func (id Id) Class() Class {
	ret, _, _ := purego.SyscallN(object_getClass, uintptr(id))
	return Class(ret)
}

// Send is a convenience method for sending messages to objects.
func (id Id) Send(sel SEL, args ...interface{}) Id {
	var tmp = make([]uintptr, len(args)+2)
	createArgs(tmp, id, sel, args...)
	ret, _, _ := purego.SyscallN(objc_msgSend, tmp...)
	return Id(ret)
}

// SendSuper is a convenience method for sending message to object's super
func (id Id) SendSuper(sel SEL, args ...interface{}) Id {
	type objc_super struct {
		reciever   Id
		superClass Class
	}
	var _super = &objc_super{
		reciever:   id,
		superClass: id.Class(),
	}
	var tmp = make([]uintptr, len(args)+2)
	createArgs(tmp, Id(unsafe.Pointer(_super)), sel, args...)
	ret, _, _ := purego.SyscallN(objc_msgSendSuper2, tmp...)
	return Id(ret)
}

func createArgs(out []uintptr, cls Id, sel SEL, args ...interface{}) {
	out[0] = uintptr(cls)
	out[1] = uintptr(sel)
	for i, a := range args {
		switch v := a.(type) {
		case Id:
			out[i+2] = uintptr(v)
		case Class:
			out[i+2] = uintptr(v)
		case SEL:
			out[i+2] = uintptr(v)
		case _IMP:
			out[i+2] = uintptr(v)
		case uintptr:
			out[i+2] = v
		case int:
			out[i+2] = uintptr(v)
		case uint:
			out[i+2] = uintptr(v)
		default:
			panic(fmt.Sprintf("objc: unknown type %T", v))
		}
	}
}

type SEL uintptr

func RegisterName(name string) SEL {
	n := strings.CString(name, false)
	ret, _, _ := purego.SyscallN(sel_registerName, uintptr(unsafe.Pointer(n)))
	runtime.KeepAlive(n)
	return SEL(ret)
}

type Class uintptr

func GetClass(name string) Class {
	n := strings.CString(name, false)
	ret, _, _ := purego.SyscallN(objc_getClass, uintptr(unsafe.Pointer(n)))
	runtime.KeepAlive(n)
	return Class(ret)
}

func AllocateClassPair(super Class, name string, extraBytes uintptr) Class {
	n := strings.CString(name, false)
	ret, _, _ := purego.SyscallN(objc_allocateClassPair, uintptr(super), uintptr(unsafe.Pointer(n)), extraBytes)
	runtime.KeepAlive(n)
	return Class(ret)
}

func (c Class) SuperClass() Class {
	ret, _, _ := purego.SyscallN(class_getSuperclass, uintptr(c))
	return Class(ret)
}

func (c Class) AddMethod(name SEL, imp _IMP, types string) bool {
	t := strings.CString(types, false)
	ret, _, _ := purego.SyscallN(class_addMethod, uintptr(c), uintptr(name), uintptr(imp), uintptr(unsafe.Pointer(t)))
	runtime.KeepAlive(t)
	return byte(ret) != 0
}

func (c Class) Register() {
	purego.SyscallN(objc_registerClassPair, uintptr(c))
}

// _IMP is unexported so that the only way to make this type is by providing a Go function and casting
// it with the IMP function
type _IMP uintptr

// IMP takes a Go function that takes (id, SEL) as its first two arguments. It returns an _IMP function
// pointer that can be called by C code
func IMP(fn interface{}) _IMP {
	// this is only here so that it is easier to port C code to Go.
	// this is not guaranteed to be here forever so make sure to port your callbacks to Go
	// If you have a C function pointer cast it to a uintptr before passing it
	// to this function.
	if x, ok := fn.(uintptr); ok {
		return _IMP(x)
	}
	val := reflect.ValueOf(fn)
	if val.Kind() != reflect.Func {
		panic("not a function")
	}
	// IMP is stricter than a normal callback
	// id (*IMP)(id, SEL, ...)
	switch {
	case val.Type().NumIn() < 2:
		fallthrough
	case val.Type().In(0).Kind() != reflect.Uintptr:
		fallthrough
	case val.Type().In(1).Kind() != reflect.Uintptr:
		panic("IMP must take a (id, SEL) as its first two arguments")
	}
	return _IMP(purego.NewCallback(fn))
}
