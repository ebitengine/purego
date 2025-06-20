// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024 The Ebitengine Authors

package purego

import (
	"math"
	"reflect"
	"unsafe"
)

func getStruct(outType reflect.Type, syscall syscall15Args) (v reflect.Value) {
	outSize := outType.Size()
	switch {
	case outSize == 0:
		return reflect.New(outType).Elem()
	case outSize <= 8:
		r1 := syscall.a1
		if isAllFloats, numFields := isAllSameFloat(outType); isAllFloats {
			if numFields == 1 {
				// Single float32/float64 - always use float register
				r1 = syscall.f1
			} else if numFields == 2 {
				// Two float32s: check if the packed value in a1 makes sense as a struct
				// vs individual values in f1/f2 making sense as floats
				a1_as_struct := *(*[2]float32)(unsafe.Pointer(&syscall.a1))
				f_as_struct := [2]float32{math.Float32frombits(uint32(syscall.f1)), math.Float32frombits(uint32(syscall.f2))}

				// Use a1 if the float registers appear to have been derived from a1 (callback case)
				// or if f1/f2 are zero/uninitialized
				if syscall.f1 == 0 && syscall.f2 == 0 {
					// Definitely use a1
				} else if a1_as_struct[0] == f_as_struct[1] && a1_as_struct[1] == f_as_struct[0] {
					// Values are swapped between a1 and f1/f2 - this indicates callback packing
					// Use a1
				} else {
					// Use float registers (regular function call)
					r1 = syscall.f2<<32 | syscall.f1
				}
			}
		}
		return reflect.NewAt(outType, unsafe.Pointer(&struct{ a uintptr }{r1})).Elem()
	case outSize <= 16:
		r1, r2 := syscall.a1, syscall.a2
		if isAllFloats, numFields := isAllSameFloat(outType); isAllFloats {
			switch numFields {
			case 4:
				r1 = syscall.f2<<32 | syscall.f1
				r2 = syscall.f4<<32 | syscall.f3
			case 3:
				r1 = syscall.f2<<32 | syscall.f1
				r2 = syscall.f3
			case 2:
				r1 = syscall.f1
				r2 = syscall.f2
			default:
				panic("unreachable")
			}
		}
		return reflect.NewAt(outType, unsafe.Pointer(&struct{ a, b uintptr }{r1, r2})).Elem()
	default:
		if isAllFloats, numFields := isAllSameFloat(outType); isAllFloats && numFields <= 4 {
			switch numFields {
			case 4:
				return reflect.NewAt(outType, unsafe.Pointer(&struct{ a, b, c, d uintptr }{syscall.f1, syscall.f2, syscall.f3, syscall.f4})).Elem()
			case 3:
				return reflect.NewAt(outType, unsafe.Pointer(&struct{ a, b, c uintptr }{syscall.f1, syscall.f2, syscall.f3})).Elem()
			default:
				panic("unreachable")
			}
		}
		// create struct from the Go pointer created in arm64_r8
		// weird pointer dereference to circumvent go vet
		return reflect.NewAt(outType, *(*unsafe.Pointer)(unsafe.Pointer(&syscall.arm64_r8))).Elem()
	}
}

// https://github.com/ARM-software/abi-aa/blob/main/sysvabi64/sysvabi64.rst
const (
	_NO_CLASS = 0b00
	_FLOAT    = 0b01
	_INT      = 0b11
)

func addStruct(v reflect.Value, numInts, numFloats, numStack *int, addInt, addFloat, addStack func(uintptr), keepAlive []any) []any {
	if v.Type().Size() == 0 {
		return keepAlive
	}

	if hva, hfa, size := isHVA(v.Type()), isHFA(v.Type()), v.Type().Size(); hva || hfa || size <= 16 {
		// if this doesn't fit entirely in registers then
		// each element goes onto the stack
		if hfa && *numFloats+v.NumField() > numOfFloatRegisters {
			*numFloats = numOfFloatRegisters
		} else if hva && *numInts+v.NumField() > numOfIntegerRegisters() {
			*numInts = numOfIntegerRegisters()
		}

		placeRegisters(v, addFloat, addInt)
	} else {
		keepAlive = placeStack(v, keepAlive, addInt)
	}
	return keepAlive // the struct was allocated so don't panic
}

func placeRegisters(v reflect.Value, addFloat func(uintptr), addInt func(uintptr)) {
	var val uint64
	var shift byte
	var flushed bool
	class := _NO_CLASS
	var place func(v reflect.Value)
	place = func(v reflect.Value) {
		var numFields int
		if v.Kind() == reflect.Struct {
			numFields = v.Type().NumField()
		} else {
			numFields = v.Type().Len()
		}
		for k := 0; k < numFields; k++ {
			flushed = false
			var f reflect.Value
			if v.Kind() == reflect.Struct {
				f = v.Field(k)
			} else {
				f = v.Index(k)
			}
			align := byte(f.Type().Align()*8 - 1)
			shift = (shift + align) &^ align
			if shift >= 64 {
				shift = 0
				flushed = true
				if class == _FLOAT {
					addFloat(uintptr(val))
				} else {
					addInt(uintptr(val))
				}
			}
			switch f.Type().Kind() {
			case reflect.Struct:
				place(f)
			case reflect.Bool:
				if f.Bool() {
					val |= 1
				}
				shift += 8
				class |= _INT
			case reflect.Uint8:
				val |= f.Uint() << shift
				shift += 8
				class |= _INT
			case reflect.Uint16:
				val |= f.Uint() << shift
				shift += 16
				class |= _INT
			case reflect.Uint32:
				val |= f.Uint() << shift
				shift += 32
				class |= _INT
			case reflect.Uint64, reflect.Uint, reflect.Uintptr:
				addInt(uintptr(f.Uint()))
				shift = 0
				flushed = true
				class = _NO_CLASS
			case reflect.Int8:
				val |= uint64(f.Int()&0xFF) << shift
				shift += 8
				class |= _INT
			case reflect.Int16:
				val |= uint64(f.Int()&0xFFFF) << shift
				shift += 16
				class |= _INT
			case reflect.Int32:
				val |= uint64(f.Int()&0xFFFF_FFFF) << shift
				shift += 32
				class |= _INT
			case reflect.Int64, reflect.Int:
				addInt(uintptr(f.Int()))
				shift = 0
				flushed = true
				class = _NO_CLASS
			case reflect.Float32:
				if class == _FLOAT {
					addFloat(uintptr(val))
					val = 0
					shift = 0
				}
				val |= uint64(math.Float32bits(float32(f.Float()))) << shift
				shift += 32
				class |= _FLOAT
			case reflect.Float64:
				addFloat(uintptr(math.Float64bits(float64(f.Float()))))
				shift = 0
				flushed = true
				class = _NO_CLASS
			case reflect.Ptr:
				addInt(f.Pointer())
				shift = 0
				flushed = true
				class = _NO_CLASS
			case reflect.Array:
				place(f)
			default:
				panic("purego: unsupported kind " + f.Kind().String())
			}
		}
	}
	place(v)
	if !flushed {
		if class == _FLOAT {
			addFloat(uintptr(val))
		} else {
			addInt(uintptr(val))
		}
	}
}

func placeStack(v reflect.Value, keepAlive []any, addInt func(uintptr)) []any {
	// Struct is too big to be placed in registers.
	// Copy to heap and place the pointer in register
	ptrStruct := reflect.New(v.Type())
	ptrStruct.Elem().Set(v)
	ptr := ptrStruct.Elem().Addr().UnsafePointer()
	keepAlive = append(keepAlive, ptr)
	addInt(uintptr(ptr))
	return keepAlive
}

// isHFA reports a Homogeneous Floating-point Aggregate (HFA) which is a Fundamental Data Type that is a
// Floating-Point type and at most four uniquely addressable members (5.9.5.1 in [Arm64 Calling Convention]).
// This type of struct will be placed more compactly than the individual fields.
//
// [Arm64 Calling Convention]: https://github.com/ARM-software/abi-aa/blob/main/sysvabi64/sysvabi64.rst
func isHFA(t reflect.Type) bool {
	// round up struct size to nearest 8 see section B.4
	structSize := roundUpTo8(t.Size())
	if structSize == 0 || t.NumField() > 4 {
		return false
	}

	// Count the total number of fundamental floating-point types
	var floatCount int
	var firstFloatKind reflect.Kind

	var countFloats func(reflect.Type) bool
	countFloats = func(t reflect.Type) bool {
		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			switch field.Type.Kind() {
			case reflect.Float32, reflect.Float64:
				if floatCount == 0 {
					firstFloatKind = field.Type.Kind()
				} else if field.Type.Kind() != firstFloatKind {
					return false // Mixed float types
				}
				floatCount++
				if floatCount > 4 {
					return false // Too many floats
				}
			case reflect.Struct:
				if !countFloats(field.Type) {
					return false
				}
			case reflect.Array:
				if field.Type.Elem().Kind() == reflect.Float32 || field.Type.Elem().Kind() == reflect.Float64 {
					if floatCount == 0 {
						firstFloatKind = field.Type.Elem().Kind()
					} else if field.Type.Elem().Kind() != firstFloatKind {
						return false
					}
					floatCount += field.Type.Len()
					if floatCount > 4 {
						return false
					}
				} else {
					return false // Non-float array
				}
			default:
				return false // Non-float field
			}
		}
		return true
	}

	if !countFloats(t) {
		return false
	}

	// Must have between 1 and 4 floating-point members
	return floatCount >= 1 && floatCount <= 4
}

// hasDirectFloatFields checks if all fields are directly float32 or float64 (no nested structs)
func hasDirectFloatFields(t reflect.Type) bool {
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		switch field.Type.Kind() {
		case reflect.Float32, reflect.Float64:
			// Direct float field is OK
		case reflect.Array:
			// Float array is OK for HFA
			if field.Type.Elem().Kind() != reflect.Float32 && field.Type.Elem().Kind() != reflect.Float64 {
				return false
			}
		default:
			// Any non-float field (including nested structs) makes this not suitable for simple HFA handling
			return false
		}
	}
	return true
}

// isHVA reports a Homogeneous Aggregate with a Fundamental Data Type that is a Short-Vector type
// and at most four uniquely addressable members (5.9.5.2 in [Arm64 Calling Convention]).
// A short vector is a machine type that is composed of repeated instances of one fundamental integral or
// floating-point type. It may be 8 or 16 bytes in total size (5.4 in [Arm64 Calling Convention]).
// This type of struct will be placed more compactly than the individual fields.
//
// [Arm64 Calling Convention]: https://github.com/ARM-software/abi-aa/blob/main/sysvabi64/sysvabi64.rst
func isHVA(t reflect.Type) bool {
	// round up struct size to nearest 8 see section B.4
	structSize := roundUpTo8(t.Size())
	if structSize == 0 || (structSize != 8 && structSize != 16) {
		return false
	}
	first := t.Field(0)
	switch first.Type.Kind() {
	case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Int8, reflect.Int16, reflect.Int32:
		firstKind := first.Type.Kind()
		for i := 0; i < t.NumField(); i++ {
			if t.Field(i).Type.Kind() != firstKind {
				return false
			}
		}
		return true
	case reflect.Array:
		switch first.Type.Elem().Kind() {
		case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Int8, reflect.Int16, reflect.Int32:
			return true
		default:
			return false
		}
	default:
		return false
	}
}
