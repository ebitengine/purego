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

// Id is an opaque pointer to some Objective-C object
type Id uintptr

// Class returns the class of the object.
func (id Id) Class() Class {
	ret, _, _ := purego.SyscallN(object_getClass, uintptr(id))
	return Class(ret)
}

// Send is a convenience method for sending messages to objects.
func (id Id) Send(sel SEL, args ...interface{}) Id {
	tmp := createArgs(id, sel, args...)
	ret, _, _ := purego.SyscallN(objc_msgSend, tmp...)
	return Id(ret)
}

// sending a message to the super requires this struct instead of the object itself
type objc_super struct {
	receiver   Id
	superClass Class
}

// SendSuper is a convenience method for sending message to object's super
func (id Id) SendSuper(sel SEL, args ...interface{}) Id {
	var super = &objc_super{
		receiver:   id,
		superClass: id.Class(),
	}
	tmp := createArgs(0, sel, args...)
	tmp[0] = uintptr(unsafe.Pointer(super)) // if createArgs splits the stack the pointer would be wrong
	ret, _, _ := purego.SyscallN(objc_msgSendSuper2, tmp...)
	return Id(ret)
}

func createArgs(cls Id, sel SEL, args ...interface{}) (out []uintptr) {
	out = make([]uintptr, len(args)+2)
	out[0] = uintptr(cls)
	out[1] = uintptr(sel)
	out = out[:2]
	for _, a := range args {
		switch v := a.(type) {
		case Id:
			out = append(out, uintptr(v))
		case Class:
			out = append(out, uintptr(v))
		case SEL:
			out = append(out, uintptr(v))
		case _IMP:
			out = append(out, uintptr(v))
		case bool:
			if v {
				out = append(out, uintptr(1))
			} else {
				out = append(out, uintptr(0))
			}
		case unsafe.Pointer:
			out = append(out, uintptr(v))
		case uintptr:
			out = append(out, v)
		case int:
			out = append(out, uintptr(v))
		case uint:
			out = append(out, uintptr(v))
		default:
			panic(fmt.Sprintf("objc: unknown type %T", v))
		}
	}
	return out
}

// SEL is an opaque type that represents a method selector
type SEL uintptr

// RegisterName registers a method with the Objective-C runtime system, maps the method name to a selector,
// and returns the selector value.
func RegisterName(name string) SEL {
	n := strings.CString(name)
	ret, _, _ := purego.SyscallN(sel_registerName, uintptr(unsafe.Pointer(n)))
	runtime.KeepAlive(n)
	return SEL(ret)
}

// Class is an opaque type that represents an Objective-C class.
type Class uintptr

// GetClass returns the Class object for the named class, or nil if the class is not registered with the Objective-C runtime.
func GetClass(name string) Class {
	n := strings.CString(name)
	ret, _, _ := purego.SyscallN(objc_getClass, uintptr(unsafe.Pointer(n)))
	runtime.KeepAlive(n)
	return Class(ret)
}

// AllocateClassPair creates a new class and metaclass. Then returns the new class, or Nil if the class could not be created
func AllocateClassPair(super Class, name string, extraBytes uintptr) Class {
	n := strings.CString(name)
	ret, _, _ := purego.SyscallN(objc_allocateClassPair, uintptr(super), uintptr(unsafe.Pointer(n)), extraBytes)
	runtime.KeepAlive(n)
	return Class(ret)
}

// SuperClass returns the superclass of a class.
// You should usually use NSObject‘s superclass method instead of this function.
func (c Class) SuperClass() Class {
	ret, _, _ := purego.SyscallN(class_getSuperclass, uintptr(c))
	return Class(ret)
}

// AddMethod adds a new method to a class with a given name and implementation.
// The types argument is a string containing the mapping of parameters and return type.
// Since the function must take at least two arguments—self and _cmd, the second and third
// characters must be “@:” (the first character is the return type).
func (c Class) AddMethod(name SEL, imp _IMP, types string) bool {
	t := strings.CString(types)
	ret, _, _ := purego.SyscallN(class_addMethod, uintptr(c), uintptr(name), uintptr(imp), uintptr(unsafe.Pointer(t)))
	runtime.KeepAlive(t)
	return byte(ret) != 0
}

// Register registers a class that was allocated using objc_allocateClassPair.
// It can now be used to make objects by sending it either alloc and init or new.
func (c Class) Register() {
	purego.SyscallN(objc_registerClassPair, uintptr(c))
}

// _IMP is unexported so that the only way to make this type is by providing a Go function and casting
// it with the IMP function
type _IMP uintptr

// IMP takes a Go function that takes (id, SEL) as its first two arguments. It returns an _IMP function
// pointer that can be called by Objective-C code. The function pointer is never deallocated.
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
