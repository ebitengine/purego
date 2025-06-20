// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 The Ebitengine Authors

//go:build darwin || freebsd || linux || netbsd

package purego

import (
	"math"
	"reflect"
	"unsafe"
)

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
