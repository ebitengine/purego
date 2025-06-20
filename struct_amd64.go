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
		if isAllFloats(outType) {
			if outType.NumField() == 1 {
				// Single float32/float64 - use float register
				return reflect.NewAt(outType, unsafe.Pointer(&struct{ a uintptr }{syscall.f1})).Elem()
			} else if outType.NumField() == 2 {
				// Two float32s: prefer a1 for callbacks, but use f1 for regular functions
				// Conservative approach: only use f1 if a1 appears uninitialized/wrong
				if syscall.a1 == 0 && syscall.f1 != 0 {
					// a1 is zero but f1 has a value - likely regular function call
					return reflect.NewAt(outType, unsafe.Pointer(&struct{ a uintptr }{syscall.f1})).Elem()
				}
				// Default: use a1 (works for callbacks and most cases)
			}
		}
		// Up to 8 bytes returned in RAX (integer case or callback case)
		return reflect.NewAt(outType, unsafe.Pointer(&struct{ a uintptr }{syscall.a1})).Elem()
	case outSize <= 16:
		r1, r2 := syscall.a1, syscall.a2
		if isAllFloats(outType) {
			r1 = syscall.f1
			r2 = syscall.f2
		} else {
			// check first 8 bytes if it's floats
			hasFirstFloat := false
			f1 := outType.Field(0).Type
			if f1.Kind() == reflect.Float64 || f1.Kind() == reflect.Float32 && outType.Field(1).Type.Kind() == reflect.Float32 {
				r1 = syscall.f1
				hasFirstFloat = true
			}

			// find index of the field that starts the second 8 bytes
			var i int
			for i = 0; i < outType.NumField(); i++ {
				if outType.Field(i).Offset == 8 {
					break
				}
			}

			// check last 8 bytes if they are floats
			f1 = outType.Field(i).Type
			if f1.Kind() == reflect.Float64 || f1.Kind() == reflect.Float32 && i+1 == outType.NumField() {
				r2 = syscall.f1
			} else if hasFirstFloat {
				// if the first field was a float then that means the second integer field
				// comes from the first integer register
				r2 = syscall.a1
			}
		}
		return reflect.NewAt(outType, unsafe.Pointer(&struct{ a, b uintptr }{r1, r2})).Elem()
	default:
		// create struct from the Go pointer created above
		// weird pointer dereference to circumvent go vet
		return reflect.NewAt(outType, *(*unsafe.Pointer)(unsafe.Pointer(&syscall.a1))).Elem()
	}
}

func isAllFloats(ty reflect.Type) bool {
	for i := 0; i < ty.NumField(); i++ {
		f := ty.Field(i)
		switch f.Type.Kind() {
		case reflect.Float64, reflect.Float32:
		default:
			return false
		}
	}
	return true
}

// https://refspecs.linuxbase.org/elf/x86_64-abi-0.99.pdf
// https://gitlab.com/x86-psABIs/x86-64-ABI
// Class determines where the 8 byte value goes.
// Higher value classes win over lower value classes
const (
	_NO_CLASS = 0b0000
	_SSE      = 0b0001
	_X87      = 0b0011 // long double not used in Go
	_INTEGER  = 0b0111
	_MEMORY   = 0b1111
)

func addStruct(v reflect.Value, numInts, numFloats, numStack *int, addInt, addFloat, addStack func(uintptr), keepAlive []any) []any {
	if v.Type().Size() == 0 {
		return keepAlive
	}

	// if greater than 64 bytes place on stack
	if v.Type().Size() > 8*8 {
		placeStack(v, addStack)
		return keepAlive
	}
	var (
		savedNumFloats = *numFloats
		savedNumInts   = *numInts
		savedNumStack  = *numStack
	)
	placeOnStack := postMerger(v.Type()) || !tryPlaceRegister(v, addFloat, addInt)
	if placeOnStack {
		// reset any values placed in registers
		*numFloats = savedNumFloats
		*numInts = savedNumInts
		*numStack = savedNumStack
		placeStack(v, addStack)
	}
	return keepAlive
}

func postMerger(t reflect.Type) (passInMemory bool) {
	// (c) If the size of the aggregate exceeds two eightbytes and the first eight- byte isn’t SSE or any other
	// eightbyte isn’t SSEUP, the whole argument is passed in memory.
	if t.Kind() != reflect.Struct {
		return false
	}
	if t.Size() <= 2*8 {
		return false
	}
	return true // Go does not have an SSE/SSEUP type so this is always true
}

func tryPlaceRegister(v reflect.Value, addFloat func(uintptr), addInt func(uintptr)) (ok bool) {
	ok = true
	var val uint64
	var shift byte // # of bits to shift
	var flushed bool
	class := _NO_CLASS
	flushIfNeeded := func() {
		if flushed {
			return
		}
		flushed = true
		if class == _SSE {
			addFloat(uintptr(val))
		} else {
			addInt(uintptr(val))
		}
		val = 0
		shift = 0
		class = _NO_CLASS
	}
	var place func(v reflect.Value)
	place = func(v reflect.Value) {
		var numFields int
		if v.Kind() == reflect.Struct {
			numFields = v.Type().NumField()
		} else {
			numFields = v.Type().Len()
		}

		for i := 0; i < numFields; i++ {
			flushed = false
			var f reflect.Value
			if v.Kind() == reflect.Struct {
				f = v.Field(i)
			} else {
				f = v.Index(i)
			}
			switch f.Kind() {
			case reflect.Struct:
				place(f)
			case reflect.Bool:
				if f.Bool() {
					val |= 1
				}
				shift += 8
				class |= _INTEGER
			case reflect.Pointer:
				ok = false
				return
			case reflect.Int8:
				val |= uint64(f.Int()&0xFF) << shift
				shift += 8
				class |= _INTEGER
			case reflect.Int16:
				val |= uint64(f.Int()&0xFFFF) << shift
				shift += 16
				class |= _INTEGER
			case reflect.Int32:
				val |= uint64(f.Int()&0xFFFF_FFFF) << shift
				shift += 32
				class |= _INTEGER
			case reflect.Int64, reflect.Int:
				val = uint64(f.Int())
				shift = 64
				class = _INTEGER
			case reflect.Uint8:
				val |= f.Uint() << shift
				shift += 8
				class |= _INTEGER
			case reflect.Uint16:
				val |= f.Uint() << shift
				shift += 16
				class |= _INTEGER
			case reflect.Uint32:
				val |= f.Uint() << shift
				shift += 32
				class |= _INTEGER
			case reflect.Uint64, reflect.Uint, reflect.Uintptr:
				val = f.Uint()
				shift = 64
				class = _INTEGER
			case reflect.Float32:
				val |= uint64(math.Float32bits(float32(f.Float()))) << shift
				shift += 32
				class |= _SSE
			case reflect.Float64:
				if v.Type().Size() > 16 {
					ok = false
					return
				}
				val = uint64(math.Float64bits(f.Float()))
				shift = 64
				class = _SSE
			case reflect.Array:
				place(f)
			default:
				panic("purego: unsupported kind " + f.Kind().String())
			}

			if shift == 64 {
				flushIfNeeded()
			} else if shift > 64 {
				// Should never happen, but may if we forget to reset shift after flush (or forget to flush),
				// better fall apart here, than corrupt arguments.
				panic("purego: tryPlaceRegisters shift > 64")
			}
		}
	}

	place(v)
	flushIfNeeded()
	return ok
}

func placeStack(v reflect.Value, addStack func(uintptr)) {
	for i := 0; i < v.Type().NumField(); i++ {
		f := v.Field(i)
		switch f.Kind() {
		case reflect.Pointer:
			addStack(f.Pointer())
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			addStack(uintptr(f.Int()))
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
			addStack(uintptr(f.Uint()))
		case reflect.Float32:
			addStack(uintptr(math.Float32bits(float32(f.Float()))))
		case reflect.Float64:
			addStack(uintptr(math.Float64bits(f.Float())))
		case reflect.Struct:
			placeStack(f, addStack)
		default:
			panic("purego: unsupported kind " + f.Kind().String())
		}
	}
}

func placeRegisters(v reflect.Value, addFloat func(uintptr), addInt func(uintptr)) {
	panic("purego: not needed on amd64")
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
		ptr := unsafe.Pointer(frame[pos])
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

	ptr := unsafe.Pointer(frame[pos])
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
		dst := reflect.NewAt(structType, unsafe.Pointer(a.structRetPtr)).Elem()
		dst.Set(returnValue)

		// No value returned in registers for large structs
		a.result = 0
	}
}
