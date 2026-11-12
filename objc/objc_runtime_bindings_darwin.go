// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 The Ebitengine Authors

//go:build go1.28

package objc

// Bindings for the Objective-C runtime functions that have one fixed C
// signature (golang/go#81450). cmd/cgo generates a Go-assembly trampoline for
// each one, so the calls skip RegisterLibFunc's reflection and no C compiler is
// needed. This file is left out with CGO_ENABLED=0; objc then falls back to
// registerRuntime.
//
// The bindings use plain uintptr, pointer and integer types rather than ID, SEL
// and Class, and the wrappers below convert. That keeps the binding signatures to
// types cmd/cgo's ABI code certainly understands. If named uintptr types turn out
// to be fine, most wrappers can go away.
import "C"

import (
	"runtime"
	"unsafe"

	"github.com/ebitengine/purego/internal/strings"
)

// Set during package variable initialization, which happens before any init
// function in the package runs, so objc's init sees it whatever the file order.
var _ = func() bool { bindRuntime = useRuntimeBindings; return true }()

//cgo:binding C.object_getClass "_"
func c_object_getClass(obj uintptr) uintptr

//cgo:binding C.objc_getClass "_"
func c_objc_getClass(name *byte) uintptr

//cgo:binding C.objc_getProtocol "_"
func c_objc_getProtocol(name *byte) unsafe.Pointer

//cgo:binding C.objc_allocateProtocol "_"
func c_objc_allocateProtocol(name *byte) unsafe.Pointer

//cgo:binding C.objc_registerProtocol "_"
func c_objc_registerProtocol(p unsafe.Pointer)

//cgo:binding C.objc_allocateClassPair "_"
func c_objc_allocateClassPair(super uintptr, name *byte, extraBytes uintptr) uintptr

//cgo:binding C.objc_registerClassPair "_"
func c_objc_registerClassPair(class uintptr)

//cgo:binding C.sel_registerName "_"
func c_sel_registerName(name *byte) uintptr

//cgo:binding C.class_getSuperclass "_"
func c_class_getSuperclass(class uintptr) uintptr

//cgo:binding C.class_getInstanceVariable "_"
func c_class_getInstanceVariable(class uintptr, name *byte) uintptr

//cgo:binding C.class_getInstanceSize "_"
func c_class_getInstanceSize(class uintptr) uintptr

//cgo:binding C.class_addMethod "_"
func c_class_addMethod(class, name, imp uintptr, types *byte) bool

//cgo:binding C.class_addIvar "_"
func c_class_addIvar(class uintptr, name *byte, size uintptr, alignment uint8, types *byte) bool

//cgo:binding C.class_addProtocol "_"
func c_class_addProtocol(class uintptr, p unsafe.Pointer) bool

//cgo:binding C.ivar_getOffset "_"
func c_ivar_getOffset(ivar uintptr) uintptr

//cgo:binding C.ivar_getName "_"
func c_ivar_getName(ivar uintptr) *byte

//cgo:binding C.object_getIvar "_"
func c_object_getIvar(obj, ivar uintptr) uintptr

//cgo:binding C.object_setIvar "_"
func c_object_setIvar(obj, ivar, value uintptr)

//cgo:binding C.protocol_getName "_"
func c_protocol_getName(p unsafe.Pointer) *byte

//cgo:binding C.protocol_isEqual "_"
func c_protocol_isEqual(p, p2 unsafe.Pointer) bool

//cgo:binding C.protocol_addMethodDescription "_"
func c_protocol_addMethodDescription(p unsafe.Pointer, name uintptr, types *byte, isRequiredMethod, isInstanceMethod bool)

//cgo:binding C.protocol_copyMethodDescriptionList "_"
func c_protocol_copyMethodDescriptionList(p unsafe.Pointer, isRequiredMethod, isInstanceMethod bool, outCount *uint32) unsafe.Pointer

//cgo:binding C.protocol_copyProtocolList "_"
func c_protocol_copyProtocolList(p unsafe.Pointer, outCount *uint32) unsafe.Pointer

//cgo:binding C.protocol_copyPropertyList2 "_"
func c_protocol_copyPropertyList2(p unsafe.Pointer, outCount *uint32, isRequiredProperty, isInstanceProperty bool) unsafe.Pointer

//cgo:binding C.protocol_addProtocol "_"
func c_protocol_addProtocol(p, p2 unsafe.Pointer)

// attributes is passed as a uintptr: see the wrapper below.
//
//cgo:binding C.protocol_addProperty "_"
func c_protocol_addProperty(p unsafe.Pointer, name *byte, attributes uintptr, attributeCount uint32, isRequiredProperty, isInstanceProperty bool)

//cgo:binding C.property_getName "_"
func c_property_getName(p uintptr) *byte

//cgo:binding C.property_getAttributes "_"
func c_property_getAttributes(p uintptr) *byte

// free, _Block_copy and _Block_release live in libSystem, which libobjc links
// against. RegisterLibFunc found them through the libobjc handle and RTLD_DEFAULT.

//cgo:binding C.free "_"
func c_free(ptr unsafe.Pointer)

//cgo:binding C._Block_copy "_"
func c__Block_copy(block uintptr) uintptr

//cgo:binding C._Block_release "_"
func c__Block_release(block uintptr)

func goString(p *byte) string { return strings.GoString(uintptr(unsafe.Pointer(p))) }

func useRuntimeBindings() {
	object_getClass = func(obj ID) Class { return Class(c_object_getClass(uintptr(obj))) }
	objc_getClass = func(name string) Class { return Class(c_objc_getClass(strings.CString(name))) }
	objc_getProtocol = func(name string) *Protocol {
		return (*Protocol)(c_objc_getProtocol(strings.CString(name)))
	}
	objc_allocateProtocol = func(name string) *Protocol {
		return (*Protocol)(c_objc_allocateProtocol(strings.CString(name)))
	}
	objc_registerProtocol = func(p *Protocol) { c_objc_registerProtocol(unsafe.Pointer(p)) }
	objc_allocateClassPair = func(super Class, name string, extraBytes uintptr) Class {
		return Class(c_objc_allocateClassPair(uintptr(super), strings.CString(name), extraBytes))
	}
	objc_registerClassPair = func(class Class) { c_objc_registerClassPair(uintptr(class)) }
	sel_registerName = func(name string) SEL { return SEL(c_sel_registerName(strings.CString(name))) }
	class_getSuperclass = func(class Class) Class { return Class(c_class_getSuperclass(uintptr(class))) }
	class_getInstanceVariable = func(class Class, name string) Ivar {
		return Ivar(c_class_getInstanceVariable(uintptr(class), strings.CString(name)))
	}
	class_getInstanceSize = func(class Class) uintptr { return c_class_getInstanceSize(uintptr(class)) }
	class_addMethod = func(class Class, name SEL, imp IMP, types string) bool {
		return c_class_addMethod(uintptr(class), uintptr(name), uintptr(imp), strings.CString(types))
	}
	class_addIvar = func(class Class, name string, size uintptr, alignment uint8, types string) bool {
		return c_class_addIvar(uintptr(class), strings.CString(name), size, alignment, strings.CString(types))
	}
	class_addProtocol = func(class Class, p *Protocol) bool {
		return c_class_addProtocol(uintptr(class), unsafe.Pointer(p))
	}
	ivar_getOffset = func(ivar Ivar) uintptr { return c_ivar_getOffset(uintptr(ivar)) }
	ivar_getName = func(ivar Ivar) string { return goString(c_ivar_getName(uintptr(ivar))) }
	object_getIvar = func(obj ID, ivar Ivar) ID { return ID(c_object_getIvar(uintptr(obj), uintptr(ivar))) }
	object_setIvar = func(obj ID, ivar Ivar, value ID) {
		c_object_setIvar(uintptr(obj), uintptr(ivar), uintptr(value))
	}
	protocol_getName = func(p *Protocol) string { return goString(c_protocol_getName(unsafe.Pointer(p))) }
	protocol_isEqual = func(p, p2 *Protocol) bool {
		return c_protocol_isEqual(unsafe.Pointer(p), unsafe.Pointer(p2))
	}
	protocol_addMethodDescription = func(p *Protocol, name SEL, types string, isRequiredMethod, isInstanceMethod bool) {
		c_protocol_addMethodDescription(unsafe.Pointer(p), uintptr(name), strings.CString(types), isRequiredMethod, isInstanceMethod)
	}
	protocol_copyMethodDescriptionList = func(p *Protocol, isRequiredMethod, isInstanceMethod bool, outCount *uint32) *MethodDescription {
		return (*MethodDescription)(c_protocol_copyMethodDescriptionList(unsafe.Pointer(p), isRequiredMethod, isInstanceMethod, outCount))
	}
	protocol_copyProtocolList = func(p *Protocol, outCount *uint32) **Protocol {
		return (**Protocol)(c_protocol_copyProtocolList(unsafe.Pointer(p), outCount))
	}
	protocol_copyPropertyList2 = func(p *Protocol, outCount *uint32, isRequiredProperty, isInstanceProperty bool) *Property {
		return (*Property)(c_protocol_copyPropertyList2(unsafe.Pointer(p), outCount, isRequiredProperty, isInstanceProperty))
	}
	protocol_addProtocol = func(p, p2 *Protocol) { c_protocol_addProtocol(unsafe.Pointer(p), unsafe.Pointer(p2)) }
	protocol_addProperty = func(p *Protocol, name string, attributes []PropertyAttribute, attributeCount uint32, isRequiredProperty, isInstanceProperty bool) {
		// PropertyAttribute holds Go pointers (Name, Value), so passing the array as
		// a pointer would trip cgo's pointer checks. RegisterLibFunc passed it the
		// same way without a check. objc4 copies the attribute list before
		// returning, so keeping the array alive across the call is enough.
		c_protocol_addProperty(unsafe.Pointer(p), strings.CString(name),
			uintptr(unsafe.Pointer(unsafe.SliceData(attributes))), attributeCount, isRequiredProperty, isInstanceProperty)
		runtime.KeepAlive(attributes)
	}
	property_getName = func(p Property) string { return goString(c_property_getName(uintptr(p))) }
	property_getAttributes = func(p Property) string { return goString(c_property_getAttributes(uintptr(p))) }
	free = func(ptr unsafe.Pointer) { c_free(ptr) }
	_Block_copy = func(b Block) Block { return Block(c__Block_copy(uintptr(b))) }
	_Block_release = func(b Block) { c__Block_release(uintptr(b)) }
}
