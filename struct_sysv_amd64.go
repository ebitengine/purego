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
		if isAllFloats(structType) {
			// Return in float register (XMM0)
			if structType.NumField() == 1 && structType.Field(0).Type.Kind() == reflect.Float64 {
				a.result = uintptr(math.Float64bits(returnValue.Field(0).Float()))
			} else if structType.NumField() == 2 && structType.Field(0).Type.Kind() == reflect.Float32 {
				// Two float32s
				f1 := uint32(math.Float32bits(float32(returnValue.Field(0).Float())))
				f2 := uint32(math.Float32bits(float32(returnValue.Field(1).Float())))
				a.result = uintptr(uint64(f2)<<32 | uint64(f1))
			} else {
				a.result = uintptr(math.Float64bits(returnValue.Field(0).Float()))
			}
		} else {
			// Return in integer register (RAX)
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
		// Large struct return (>16 bytes) - use hidden pointer
		if a.structRetPtr == 0 {
			panic("purego: large struct return requested but no struct return pointer provided")
		}

		// Copy the struct to the location specified by the caller
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

	// if greater than 64 bytes, passed by reference
	if structType.Size() > 8*8 {
		var pos int
		if *intsN >= numOfIntegerRegisters() {
			pos = *stack
			*stack++
		} else {
			pos = *intsN + numOfFloatRegisters
		}
		*intsN++

		// The frame contains a pointer to the struct
		// Convert uintptr to unsafe.Pointer - this is safe here because
		// frame[pos] contains a valid pointer from the calling convention
		ptr := *(*unsafe.Pointer)(unsafe.Pointer(&frame[pos]))
		return reflect.NewAt(structType, ptr).Elem()
	}

	// Check if this is an all-floats struct
	if isAllFloats(structType) {
		if structType.Size() <= 8 {
			// Single float register
			var pos int
			if *floatsN >= numOfFloatRegisters {
				pos = *stack
				*stack++
			} else {
				pos = *floatsN
			}
			*floatsN++
			return reflect.NewAt(structType, unsafe.Pointer(&frame[pos])).Elem()
		} else if structType.Size() <= 16 {
			// Two float registers
			var pos1, pos2 int
			if *floatsN+1 >= numOfFloatRegisters {
				// Both on stack
				pos1 = *stack
				pos2 = *stack + 1
				*stack += 2
			} else {
				// Both in registers
				pos1 = *floatsN
				pos2 = *floatsN + 1
			}
			*floatsN += 2
			r1, r2 := frame[pos1], frame[pos2]
			return reflect.NewAt(structType, unsafe.Pointer(&struct{ a, b uintptr }{r1, r2})).Elem()
		}
	}

	// Mixed or integer-only struct
	if structType.Size() <= 8 {
		// Single integer register
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
		// Two integer registers (or register + stack)
		var pos1, pos2 int

		// Check if first 8 bytes are floats
		hasFirstFloat := false
		if structType.NumField() > 0 {
			f1 := structType.Field(0).Type
			if f1.Kind() == reflect.Float64 || (f1.Kind() == reflect.Float32 && structType.NumField() > 1 && structType.Field(1).Type.Kind() == reflect.Float32) {
				if *floatsN >= numOfFloatRegisters {
					pos1 = *stack
					*stack++
				} else {
					pos1 = *floatsN
				}
				*floatsN++
				hasFirstFloat = true
			}
		}

		if !hasFirstFloat {
			if *intsN >= numOfIntegerRegisters() {
				pos1 = *stack
				*stack++
			} else {
				pos1 = *intsN + numOfFloatRegisters
			}
			*intsN++
		}

		// Second 8 bytes
		var i int
		for i = 0; i < structType.NumField(); i++ {
			if structType.Field(i).Offset == 8 {
				break
			}
		}

		hasSecondFloat := false
		if i < structType.NumField() {
			f := structType.Field(i).Type
			if f.Kind() == reflect.Float64 || (f.Kind() == reflect.Float32 && i+1 == structType.NumField()) {
				if *floatsN >= numOfFloatRegisters {
					pos2 = *stack
					*stack++
				} else {
					pos2 = *floatsN
				}
				*floatsN++
				hasSecondFloat = true
			}
		}

		if !hasSecondFloat {
			if hasFirstFloat {
				// Second part is integer, use first integer register
				if *intsN >= numOfIntegerRegisters() {
					pos2 = *stack
					*stack++
				} else {
					pos2 = *intsN + numOfFloatRegisters
				}
			} else {
				// Both are integers
				if *intsN >= numOfIntegerRegisters() {
					pos2 = *stack
					*stack++
				} else {
					pos2 = *intsN + numOfFloatRegisters
				}
			}
			*intsN++
		}

		r1, r2 := frame[pos1], frame[pos2]
		return reflect.NewAt(structType, unsafe.Pointer(&struct{ a, b uintptr }{r1, r2})).Elem()
	}

	// Fallback for complex cases - assume passed by reference
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
