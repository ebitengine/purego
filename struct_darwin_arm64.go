// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 The Ebitengine Authors

package purego

import (
	"reflect"
	"strconv"
	"unsafe"

	stdstrings "strings"

	"github.com/ebitengine/purego/internal/strings"
)

// copyStruct8ByteChunks copies struct memory in 8-byte chunks to the provided callback.
// This is used for Darwin ARM64's byte-level packing of non-HFA/HVA structs.
func copyStruct8ByteChunks(ptr unsafe.Pointer, size uintptr, addChunk func(uintptr)) {
	for offset := uintptr(0); offset < size; offset += 8 {
		var chunk uintptr
		remaining := size - offset
		if remaining >= 8 {
			chunk = *(*uintptr)(unsafe.Add(ptr, offset))
		} else {
			// Read byte-by-byte to avoid reading beyond allocation
			for i := uintptr(0); i < remaining; i++ {
				b := *(*byte)(unsafe.Add(ptr, offset+i))
				chunk |= uintptr(b) << (i * 8)
			}
		}
		addChunk(chunk)
	}
}

// placeRegisters implements Darwin ARM64 calling convention for struct arguments.
//
// For HFA/HVA structs, each element must go in a separate register (or stack slot for elements
// that don't fit in registers). We use placeRegistersArm64 for this.
//
// For non-HFA/HVA structs, Darwin uses byte-level packing. We copy the struct memory in
// 8-byte chunks, which works correctly for both register and stack placement.
func placeRegisters(v reflect.Value, addFloat func(uintptr), addInt func(uintptr)) {
	// Check if this is an HFA/HVA
	hfa := isHFA(v.Type())
	hva := isHVA(v.Type())

	// For HFA/HVA structs, use the standard ARM64 logic which places each element separately
	if hfa || hva {
		placeRegistersArm64(v, addFloat, addInt)
		return
	}

	// For non-HFA/HVA structs, use byte-level copying
	// If the value is not addressable, create an addressable copy
	if !v.CanAddr() {
		addressable := reflect.New(v.Type()).Elem()
		addressable.Set(v)
		v = addressable
	}
	ptr := unsafe.Pointer(v.Addr().Pointer())
	size := v.Type().Size()
	copyStruct8ByteChunks(ptr, size, addInt)
}

// shouldBundleStackArgs determines if we need to start C-style packing for
// Darwin ARM64 stack arguments. This happens when registers are exhausted.
func shouldBundleStackArgs(v reflect.Value, numInts, numFloats int) bool {
	kind := v.Kind()
	isFloat := kind == reflect.Float32 || kind == reflect.Float64
	isInt := !isFloat && kind != reflect.Struct
	primitiveOnStack :=
		(isInt && numInts >= numOfIntegerRegisters()) ||
			(isFloat && numFloats >= numOfFloatRegisters)
	if primitiveOnStack {
		return true
	}
	if kind != reflect.Struct {
		return false
	}
	hfa := isHFA(v.Type())
	hva := isHVA(v.Type())
	size := v.Type().Size()
	eligible := hfa || hva || size <= 16
	if !eligible {
		return false
	}

	if hfa {
		need := v.NumField()
		return numFloats+need > numOfFloatRegisters
	}

	if hva {
		need := v.NumField()
		return numInts+need > numOfIntegerRegisters()
	}

	slotsNeeded := int((size + align8ByteMask) / align8ByteSize)
	return numInts+slotsNeeded > numOfIntegerRegisters()
}

// structFitsInRegisters determines if a struct can still fit in remaining
// registers, used during stack argument bundling to decide if a struct
// should go through normal register allocation or be bundled with stack args.
func structFitsInRegisters(val reflect.Value, tempNumInts, tempNumFloats int) (bool, int, int) {
	hfa := isHFA(val.Type())
	hva := isHVA(val.Type())
	size := val.Type().Size()

	if hfa {
		// HFA: check if elements fit in float registers
		if tempNumFloats+val.NumField() <= numOfFloatRegisters {
			return true, tempNumInts, tempNumFloats + val.NumField()
		}
	} else if hva {
		// HVA: check if elements fit in int registers
		if tempNumInts+val.NumField() <= numOfIntegerRegisters() {
			return true, tempNumInts + val.NumField(), tempNumFloats
		}
	} else if size <= 16 {
		// Non-HFA/HVA small structs use int registers for byte-packing
		slotsNeeded := int((size + align8ByteMask) / align8ByteSize)
		if tempNumInts+slotsNeeded <= numOfIntegerRegisters() {
			return true, tempNumInts + slotsNeeded, tempNumFloats
		}
	}

	return false, tempNumInts, tempNumFloats
}

// collectStackArgs separates remaining arguments into those that fit in registers vs those that go on stack.
// It returns the stack arguments and processes register arguments through addValue.
func collectStackArgs(args []reflect.Value, startIdx int, numInts, numFloats int,
	keepAlive []any, addInt, addFloat, addStack func(uintptr),
	pNumInts, pNumFloats, pNumStack *int) ([]reflect.Value, []any) {

	var stackArgs []reflect.Value
	tempNumInts := numInts
	tempNumFloats := numFloats

	for j, val := range args[startIdx:] {
		// Determine if this argument goes to register or stack
		var fitsInRegister bool
		var newNumInts, newNumFloats int

		if val.Kind() == reflect.Struct {
			// Check if struct still fits in remaining registers
			fitsInRegister, newNumInts, newNumFloats = structFitsInRegisters(val, tempNumInts, tempNumFloats)
		} else {
			// Primitive argument
			isFloat := val.Kind() == reflect.Float32 || val.Kind() == reflect.Float64
			if isFloat {
				fitsInRegister = tempNumFloats < numOfFloatRegisters
				newNumFloats = tempNumFloats + 1
				newNumInts = tempNumInts
			} else {
				fitsInRegister = tempNumInts < numOfIntegerRegisters()
				newNumInts = tempNumInts + 1
				newNumFloats = tempNumFloats
			}
		}

		if fitsInRegister {
			// Process through normal register allocation
			tempNumInts = newNumInts
			tempNumFloats = newNumFloats
			keepAlive = addValue(val, keepAlive, addInt, addFloat, addStack, pNumInts, pNumFloats, pNumStack)
		} else {
			// Convert strings to C strings before bundling
			if val.Kind() == reflect.String {
				ptr := strings.CString(val.String())
				keepAlive = append(keepAlive, ptr)
				val = reflect.ValueOf(ptr)
				args[startIdx+j] = val
			}
			stackArgs = append(stackArgs, val)
		}
	}

	return stackArgs, keepAlive
}

const (
	paddingFieldPrefix = "Pad"
)

// bundleStackArgs bundles remaining arguments for Darwin ARM64 C-style stack packing.
// It creates a packed struct with proper alignment and copies it to the stack in 8-byte chunks.
func bundleStackArgs(stackArgs []reflect.Value, addStack func(uintptr)) {
	if len(stackArgs) == 0 {
		return
	}

	// Build struct fields with proper C alignment and padding
	var fields []reflect.StructField
	currentOffset := uintptr(0)
	fieldIndex := 0

	for j, val := range stackArgs {
		valSize := val.Type().Size()
		valAlign := val.Type().Align()

		// ARM64 requires 8-byte alignment for 8-byte or larger structs
		if val.Kind() == reflect.Struct && valSize >= 8 {
			valAlign = 8
		}

		// Add padding field if needed for alignment
		if currentOffset%uintptr(valAlign) != 0 {
			paddingNeeded := uintptr(valAlign) - (currentOffset % uintptr(valAlign))
			fields = append(fields, reflect.StructField{
				Name: paddingFieldPrefix + strconv.Itoa(fieldIndex),
				Type: reflect.ArrayOf(int(paddingNeeded), reflect.TypeOf(byte(0))),
			})
			currentOffset += paddingNeeded
			fieldIndex++
		}

		fields = append(fields, reflect.StructField{
			Name: "X" + strconv.Itoa(j),
			Type: val.Type(),
		})
		currentOffset += valSize
		fieldIndex++
	}

	// Create and populate the packed struct
	structType := reflect.StructOf(fields)
	structInstance := reflect.New(structType).Elem()

	// Set values (skip padding fields)
	argIndex := 0
	for j := 0; j < structInstance.NumField(); j++ {
		fieldName := structType.Field(j).Name
		if stdstrings.HasPrefix(fieldName, paddingFieldPrefix) {
			continue
		}
		structInstance.Field(j).Set(stackArgs[argIndex])
		argIndex++
	}

	ptr := unsafe.Pointer(structInstance.Addr().Pointer())
	size := structType.Size()
	copyStruct8ByteChunks(ptr, size, addStack)
}
