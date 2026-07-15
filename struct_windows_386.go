// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 The Ebitengine Authors

package purego

import (
	"math"
	"reflect"
	"unsafe"
)

// Win32 passes struct arguments inline on the stack. Each argument consumes
// enough four-byte slots to hold its bytes; padding in the final slot is not
// part of the value and is cleared here for deterministic calls.
func addStruct(v reflect.Value, numInts, numFloats, numStack *int, addInt, addFloat, addStack func(uintptr), keepAlive []any) []any {
	size := v.Type().Size()
	if size == 0 {
		// Empty structs are a GNU C extension. The Windows GNU ABI does not
		// allocate a stack slot for them.
		return keepAlive
	}

	copyOfValue := reflect.New(v.Type())
	copyOfValue.Elem().Set(v)
	base := unsafe.Pointer(copyOfValue.Pointer())
	for offset := uintptr(0); offset < size; offset += unsafe.Sizeof(uintptr(0)) {
		var slot uintptr
		bytesLeft := size - offset
		if bytesLeft > unsafe.Sizeof(slot) {
			bytesLeft = unsafe.Sizeof(slot)
		}
		dst := unsafe.Slice((*byte)(unsafe.Pointer(&slot)), int(bytesLeft))
		src := unsafe.Slice((*byte)(unsafe.Add(base, offset)), int(bytesLeft))
		copy(dst, src)
		addInt(slot)
	}
	return keepAlive
}

// structReturnInMemory reports whether Win32 returns the struct through a
// caller-allocated hidden pointer passed as the first stack argument. One-,
// two-, and four-byte structs are returned in EAX and eight-byte structs in
// EDX:EAX.
func structReturnInMemory(outType reflect.Type) bool {
	switch outType.Size() {
	case 0, 1, 2, 4, 8:
		return false
	default:
		return true
	}
}

func getStruct(outType reflect.Type, syscall syscallArgs) reflect.Value {
	if isSingleFloatStruct(outType) && syscall.floatReturn != 0 {
		bits := uint64(syscall.f1) | uint64(syscall.f2)<<32
		if outType.Field(0).Type.Kind() == reflect.Float32 {
			result := uintptr(math.Float32bits(float32(math.Float64frombits(bits))))
			return reflect.NewAt(outType, unsafe.Pointer(&result)).Elem()
		}
		result := struct {
			low  uintptr
			high uintptr
		}{uintptr(bits), uintptr(bits >> 32)}
		return reflect.NewAt(outType, unsafe.Pointer(&result)).Elem()
	}
	switch size := outType.Size(); {
	case size == 0:
		return reflect.New(outType).Elem()
	case structReturnInMemory(outType):
		// Win32 aggregate-return functions return the hidden result pointer in
		// EAX after filling it.
		return reflect.NewAt(outType, *(*unsafe.Pointer)(unsafe.Pointer(&syscall.a1))).Elem()
	case size <= unsafe.Sizeof(uintptr(0)):
		result := syscall.a1
		return reflect.NewAt(outType, unsafe.Pointer(&result)).Elem()
	default:
		result := struct {
			low  uintptr
			high uintptr
		}{syscall.a1, syscall.a2}
		return reflect.NewAt(outType, unsafe.Pointer(&result)).Elem()
	}
}

func shouldBundleStackArgs(v reflect.Value, numInts, numFloats int) bool {
	return false
}

func collectStackArgs(args []reflect.Value, startIdx int, numInts, numFloats int,
	keepAlive []any, addInt, addFloat, addStack func(uintptr),
	pNumInts, pNumFloats, pNumStack *int) ([]reflect.Value, []any) {
	panic("purego: collectStackArgs should not be called on windows/386")
}

func bundleStackArgs(stackArgs []reflect.Value, addStack func(uintptr)) {
	panic("purego: bundleStackArgs should not be called on windows/386")
}

func getCallbackStruct(inType reflect.Type, frame unsafe.Pointer, floatsN *int, intsN *int, stackSlot *int, stackByteOffset *uintptr) reflect.Value {
	panic("purego: struct callback arguments are not supported on windows/386")
}

func setStruct(a *callbackArgs, ret reflect.Value) {
	panic("purego: struct callback returns are not supported on windows/386")
}
