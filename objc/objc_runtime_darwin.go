// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2022 The Ebitengine Authors

// Package objc is a low-level pure Go objective-c runtime. This package is easy to use incorrectly, so it is best
// to use a wrapper that provides the functionality you need in a safer way.
package objc

import (
	"errors"
	"fmt"
	"math"
	"reflect"
	"regexp"
	"runtime"
	"unicode"
	"unsafe"

	"github.com/ebitengine/purego"
	"github.com/ebitengine/purego/internal/strings"
)

// TODO: support try/catch?
// https://stackoverflow.com/questions/7062599/example-of-how-objective-cs-try-catch-implementation-is-executed-at-runtime
var (
	objc_msgSend_fn                    uintptr
	objc_msgSend_stret_fn              uintptr
	objc_msgSend                       func(obj ID, cmd SEL, args ...any) ID
	objc_msgSendSuper2_fn              uintptr
	objc_msgSendSuper2_stret_fn        uintptr
	objc_msgSendSuper2                 func(super *objc_super, cmd SEL, args ...any) ID
	objc_getClass                      func(name string) Class
	objc_getProtocol                   func(name string) *Protocol
	objc_allocateProtocol              func(name string) *Protocol
	objc_registerProtocol              func(protocol *Protocol)
	objc_allocateClassPair             func(super Class, name string, extraBytes uintptr) Class
	objc_registerClassPair             func(class Class)
	sel_registerName                   func(name string) SEL
	class_getSuperclass                func(class Class) Class
	class_getInstanceVariable          func(class Class, name string) Ivar
	class_getInstanceSize              func(class Class) uintptr
	class_addMethod                    func(class Class, name SEL, imp IMP, types string) bool
	class_addIvar                      func(class Class, name string, size uintptr, alignment uint8, types string) bool
	class_addProtocol                  func(class Class, protocol *Protocol) bool
	ivar_getOffset                     func(ivar Ivar) uintptr
	ivar_getName                       func(ivar Ivar) string
	object_getClass                    func(obj ID) Class
	object_getIvar                     func(obj ID, ivar Ivar) ID
	object_setIvar                     func(obj ID, ivar Ivar, value ID)
	protocol_getName                   func(protocol *Protocol) string
	protocol_isEqual                   func(p *Protocol, p2 *Protocol) bool
	protocol_addMethodDescription      func(p *Protocol, name SEL, types string, isRequiredMethod bool, isInstanceMethod bool)
	protocol_copyMethodDescriptionList func(p *Protocol, isRequiredMethod bool, isInstanceMethod bool, outCount *uint32) *MethodDescription
	protocol_copyProtocolList          func(p *Protocol, outCount *uint32) **Protocol
	protocol_copyPropertyList2         func(p *Protocol, outCount *uint32, isRequiredProperty, isInstanceProperty bool) *Property
	protocol_addProtocol               func(p *Protocol, p2 *Protocol)
	protocol_addProperty               func(p *Protocol, name string, attributes []PropertyAttribute, attributeCount uint32, isRequiredProperty bool, isInstanceProperty bool)
	property_getName                   func(p Property) string
	property_getAttributes             func(p Property) string

	free           func(ptr unsafe.Pointer)
	_Block_copy    func(Block) Block
	_Block_release func(Block)
)

func init() {
	objc, err := purego.Dlopen("/usr/lib/libobjc.A.dylib", purego.RTLD_GLOBAL|purego.RTLD_NOW)
	if err != nil {
		panic(fmt.Errorf("objc: %w", err))
	}
	objc_msgSend_fn, err = purego.Dlsym(objc, "objc_msgSend")
	if err != nil {
		panic(fmt.Errorf("objc: %w", err))
	}
	if runtime.GOARCH == "amd64" {
		objc_msgSend_stret_fn, err = purego.Dlsym(objc, "objc_msgSend_stret")
		if err != nil {
			panic(fmt.Errorf("objc: %w", err))
		}
		objc_msgSendSuper2_stret_fn, err = purego.Dlsym(objc, "objc_msgSendSuper2_stret")
		if err != nil {
			panic(fmt.Errorf("objc: %w", err))
		}
	}
	purego.RegisterFunc(&objc_msgSend, objc_msgSend_fn)
	objc_msgSendSuper2_fn, err = purego.Dlsym(objc, "objc_msgSendSuper2")
	if err != nil {
		panic(fmt.Errorf("objc: %w", err))
	}
	purego.RegisterFunc(&objc_msgSendSuper2, objc_msgSendSuper2_fn)
	purego.RegisterLibFunc(&object_getClass, objc, "object_getClass")
	purego.RegisterLibFunc(&objc_getClass, objc, "objc_getClass")
	purego.RegisterLibFunc(&objc_getProtocol, objc, "objc_getProtocol")
	purego.RegisterLibFunc(&objc_allocateProtocol, objc, "objc_allocateProtocol")
	purego.RegisterLibFunc(&objc_registerProtocol, objc, "objc_registerProtocol")
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
	purego.RegisterLibFunc(&ivar_getName, objc, "ivar_getName")
	purego.RegisterLibFunc(&protocol_getName, objc, "protocol_getName")
	purego.RegisterLibFunc(&protocol_isEqual, objc, "protocol_isEqual")
	purego.RegisterLibFunc(&protocol_addMethodDescription, objc, "protocol_addMethodDescription")
	purego.RegisterLibFunc(&protocol_copyMethodDescriptionList, objc, "protocol_copyMethodDescriptionList")
	purego.RegisterLibFunc(&protocol_copyProtocolList, objc, "protocol_copyProtocolList")
	purego.RegisterLibFunc(&protocol_addProtocol, objc, "protocol_addProtocol")
	purego.RegisterLibFunc(&protocol_addProperty, objc, "protocol_addProperty")
	purego.RegisterLibFunc(&protocol_copyPropertyList2, objc, "protocol_copyPropertyList2")
	purego.RegisterLibFunc(&property_getName, objc, "property_getName")
	purego.RegisterLibFunc(&property_getAttributes, objc, "property_getAttributes")
	purego.RegisterLibFunc(&object_getIvar, objc, "object_getIvar")
	purego.RegisterLibFunc(&object_setIvar, objc, "object_setIvar")
	purego.RegisterLibFunc(&free, purego.RTLD_DEFAULT, "free")

	purego.RegisterLibFunc(&_Block_copy, objc, "_Block_copy")
	purego.RegisterLibFunc(&_Block_release, objc, "_Block_release")
	theBlocksCache = newBlockCache()
}

// ID is an opaque pointer to some Objective-C object
type ID uintptr

// Class returns the class of the object.
func (id ID) Class() Class {
	return object_getClass(id)
}

// Send is a convenience method for sending messages to objects. This function takes a SEL
// instead of a string since RegisterName grabs the global Objective-C lock. It is best to cache the result
// of RegisterName.
func (id ID) Send(sel SEL, args ...any) ID {
	return objc_msgSend(id, sel, args...)
}

// GetIvar reads the value of an instance variable in an object.
func (id ID) GetIvar(ivar Ivar) ID {
	return object_getIvar(id, ivar)
}

// SetIvar sets the value of an instance variable in an object.
func (id ID) SetIvar(ivar Ivar, value ID) {
	object_setIvar(id, ivar, value)
}

// keep in sync with func.go
const maxRegAllocStructSize = 16

// Send is a convenience method for sending messages to objects that can return any type.
// This function takes a SEL instead of a string since RegisterName grabs the global Objective-C lock.
// It is best to cache the result of RegisterName.
func Send[T any](id ID, sel SEL, args ...any) T {
	var fn func(id ID, sel SEL, args ...any) T
	var zero T
	if runtime.GOARCH == "amd64" &&
		reflect.ValueOf(zero).Kind() == reflect.Struct &&
		reflect.ValueOf(zero).Type().Size() > maxRegAllocStructSize {
		purego.RegisterFunc(&fn, objc_msgSend_stret_fn)
	} else {
		purego.RegisterFunc(&fn, objc_msgSend_fn)
	}
	return fn(id, sel, args...)
}

// objc_super data structure is generated by the Objective-C compiler when it encounters the super keyword
// as the receiver of a message. It specifies the class definition of the particular superclass that should
// be messaged.
type objc_super struct {
	receiver   ID
	superClass Class
}

// SendSuper is a convenience method for sending message to object's super. This function takes a SEL
// instead of a string since RegisterName grabs the global Objective-C lock. It is best to cache the result
// of RegisterName.
func (id ID) SendSuper(sel SEL, args ...any) ID {
	super := &objc_super{
		receiver:   id,
		superClass: id.Class(),
	}
	return objc_msgSendSuper2(super, sel, args...)
}

// SendSuper is a convenience method for sending message to object's super that can return any type.
// This function takes a SEL instead of a string since RegisterName grabs the global Objective-C lock.
// It is best to cache the result of RegisterName.
func SendSuper[T any](id ID, sel SEL, args ...any) T {
	super := &objc_super{
		receiver:   id,
		superClass: id.Class(),
	}
	var fn func(objcSuper *objc_super, sel SEL, args ...any) T
	var zero T
	if runtime.GOARCH == "amd64" &&
		reflect.ValueOf(zero).Kind() == reflect.Struct &&
		reflect.ValueOf(zero).Type().Size() > maxRegAllocStructSize {
		purego.RegisterFunc(&fn, objc_msgSendSuper2_stret_fn)
	} else {
		purego.RegisterFunc(&fn, objc_msgSendSuper2_fn)
	}
	return fn(super, sel, args...)
}

// SEL is an opaque type that represents a method selector
type SEL uintptr

// RegisterName registers a method with the Objective-C runtime system, maps the method name to a selector,
// and returns the selector value. This function grabs the global Objective-c lock. It is best the cache the
// result of this function.
func RegisterName(name string) SEL {
	return sel_registerName(name)
}

// Class is an opaque type that represents an Objective-C class.
type Class uintptr

// GetClass returns the Class object for the named class, or nil if the class is not registered with the Objective-C runtime.
func GetClass(name string) Class {
	return objc_getClass(name)
}

// MethodDef represents the Go function and the selector that ObjC uses to access that function.
type MethodDef struct {
	Cmd SEL
	Fn  any
}

// IvarAttrib is the attribute that an ivar has. It affects if and which methods are automatically
// generated when creating a class with RegisterClass. See [Apple Docs] for an understanding of these attributes.
// The fields are still accessible using objc.GetIvar and objc.SetIvar regardless of the value of IvarAttrib.
//
// Take for example this Objective-C code:
//
//	@property (readwrite) float value;
//
// In Go, the functions can be accessed as followed:
//
//	var value = purego.Send[float32](id, purego.RegisterName("value"))
//	id.Send(purego.RegisterName("setValue:"), 3.46)
//
// [Apple Docs]: https://developer.apple.com/library/archive/documentation/Cocoa/Conceptual/ObjectiveC/Chapters/ocProperties.html
type IvarAttrib int

const (
	ReadOnly IvarAttrib = 1 << iota
	ReadWrite
)

// FieldDef is a definition of a field to add to an Objective-C class.
// The name of the field is what will be used to access it through the Ivar. If the type is bool
// the name cannot start with `is` since a getter will be generated with the name `isBoolName`.
// The name also cannot contain any spaces.
// The type is the Go equivalent type of the Ivar.
// Attribute determines if a getter and or setter method is generated for this field.
type FieldDef struct {
	Name      string
	Type      reflect.Type
	Attribute IvarAttrib
}

// ivarRegex checks to make sure the Ivar is correctly formatted
var ivarRegex = regexp.MustCompile("[a-z_][a-zA-Z0-9_]*")

// RegisterClass takes the name of the class to create, the superclass, a list of protocols this class
// implements, a list of fields this class has and a list of methods. It returns the created class or an error
// describing what went wrong.
func RegisterClass(name string, superClass Class, protocols []*Protocol, ivars []FieldDef, methods []MethodDef) (Class, error) {
	class := objc_allocateClassPair(superClass, name, 0)
	if class == 0 {
		return 0, fmt.Errorf("objc: failed to create class with name '%s'", name)
	}
	// Add Protocols
	for _, p := range protocols {
		if !class.AddProtocol(p) {
			return 0, fmt.Errorf("objc: couldn't add Protocol %s", protocol_getName(p))
		}
	}
	// Add exported methods based on the selectors returned from ClassDef(string) SEL
	for idx, def := range methods {
		imp, err := func() (imp IMP, err error) {
			defer func() {
				if r := recover(); r != nil {
					err = fmt.Errorf("objc: failed to create IMP: %s", r)
				}
			}()
			return NewIMP(def.Fn), nil
		}()
		if err != nil {
			return 0, fmt.Errorf("objc: couldn't add Method at index %d: %w", idx, err)
		}
		encoding, err := encodeFunc(def.Fn)
		if err != nil {
			return 0, fmt.Errorf("objc: couldn't add Method at index %d: %w", idx, err)
		}
		if !class.AddMethod(def.Cmd, imp, encoding) {
			return 0, fmt.Errorf("objc: couldn't add Method at index %d", idx)
		}
	}
	// Add Ivars
	for _, instVar := range ivars {
		ivar := instVar
		if !ivarRegex.MatchString(ivar.Name) {
			return 0, fmt.Errorf("objc: Ivar must start with a lowercase letter and only contain ASCII letters and numbers: '%s'", ivar.Name)
		}
		size := ivar.Type.Size()
		alignment := uint8(math.Log2(float64(ivar.Type.Align())))
		enc, err := encodeType(ivar.Type, false)
		if err != nil {
			return 0, fmt.Errorf("objc: couldn't add Ivar %s: %w", ivar.Name, err)
		}
		if !class_addIvar(class, ivar.Name, size, alignment, enc) {
			return 0, fmt.Errorf("objc: couldn't add Ivar %s", ivar.Name)
		}
		offset := class.InstanceVariable(ivar.Name).Offset()
		switch ivar.Attribute {
		case ReadWrite:
			ty := reflect.FuncOf(
				[]reflect.Type{
					reflect.TypeOf(ID(0)), reflect.TypeOf(SEL(0)), ivar.Type,
				},
				nil, false,
			)
			var encoding string
			if encoding, err = encodeFunc(reflect.New(ty).Elem().Interface()); err != nil {
				return 0, fmt.Errorf("objc: failed to create read method for '%s': %w", ivar.Name, err)
			}
			val := reflect.MakeFunc(ty, func(args []reflect.Value) (results []reflect.Value) {
				// on entry the first and second arguments are ID and SEL followed by the value
				if len(args) != 3 {
					panic(fmt.Sprintf("objc: incorrect number of args. expected 3 got %d", len(args)))
				}
				// The following reflect code does the equivalent of this:
				//
				//	((*struct {
				//		Padding [offset]byte
				//		Value int
				//	})(unsafe.Pointer(args[0].Interface().(ID)))).v = 123
				//
				// However, since the type of the variable is unknown reflection is used to actually assign the value
				id := args[0].Interface().(ID)
				ptr := *(*unsafe.Pointer)(unsafe.Pointer(&id)) // circumvent go vet
				reflect.NewAt(ivar.Type, unsafe.Add(ptr, offset)).Elem().Set(args[2])
				return nil
			}).Interface()
			// this code only works for ascii but that shouldn't be a problem
			selector := "set" + string(unicode.ToUpper(rune(ivar.Name[0]))) + ivar.Name[1:] + ":\x00"
			class.AddMethod(RegisterName(selector), NewIMP(val), encoding)
			fallthrough // also implement the read method
		case ReadOnly:
			ty := reflect.FuncOf(
				[]reflect.Type{
					reflect.TypeOf(ID(0)), reflect.TypeOf(SEL(0)),
				},
				[]reflect.Type{ivar.Type}, false,
			)
			var encoding string
			if encoding, err = encodeFunc(reflect.New(ty).Elem().Interface()); err != nil {
				return 0, fmt.Errorf("objc: failed to create read method for '%s': %w", ivar.Name, err)
			}
			val := reflect.MakeFunc(ty, func(args []reflect.Value) (results []reflect.Value) {
				// on entry the first and second arguments are ID and SEL
				if len(args) != 2 {
					panic(fmt.Sprintf("objc: incorrect number of args. expected 2 got %d", len(args)))
				}
				id := args[0].Interface().(ID)
				ptr := *(*unsafe.Pointer)(unsafe.Pointer(&id)) // circumvent go vet
				// the variable is located at an offset from the id
				return []reflect.Value{reflect.NewAt(ivar.Type, unsafe.Add(ptr, offset)).Elem()}
			}).Interface()
			if ivar.Type.Kind() == reflect.Bool {
				// this code only works for ascii but that shouldn't be a problem
				ivar.Name = "is" + string(unicode.ToUpper(rune(ivar.Name[0]))) + ivar.Name[1:]
			}
			class.AddMethod(RegisterName(ivar.Name), NewIMP(val), encoding)
		default:
			return 0, fmt.Errorf("objc: unknown Ivar Attribute (%d)", ivar.Attribute)
		}
	}
	objc_registerClassPair(class)
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
	encFloat       = "f"
	encDouble      = "d"
	encBool        = "B"
	encVoid        = "v"
	encPtr         = "^"
	encCharPtr     = "*"
	encStructBegin = "{"
	encStructEnd   = "}"
	encUnsafePtr   = "^v"
)

// encodeType returns a string representing a type as if it was given to @encode(typ)
// Source: https://developer.apple.com/library/archive/documentation/Cocoa/Conceptual/ObjCRuntimeGuide/Articles/ocrtTypeEncodings.html#//apple_ref/doc/uid/TP40008048-CH100
func encodeType(typ reflect.Type, insidePtr bool) (string, error) {
	switch typ {
	case reflect.TypeOf(Class(0)):
		return encClass, nil
	case reflect.TypeOf(ID(0)), reflect.TypeOf(Block(0)):
		return encId, nil
	case reflect.TypeOf(SEL(0)):
		return encSelector, nil
	}

	kind := typ.Kind()
	switch kind {
	case reflect.Bool:
		return encBool, nil
	case reflect.Int:
		return encLong, nil
	case reflect.Int8:
		return encChar, nil
	case reflect.Int16:
		return encShort, nil
	case reflect.Int32:
		return encInt, nil
	case reflect.Int64:
		return encULong, nil
	case reflect.Uint:
		return encULong, nil
	case reflect.Uint8:
		return encUChar, nil
	case reflect.Uint16:
		return encUShort, nil
	case reflect.Uint32:
		return encUInt, nil
	case reflect.Uint64:
		return encULong, nil
	case reflect.Uintptr:
		return encPtr, nil
	case reflect.Float32:
		return encFloat, nil
	case reflect.Float64:
		return encDouble, nil
	case reflect.Ptr:
		enc, err := encodeType(typ.Elem(), true)
		return encPtr + enc, err
	case reflect.Struct:
		if insidePtr {
			return encStructBegin + typ.Name() + encStructEnd, nil
		}
		encoding := encStructBegin
		encoding += typ.Name()
		encoding += "="
		for i := 0; i < typ.NumField(); i++ {
			f := typ.Field(i)
			tmp, err := encodeType(f.Type, false)
			if err != nil {
				return "", err
			}
			encoding += tmp
		}
		encoding += encStructEnd
		return encoding, nil
	case reflect.UnsafePointer:
		return encUnsafePtr, nil
	case reflect.String:
		return encCharPtr, nil
	}

	return "", errors.New(fmt.Sprintf("unhandled/invalid kind %v typed %v", kind, typ))
}

// encodeFunc returns a functions type as if it was given to @encode(fn)
func encodeFunc(fn any) (string, error) {
	typ := reflect.TypeOf(fn)
	if typ.Kind() != reflect.Func {
		return "", errors.New("not a func")
	}

	encoding := ""
	switch typ.NumOut() {
	case 0:
		encoding += encVoid
	case 1:
		tmp, err := encodeType(typ.Out(0), false)
		if err != nil {
			return "", err
		}
		encoding += tmp
	default:
		return "", errors.New("too many output parameters")
	}

	if typ.NumIn() < 2 {
		return "", errors.New("func doesn't take ID and SEL as its first two parameters")
	}

	encoding += encId

	for i := 1; i < typ.NumIn(); i++ {
		tmp, err := encodeType(typ.In(i), false)
		if err != nil {
			return "", err
		}
		encoding += tmp
	}
	return encoding, nil
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

// Ivar an opaque type that represents an instance variable.
type Ivar uintptr

// Offset returns the offset of an instance variable that can be used to assign and read the Ivar's value.
//
// For instance variables of type ID or other object types, call Ivar and SetIvar instead
// of using this offset to access the instance variable data directly.
func (i Ivar) Offset() uintptr {
	return ivar_getOffset(i)
}

func (i Ivar) Name() string {
	return ivar_getName(i)
}

// MethodDescription holds the name and type definition of a method.
type MethodDescription struct {
	name, types uintptr
}

// Name returns the name of this method.
func (m MethodDescription) Name() string {
	return strings.GoString(m.name)
}

// Types returns the OBJC runtime encoded type description.
func (m MethodDescription) Types() string {
	return strings.GoString(m.types)
}

// PropertyAttribute contains the null-terminated Name and Value pair of a Properties internal description.
type PropertyAttribute struct {
	Name, Value *byte
}

// Property is an opaque type for Objective-C property metadata.
type Property uintptr

// Name returns the name of this property.
func (p Property) Name() string {
	return property_getName(p)
}

// Attributes returns a comma separated list of PropertyAttribute
func (p Property) Attributes() string {
	return property_getAttributes(p)
}

// Protocol is a type that declares methods that can be implemented by any class.
type Protocol [0]func()

// GetProtocol returns the protocol for the given name or nil if there is no protocol by that name.
func GetProtocol(name string) *Protocol {
	return objc_getProtocol(name)
}

// AllocateProtocol creates a new protocol in the OBJC runtime or nil if the protocol already exists.
func AllocateProtocol(name string) *Protocol {
	return objc_allocateProtocol(name)
}

// Register registers the protocol created using AllocateProtocol with the runtime. This must be done
// before it is used anywhere and can only be called once.
func (p *Protocol) Register() {
	objc_registerProtocol(p)
}

// CopyMethodDescriptionList returns a list of methods that this protocol has given it isRequiredMethod and isInstanceMethod.
func (p *Protocol) CopyMethodDescriptionList(isRequiredMethod, isInstanceMethod bool) []MethodDescription {
	count := uint32(0)
	desc := protocol_copyMethodDescriptionList(p, isRequiredMethod, isInstanceMethod, &count)
	methods := clone(unsafe.Slice(desc, count))
	free(unsafe.Pointer(desc))
	return methods
}

// CopyProtocolList returns a list of the protocols that this protocol inherits from.
func (p *Protocol) CopyProtocolList() []*Protocol {
	count := uint32(0)
	desc := protocol_copyProtocolList(p, &count)
	protocols := clone(unsafe.Slice(desc, count))
	free(unsafe.Pointer(desc))
	return protocols
}

// CopyPropertyList returns a list of properties that this protocol has given it isRequiredProperty and isInstanceProperty.
func (p *Protocol) CopyPropertyList(isRequiredProperty, isInstanceProperty bool) []Property {
	count := uint32(0)
	desc := protocol_copyPropertyList2(p, &count, isRequiredProperty, isInstanceProperty)
	protocols := clone(unsafe.Slice(desc, count))
	free(unsafe.Pointer(desc))
	return protocols
}

// Name returns the name of this protocol.
func (p *Protocol) Name() string {
	return protocol_getName(p)
}

// Equals return true if the two protocols are the same.
func (p *Protocol) Equals(p2 *Protocol) bool {
	return protocol_isEqual(p, p2)
}

// AddMethodDescription adds a method to a protocol. This can only be called between AllocateProtocol and Protocol.Register.
func (p *Protocol) AddMethodDescription(name SEL, types string, isRequiredMethod, isInstanceMethod bool) {
	protocol_addMethodDescription(p, name, types, isRequiredMethod, isInstanceMethod)
}

// AddProtocol marks the protocol as inheriting from another. This can only be called between AllocateProtocol and Protocol.Register.
func (p *Protocol) AddProtocol(protocol *Protocol) {
	protocol_addProtocol(p, protocol)
}

// AddProperty adds a property to the protocol. This can only be called between AllocateProtocol and Protocol.Register.
func (p *Protocol) AddProperty(name string, attributes []PropertyAttribute, isRequiredProperty, isInstanceProperty bool) {
	protocol_addProperty(p, name, attributes, uint32(len(attributes)), isRequiredProperty, isInstanceProperty)
}

// IMP is a function pointer that can be called by Objective-C code.
type IMP uintptr

// NewIMP takes a Go function that takes (ID, SEL) as its first two arguments.
// It returns an IMP function pointer that can be called by Objective-C code.
// The function panics if an error occurs.
// The function pointer is never deallocated.
func NewIMP(fn any) IMP {
	ty := reflect.TypeOf(fn)
	if ty.Kind() != reflect.Func {
		panic("objc: not a function")
	}
	// IMP is stricter than a normal callback
	// id (*IMP)(id, SEL, ...)
	switch {
	case ty.NumIn() < 2:
		fallthrough
	case ty.In(0) != reflect.TypeOf(ID(0)):
		fallthrough
	case ty.In(1) != reflect.TypeOf(SEL(0)):
		panic("objc: NewIMP must take a (id, SEL) as its first two arguments; got " + ty.String())
	}
	return IMP(purego.NewCallback(fn))
}

// TODO: remove and use slices.Clone when minimum version for purego is 1.21
func clone[S ~[]E, E any](s S) S {
	// Preserve nilness in case it matters.
	if s == nil {
		return nil
	}
	// Avoid s[:0:0] as it leads to unwanted liveness when cloning a
	// zero-length slice of a large array; see https://go.dev/issue/68488.
	return append(S{}, s...)
}
