// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2022 The Ebitengine Authors

// Package objc is a low-level pure Go objective-c runtime. This package is easy to use incorrectly, so it is best
// to use a wrapper that provides the functionality you need in a safer way.
package objc

import (
	"errors"
	"fmt"
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
	class_getInstanceSize     func(class Class) uintptr
	class_addMethod           func(class Class, name SEL, imp IMP, types string) bool
	class_addIvar             func(class Class, name string, size uintptr, alignment uint8, types string) bool
	class_addProtocol         func(class Class, protocol *Protocol) bool
	ivar_getOffset            func(ivar Ivar) uintptr
	object_getClass           func(obj ID) Class
)

func init() {
	objc := purego.Dlopen("/usr/lib/libobjc.A.dylib", purego.RTLD_GLOBAL)

	purego.RegisterLibFunc(&objc_msgSend, objc, "objc_msgSend")
	purego.RegisterLibFunc(&objc_msgSendSuper2, objc, "objc_msgSendSuper2")
	purego.RegisterLibFunc(&object_getClass, objc, "object_getClass")
	purego.RegisterLibFunc(&objc_getClass, objc, "objc_getClass")
	purego.RegisterLibFunc(&objc_getProtocol, objc, "objc_getProtocol")
	purego.RegisterLibFunc(&objc_allocateClassPair, objc, "objc_allocateClassPair")
	purego.RegisterLibFunc(&objc_registerClassPair, objc, "objc_registerClassPair")
	purego.RegisterLibFunc(&sel_registerName, objc, "sel_registerName")
	purego.RegisterLibFunc(&class_getSuperclass, objc, "class_getSuperclass")
	purego.RegisterLibFunc(&class_getInstanceVariable, objc, "class_getInstanceVariable")
	purego.RegisterLibFunc(&class_addMethod, objc, "class_addMethod")
	purego.RegisterLibFunc(&class_addIvar, objc, "class_addIvar")
	purego.RegisterLibFunc(&class_addProtocol, objc, "class_addProtocol")
	purego.RegisterLibFunc(&class_getInstanceSize, objc, "class_getInstanceSize")
	purego.RegisterLibFunc(&ivar_getOffset, objc, "ivar_getOffset")
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

// Selector is an interface that takes a Go method
// and returns the selector equivalent name.
// If it returns an nil SEL then that method
// is not added to the class object.
type Selector interface {
	Selector(string) SEL
}

var TagFormatError = errors.New(`objc tag is doesn't match "ClassName : SuperClassName <Protocol, ...>""`)

func RegisterClass(object Selector) (Class, error) {
	ptr := reflect.TypeOf(object)
	strct := ptr.Elem()
	if strct.NumField() == 0 {
		return 0, fmt.Errorf("struct doesn't have objc.Class as first field")
	}
	isa := strct.Field(0)
	if isa.Type != reflect.TypeOf(Class(0)) {
		return 0, fmt.Errorf("struct doesn't have objc.Class as first field")
	}
	tag := isa.Tag.Get("objc")
	if tag == "" {
		return 0, fmt.Errorf("missing objc tag: %w", TagFormatError)
	}
	var split = make([]string, 2) // start with two for ClassName : SuperClassName
	{
		// This is a simple parser for the objc tag that looks for the format
		//  	"ClassName : SuperClassName <Protocol, ...>"
		// It appends to the split variable with the [ClassName, SuperClassName, Protocol, ...]

		var i int  // from tag[0:i] is whatever identifier is next
		var r rune // r is the current rune
		skipSpace := func() {
			for _, c := range tag {
				if c == ' ' {
					tag = tag[1:]
					continue
				}
				break
			}
		}
		skipSpace()
		// get ClassName
		for i, r = range tag {
			if r == ' ' {
				break
			}
		}
		split[0] = tag[0:i] // store ClassName
		tag = tag[i:]

		skipSpace()

		// check for ':'
		if len(tag) > 0 && tag[0] != ':' {
			return 0, fmt.Errorf("missing ':': %w", TagFormatError)
		}
		tag = tag[1:] // skip ':'
		skipSpace()

		// get SuperClassName
		for i, r = range tag {
			if r == ' ' {
				break
			} else if i+1 == len(tag) {
				// if this is the last character in the string
				// make sure to increment i so that tag[:i]
				// includes the last character
				i++
				break
			}
		}
		split[1] = tag[:i] // store SuperClassName
		tag = tag[i:]      // drop SuperClassName
		skipSpace()
		if len(tag) > 0 {
			if tag[0] != '<' {
				return 0, fmt.Errorf("expected '<': %w", TagFormatError)
			}
			tag = tag[1:] // drop '<'
			// get Protocols
		outer:
			for {
				skipSpace()
				for i, r = range tag {
					switch r {
					case ' ':
						split = append(split, tag[:i])
						tag = tag[i:]
						continue outer
					case ',':
						// If there is actually an identifier - add it.
						if i > 0 {
							split = append(split, tag[:i])
							tag = tag[i:]
						} else {
							// Otherwise, drop ','
							tag = tag[1:]
						}
						continue outer
					case '>':
						// If there is actually an identifier - add it.
						if i > 0 {
							split = append(split, tag[:i])
							tag = tag[i:]
						}
						break outer
					}
				}
				return 0, fmt.Errorf("expected '>': %w", TagFormatError)
			}
		}
	}
	super := GetClass(split[1])
	class := objc_allocateClassPair(super, split[0], 0)
	if class == 0 {
		return 0, fmt.Errorf("failed to create class with name '%s'", split[0])
	}
	if len(split) > 2 {
		// Add Protocols
		for _, n := range split[2:] {
			succeed := class_addProtocol(class, objc_getProtocol(n))
			if !succeed {
				return 0, fmt.Errorf("couldn't add Protocol %s", n)
			}
		}
	}
	// Add exported methods based on the selectors returned from Selector(string) SEL
	for i := 0; i < ptr.NumMethod(); i++ {
		met := ptr.Method(i)
		// TODO: figure a better way to determine if this is the Selector interface method
		if met.Name == "Selector" {
			continue
		}
		sel := object.Selector(met.Name)
		if sel == 0 {
			continue
		}
		fn := met.Func.Interface()
		//TODO: catch NewIMP panics and return as error
		succeed := class_addMethod(class, sel, NewIMP(fn), funcTypeInfo(fn))
		if !succeed {
			return 0, fmt.Errorf("couldn't add Method %s", met.Name)
		}
	}
	// Add Ivars
	// Start at 1 because we skip the class object which is first
	for i := 1; i < strct.NumField(); i++ {
		f := strct.Field(i)
		size := f.Type.Size()
		alignment := uint8(math.Log2(float64(f.Type.Align())))
		succeed := class_addIvar(class, f.Name, size, alignment, typeInfoForType(f.Type))
		if !succeed {
			return 0, fmt.Errorf("couldn't add Ivar %s", f.Name)
		}
	}
	objc_registerClassPair(class)
	if size1, size2 := class_getInstanceSize(class), strct.Size(); size1 != size2 {
		return 0, fmt.Errorf("objective-c size and struct size don't match %d != %d", size1, size2)
	}
	return class, nil
}

const (
	encId          = "@"
	encClass       = "#"
	encSelector    = ":"
	encChar        = "c"
	encUChar       = "C"
	encShort       = "s"
	encUShort      = "S"
	encInt         = "i"
	encUInt        = "I"
	encLong        = "l"
	encULong       = "L"
	encLongLong    = "q"
	encULongLong   = "Q"
	encFloat       = "f"
	encDouble      = "d"
	encDFLD        = "b"
	encBool        = "B"
	encVoid        = "v"
	encUndef       = "?"
	encPtr         = "^"
	encCharPtr     = "*"
	encAtom        = "%"
	encArrayBegin  = "["
	encArrayEnd    = "]"
	encUnionBegin  = "("
	encUnionEnd    = ")"
	encStructBegin = "{"
	encStructEnd   = "}"
	encVector      = "!"
	encConst       = "r"
)

func typeInfoForType(typ reflect.Type) string {
	if typ == reflect.TypeOf(Class(0)) {
		return encClass
	} else if typ == reflect.TypeOf(ID(0)) {
		return encId
	} else if typ == reflect.TypeOf(SEL(0)) {
		return encSelector
	}

	kind := typ.Kind()
	switch kind {
	case reflect.Bool:
		return encBool
	case reflect.Int:
		return encInt
	case reflect.Int8:
		return encChar
	case reflect.Int16:
		return encShort
	case reflect.Int32:
		return encInt
	case reflect.Int64:
		return encULong
	case reflect.Uint:
		return encUInt
	case reflect.Uint8:
		return encUChar
	case reflect.Uint16:
		return encUShort
	case reflect.Uint32:
		return encUInt
	case reflect.Uint64:
		return encULong
	case reflect.Uintptr:
		return encPtr
	case reflect.Float32:
		return encFloat
	case reflect.Float64:
		return encDouble
	case reflect.Ptr:
		return encPtr
	}

	panic("typeinfo: unhandled/invalid kind " + fmt.Sprintf("%v", kind) + " " + fmt.Sprintf("%v", typ))
}

// Returns the function's typeInfo
func funcTypeInfo(fn interface{}) string {
	typ := reflect.TypeOf(fn)
	kind := typ.Kind()
	if kind != reflect.Func {
		panic("not a func")
	}

	typeInfo := ""
	numOut := typ.NumOut()
	switch numOut {
	case 0:
		typeInfo += encVoid
	case 1:
		typeInfo += typeInfoForType(typ.Out(0))
	default:
		panic("too many output parameters")
	}

	if typ.NumIn() == 0 {
		panic("funcTypeInfo: bad func")
	}

	typeInfo += typeInfoForType(typ.In(0))
	typeInfo += encSelector

	for i := 1; i < typ.NumIn(); i++ {
		typeInfo += typeInfoForType(typ.In(i))
	}
	return typeInfo
}

func (c Class) AddMethod2(selector string, fn interface{}) bool {
	val := reflect.ValueOf(fn)
	ty := val.Type()
	if ty.Kind() != reflect.Func {
		return true // TODO: ...
	}
	info := funcTypeInfo(fn)
	fmt.Println(info)
	return class_addMethod(c, RegisterName(selector), NewIMP(fn), info)
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

// InstanceSize returns the size in bytes of instances of the class or 0 if cls is nil
func (c Class) InstanceSize() uintptr {
	return class_getInstanceSize(c)
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

// NewIMP takes a Go function that takes (ID, SEL) as its first two arguments.
// ID may instead be a pointer to a struct whose first field has type Class.
// It returns an IMP function pointer that can be called by Objective-C code.
// The function pointer is never deallocated.
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
	case ty.In(0).Kind() != reflect.Uintptr && // checks if it's objc.ID
		// or that it's a pointer to a struct
		(ty.In(0).Kind() != reflect.Pointer || ty.In(0).Elem().Kind() != reflect.Struct ||
			// and that the structs first field is an objc.Class
			ty.In(0).Elem().Field(0).Type != reflect.TypeOf(Class(0))):
		fallthrough
	case ty.In(1).Kind() != reflect.Uintptr:
		panic("objc: NewIMP must take a (id, SEL) as its first two arguments" + ty.String())
	}
	return IMP(purego.NewCallback(fn))
}
