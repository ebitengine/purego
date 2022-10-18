// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2022 The Ebitengine Authors

// Package objc is a low-level pure Go objective-c runtime. This package is easy to use incorrectly, so it is best
// to use a wrapper that provides the functionality you need in a safer way.
package objc

import (
	"github.com/ebitengine/purego"
	"math"
	"reflect"
)

//TODO: support try/catch?
//https://stackoverflow.com/questions/7062599/example-of-how-objective-cs-try-catch-implementation-is-executed-at-runtime

var (
	objc_msgSend              func(obj ID, cmd SEL, args ...interface{}) ID
	objc_msgSendSuper2        func(super *objc_super, cmd SEL, args ...interface{}) ID
	objc_getClass             func(name string) Class
	objc_getProtocol          func(name string) *Protocol
	objc_allocateClassPair    func(super Class, name string, extraBytes uintptr) Class
	objc_registerClassPair    func(class Class)
	sel_registerName          func(name string) SEL
	class_getSuperclass       func(class Class) Class
	class_getInstanceVariable func(class Class, name string) Ivar
	class_addMethod           func(class Class, name SEL, imp IMP, types string) bool
	class_addIvar             func(class Class, name string, size uintptr, alignment uint8, types string) bool
	class_addProtocol         func(class Class, protocol *Protocol) bool
	ivar_getOffset            func(ivar Ivar) uintptr
	object_getClass           func(obj ID) Class
)

func init() {
	objc := purego.Dlopen("/usr/lib/libobjc.A.dylib", purego.RTLD_GLOBAL)

	purego.Func(objc, "objc_msgSend", &objc_msgSend)
	purego.Func(objc, "objc_msgSendSuper2", &objc_msgSendSuper2)
	purego.Func(objc, "object_getClass", &object_getClass)
	purego.Func(objc, "objc_getClass", &objc_getClass)
	purego.Func(objc, "objc_getProtocol", &objc_getProtocol)
	purego.Func(objc, "objc_allocateClassPair", &objc_allocateClassPair)
	purego.Func(objc, "objc_registerClassPair", &objc_registerClassPair)
	purego.Func(objc, "sel_registerName", &sel_registerName)
	purego.Func(objc, "class_getSuperclass", &class_getSuperclass)
	purego.Func(objc, "class_getInstanceVariable", &class_getInstanceVariable)
	purego.Func(objc, "class_addMethod", &class_addMethod)
	purego.Func(objc, "class_addIvar", &class_addIvar)
	purego.Func(objc, "class_addProtocol", &class_addProtocol)
	purego.Func(objc, "ivar_getOffset", &ivar_getOffset)
}

// ID is an opaque pointer to some Objective-C object
type ID uintptr

// Class returns the class of the object.
func (id ID) Class() Class {
	return object_getClass(id)
}

// Send is a convenience method for sending messages to objects.
func (id ID) Send(sel SEL, args ...interface{}) ID {
	return objc_msgSend(id, sel, args...)
}

// objc_super data structure is generated by the Objective-C compiler when it encounters the super keyword
// as the receiver of a message. It specifies the class definition of the particular superclass that should
// be messaged.
type objc_super struct {
	receiver   ID
	superClass Class
}

// SendSuper is a convenience method for sending message to object's super
func (id ID) SendSuper(sel SEL, args ...interface{}) ID {
	var super = &objc_super{
		receiver:   id,
		superClass: id.Class(),
	}
	return objc_msgSendSuper2(super, sel, args...)
}

// SEL is an opaque type that represents a method selector
type SEL uintptr

// RegisterName registers a method with the Objective-C runtime system, maps the method name to a selector,
// and returns the selector value.
func RegisterName(name string) SEL {
	return sel_registerName(name)
}

// Class is an opaque type that represents an Objective-C class.
type Class uintptr

// GetClass returns the Class object for the named class, or nil if the class is not registered with the Objective-C runtime.
func GetClass(name string) Class {
	return objc_getClass(name)
}

// AllocateClassPair creates a new class and metaclass. Then returns the new class, or Nil if the class could not be created
func AllocateClassPair(super Class, name string, extraBytes uintptr) Class {
	return objc_allocateClassPair(super, name, extraBytes)
}

// SuperClass returns the superclass of a class.
// You should usually use NSObject‘s superclass method instead of this function.
func (c Class) SuperClass() Class {
	return class_getSuperclass(c)
}

// AddMethod adds a new method to a class with a given name and implementation.
// The types argument is a string containing the mapping of parameters and return type.
// Since the function must take at least two arguments—self and _cmd, the second and third
// characters must be “@:” (the first character is the return type).
func (c Class) AddMethod(name SEL, imp IMP, types string) bool {
	return class_addMethod(c, name, imp, types)
}

// AddIvar adds a new instance variable to a class.
// It may only be called after AllocateClassPair and before Register.
// Adding an instance variable to an existing class is not supported.
// The class must not be a metaclass. Adding an instance variable to a metaclass is not supported.
// It takes the instance of the type of the Ivar and a string representing the type.
func (c Class) AddIvar(name string, ty interface{}, types string) bool {
	typeOf := reflect.TypeOf(ty)
	size := typeOf.Size()
	alignment := uint8(math.Log2(float64(typeOf.Align())))
	return class_addIvar(c, name, size, alignment, types)
}

// AddProtocol adds a protocol to a class.
// Returns true if the protocol was added successfully, otherwise false (for example,
// the class already conforms to that protocol).
func (c Class) AddProtocol(protocol *Protocol) bool {
	return class_addProtocol(c, protocol)
}

// InstanceVariable returns an Ivar data structure containing information about the instance variable specified by name.
func (c Class) InstanceVariable(name string) Ivar {
	return class_getInstanceVariable(c, name)
}

// Register registers a class that was allocated using AllocateClassPair.
// It can now be used to make objects by sending it either alloc and init or new.
func (c Class) Register() {
	objc_registerClassPair(c)
}

// Ivar an opaque type that represents an instance variable.
type Ivar uintptr

// Offset returns the offset of an instance variable that can be used to assign and read the Ivar's value.
//
// For instance variables of type ID or other object types, call Ivar and SetIvar instead
// of using this offset to access the instance variable data directly.
func (i Ivar) Offset() uintptr {
	return ivar_getOffset(i)
}

// Protocol is a type that declares methods that can be implemented by any class.
type Protocol uintptr

// GetProtocol returns the protocol for the given name or nil if there is no protocol by that name.
func GetProtocol(name string) *Protocol {
	return objc_getProtocol(name)
}

// IMP is a function pointer that can be called by Objective-C code.
type IMP uintptr

// NewIMP takes a Go function that takes (ID, SEL) as its first two arguments. It returns an IMP function
// pointer that can be called by Objective-C code. The function pointer is never deallocated.
func NewIMP(fn interface{}) IMP {
	ty := reflect.TypeOf(fn)
	if ty.Kind() != reflect.Func {
		panic("objc: not a function")
	}
	// IMP is stricter than a normal callback
	// id (*IMP)(id, SEL, ...)
	switch {
	case ty.NumIn() < 2:
		fallthrough
	case ty.In(0).Kind() != reflect.Uintptr:
		fallthrough
	case ty.In(1).Kind() != reflect.Uintptr:
		panic("objc: NewIMP must take a (id, SEL) as its first two arguments")
	}
	return IMP(purego.NewCallback(fn))
}
