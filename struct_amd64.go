// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024 The Ebitengine Authors

package purego

import (
	"math"
	"reflect"
)

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

func addStruct(v reflect.Value, numInts, numFloats, numStack *int, addInt, addFloat, addStack func(uintptr), keepAlive []interface{}) []interface{} {
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

func postMerger(t reflect.Type) bool {
	// (c) If the size of the aggregate exceeds two eightbytes and the first eight- byte isn’t SSE or any other
	// eightbyte isn’t SSEUP, the whole argument is passed in memory.
	if t.Kind() != reflect.Struct {
		return false
	}
	if t.Size() <= 2*8 {
		return false
	}
	first := getFirst(t).Kind()
	if first != reflect.Float32 && first != reflect.Float64 {
		return false
	}
	return true
}

func getFirst(t reflect.Type) reflect.Type {
	first := t.Field(0).Type
	if first.Kind() == reflect.Struct {
		return getFirst(first)
	}
	return first
}

func tryPlaceRegister(v reflect.Value, addFloat func(uintptr), addInt func(uintptr)) (ok bool) {
	ok = true
	var val uint64
	var shift byte // # of bits to shift
	var flushed bool
	class := _NO_CLASS
	flush := func() {
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
			case reflect.Int64:
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
			case reflect.Uint64:
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
				flush()
			} else if shift > 64 {
				// Should never happen, but may if we forget to reset shift after flush (or forget to flush),
				// better fall apart here, than corrupt arguments.
				panic("purego: tryPlaceRegisters shift > 64")
			}
		}
	}

	place(v)
	if !flushed {
		flush()
	}
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
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
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
