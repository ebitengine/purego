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

// parseCallbackStruct parses a struct argument from callback frame data
func parseCallbackStruct(structType reflect.Type, frame *[callbackMaxFrame]uintptr, floatsN, intsN, stack *int) reflect.Value {
	if structType.Size() == 0 {
		return reflect.New(structType).Elem()
	}

	// Check for HFA/HVA or small structs that fit in registers
	hva := isHVA(structType)
	hfa := hasDirectFloatFields(structType) && isHFA(structType)
	size := structType.Size()

	if hva || hfa || size <= 16 {

		if hfa {
			// Homogeneous Floating-point Aggregate
			numFields := structType.NumField()
			if *floatsN+numFields > numOfFloatRegisters {
				// Not enough float registers, use stack
				result := reflect.New(structType).Elem()
				for i := 0; i < numFields; i++ {
					pos := *stack + i
					field := result.Field(i)
					if structType.Field(i).Type.Kind() == reflect.Float32 {
						field.SetFloat(float64(math.Float32frombits(uint32(frame[pos]))))
					} else {
						field.SetFloat(math.Float64frombits(uint64(frame[pos])))
					}
				}
				*stack += numFields
				return result
			} else {
				// Use float registers
				result := reflect.New(structType).Elem()
				for i := 0; i < numFields; i++ {
					pos := *floatsN + i
					field := result.Field(i)
					fieldType := structType.Field(i).Type
					if fieldType.Kind() == reflect.Float32 {
						field.SetFloat(float64(math.Float32frombits(uint32(frame[pos]))))
					} else if fieldType.Kind() == reflect.Float64 {
						field.SetFloat(math.Float64frombits(uint64(frame[pos])))
					} else {
						// This shouldn't happen in an HFA, but let's handle it gracefully
						panic("purego: non-float field in HFA struct")
					}
				}
				*floatsN += numFields
				return result
			}
		} else if hva {
			// Homogeneous Vector Aggregate
			numFields := structType.NumField()
			if *intsN+numFields > numOfIntegerRegisters() {
				// Not enough integer registers, use stack
				var pos int
				if *stack == numOfIntegerRegisters()+numOfFloatRegisters {
					pos = *stack
				} else {
					pos = *stack
				}
				*stack += (int(structType.Size()) + int(unsafe.Sizeof(uintptr(0))) - 1) / int(unsafe.Sizeof(uintptr(0)))
				return reflect.NewAt(structType, unsafe.Pointer(&frame[pos])).Elem()
			} else {
				// Use integer registers
				return parseStructFromRegisters(structType, frame, floatsN, intsN, stack)
			}
		} else if isAllFloats, numFields := isAllSameFloat(structType); isAllFloats && numFields <= 4 {
			// All same float type struct
			if *floatsN+numFields > numOfFloatRegisters {
				// Use stack
				var pos int
				if numFields == 1 {
					pos = *stack
					*stack++
				} else {
					pos = *stack
					*stack += numFields
				}
				return reflect.NewAt(structType, unsafe.Pointer(&frame[pos])).Elem()
			} else {
				// Use float registers
				switch numFields {
				case 1:
					pos := *floatsN
					*floatsN++
					return reflect.NewAt(structType, unsafe.Pointer(&frame[pos])).Elem()
				case 2:
					pos1, pos2 := *floatsN, *floatsN+1
					*floatsN += 2
					if structType.Field(0).Type.Kind() == reflect.Float32 {
						// Two float32s packed in one register
						val := frame[pos1]<<32 | frame[pos2]
						return reflect.NewAt(structType, unsafe.Pointer(&val)).Elem()
					} else {
						// Two float64s in separate registers
						r1, r2 := frame[pos1], frame[pos2]
						return reflect.NewAt(structType, unsafe.Pointer(&struct{ a, b uintptr }{r1, r2})).Elem()
					}
				case 3, 4:
					result := reflect.New(structType).Elem()
					for i := 0; i < numFields; i++ {
						pos := *floatsN + i
						field := result.Field(i)
						if structType.Field(i).Type.Kind() == reflect.Float32 {
							field.SetFloat(float64(math.Float32frombits(uint32(frame[pos]))))
						} else {
							field.SetFloat(math.Float64frombits(uint64(frame[pos])))
						}
					}
					*floatsN += numFields
					return result
				}
			}
		} else {
			// General small struct (≤ 16 bytes)
			return parseStructFromRegisters(structType, frame, floatsN, intsN, stack)
		}
	} else {
		// Large struct passed by reference
		var pos int
		if *intsN >= numOfIntegerRegisters() {
			pos = *stack
			*stack++
		} else {
			pos = *intsN + numOfFloatRegisters
		}
		*intsN++

		if frame[pos] == 0 {
			// Return zero value for nil pointer
			return reflect.New(structType).Elem()
		}
		// Convert uintptr to unsafe.Pointer - this is safe here because
		// frame[pos] contains a valid pointer from the calling convention
		ptr := *(*unsafe.Pointer)(unsafe.Pointer(&frame[pos]))
		return reflect.NewAt(structType, ptr).Elem()
	}

	// Should not reach here
	panic("purego: unsupported struct layout for callbacks")
}

// parseStructFromRegisters handles general struct parsing for structs that fit in registers
func parseStructFromRegisters(structType reflect.Type, frame *[callbackMaxFrame]uintptr, floatsN, intsN, stack *int) reflect.Value {
	if structType.Size() <= 8 {
		// Single register
		var pos int
		if *intsN >= numOfIntegerRegisters() {
			pos = *stack
			*stack++
		} else {
			pos = *intsN + numOfFloatRegisters
		}
		*intsN++
		return reflect.NewAt(structType, unsafe.Pointer(&frame[pos])).Elem()
	} else if structType.Size() <= 16 {
		// Two registers
		var pos1, pos2 int
		if *intsN+1 >= numOfIntegerRegisters() {
			// Use stack
			pos1 = *stack
			pos2 = *stack + 1
			*stack += 2
		} else {
			// Use registers
			pos1 = *intsN + numOfFloatRegisters
			pos2 = *intsN + 1 + numOfFloatRegisters
		}
		*intsN += 2

		r1, r2 := frame[pos1], frame[pos2]
		return reflect.NewAt(structType, unsafe.Pointer(&struct{ a, b uintptr }{r1, r2})).Elem()
	}

	// Fallback for edge cases
	var pos int
	if *intsN >= numOfIntegerRegisters() {
		pos = *stack
		*stack++
	} else {
		pos = *intsN + numOfFloatRegisters
	}
	*intsN++

	// Convert uintptr to unsafe.Pointer - this is safe here because
	// frame[pos] contains a valid pointer from the calling convention
	ptr := *(*unsafe.Pointer)(unsafe.Pointer(&frame[pos]))
	return reflect.NewAt(structType, ptr).Elem()
}

// handleStructReturn handles struct return values for callbacks
func handleStructReturn(returnValue reflect.Value, a *callbackArgs) {
	structType := returnValue.Type()
	size := structType.Size()

	switch {
	case size == 0:
		// Empty struct, nothing to return
		a.result = 0
	case size <= 8:
		// Single register return
		if hfa := isHFA(structType); hfa {
			// For callback returns, HFA structs must be packed into the integer result register
			// because the callback mechanism only supports one result value
			// Return in float register (D0/S0)
			if structType.NumField() == 1 {
				field := returnValue.Field(0)
				if field.Type().Kind() == reflect.Float64 {
					a.result = uintptr(math.Float64bits(field.Float()))
				} else {
					a.result = uintptr(math.Float32bits(float32(field.Float())))
				}
			} else if structType.NumField() == 2 && structType.Field(0).Type.Kind() == reflect.Float32 {
				// Two float32s: Pack them into a single register for the callback result
				// The receiving side will read this from a1 and interpret it as a packed struct
				f1 := uint32(math.Float32bits(float32(returnValue.Field(0).Float()))) // field 0 (A)
				f2 := uint32(math.Float32bits(float32(returnValue.Field(1).Float()))) // field 1 (B)
				// Pack as memory layout: field0 in low 32 bits, field1 in high 32 bits
				a.result = uintptr(uint64(f2)<<32 | uint64(f1))
			} else {
				a.result = uintptr(math.Float64bits(returnValue.Field(0).Float()))
			}
		} else {
			// Return in integer register (X0)
			// Create an addressable copy since returnValue from fn.Call() is not addressable
			copy := reflect.New(returnValue.Type()).Elem()
			copy.Set(returnValue)
			ptr := unsafe.Pointer(copy.UnsafeAddr())
			a.result = *(*uintptr)(ptr)
		}
	case size <= 16:
		// Two register return would require assembly modifications
		panic("purego: struct return values of 9-16 bytes not yet supported in callbacks (requires assembly changes)")
	default:
		// Large struct return (>16 bytes) - use X8 register pointer
		if a.structRetPtr == 0 {
			panic("purego: large struct return requested but no struct return pointer provided (X8)")
		}

		// Copy the struct to the location specified by the caller (via X8)
		// Convert uintptr to unsafe.Pointer - this is safe here because
		// a.structRetPtr contains a valid pointer from the calling convention
		ptr := *(*unsafe.Pointer)(unsafe.Pointer(&a.structRetPtr))
		dst := reflect.NewAt(structType, ptr).Elem()
		dst.Set(returnValue)

		// No value returned in registers for large structs
		a.result = 0
	}
}
