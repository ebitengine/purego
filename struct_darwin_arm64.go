// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 The Ebitengine Authors

//go:build darwin && arm64

package purego

import (
	"reflect"
	"strconv"
	"unsafe"

	stdstrings "strings"

	"github.com/ebitengine/purego/internal/strings"
)

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

	// Copy the struct memory in 8-byte chunks
	for offset := uintptr(0); offset < size; offset += 8 {
		// Read 8 bytes (or whatever remains) from the struct
		var chunk uintptr
		remaining := size - offset
		if remaining >= 8 {
			chunk = *(*uintptr)(unsafe.Add(ptr, offset))
		} else {
			// For the last partial chunk, read only the remaining bytes
			bytes := (*[8]byte)(unsafe.Add(ptr, offset))
			for i := uintptr(0); i < remaining; i++ {
				chunk |= uintptr(bytes[i]) << (i * 8)
			}
		}
		addInt(chunk)
	}
}

// shouldBundleStackArgs determines if we need to start C-style packing for
// Darwin ARM64 stack arguments. This happens when registers are exhausted.
func shouldBundleStackArgs(v reflect.Value, numInts, numFloats int) bool {
	// Check primitives first
	isFloat := v.Kind() == reflect.Float32 || v.Kind() == reflect.Float64
	isInt := !isFloat && v.Kind() != reflect.Struct
	primitiveOnStack := (isInt && numInts >= numOfIntegerRegisters()) ||
		(isFloat && numFloats >= numOfFloatRegisters)

	// Check if struct would go on stack
	structOnStack := false
	if v.Kind() == reflect.Struct {
		hfa := isHFA(v.Type())
		hva := isHVA(v.Type())
		size := v.Type().Size()

		if hfa || hva || size <= 16 {
			if hfa && numFloats+v.NumField() > numOfFloatRegisters {
				structOnStack = true
			} else if hva && numInts+v.NumField() > numOfIntegerRegisters() {
				structOnStack = true
			} else if size <= 16 {
				slotsNeeded := int((size + align8ByteMask) / align8ByteSize)
				if numInts+slotsNeeded > numOfIntegerRegisters() {
					structOnStack = true
				}
			}
		}
	}

	return primitiveOnStack || structOnStack
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

	// Copy struct memory to stack in 8-byte chunks
	ptr := unsafe.Pointer(structInstance.Addr().Pointer())
	size := structType.Size()
	for offset := uintptr(0); offset < size; offset += 8 {
		var chunk uintptr
		remaining := size - offset
		if remaining >= 8 {
			chunk = *(*uintptr)(unsafe.Add(ptr, offset))
		} else {
			// Handle partial chunk at the end
			bytes := (*[8]byte)(unsafe.Add(ptr, offset))
			for k := uintptr(0); k < remaining; k++ {
				chunk |= uintptr(bytes[k]) << (k * 8)
			}
		}
		addStack(chunk)
	}
}

// estimateStackBytes estimates stack bytes needed for Darwin ARM64 validation.
// This is a conservative estimate used only for early error detection.
// See: https://developer.apple.com/documentation/xcode/writing-arm64-code-for-apple-platforms
func estimateStackBytes(ty reflect.Type) int {
	numInts, numFloats := 0, 0
	stackBytes := 0

	for i := 0; i < ty.NumIn(); i++ {
		arg := ty.In(i)
		size := int(arg.Size())

		// Handle struct arguments with special register allocation rules
		if arg.Kind() == reflect.Struct {
			hfa := isHFA(arg)
			hva := isHVA(arg)

			if hfa {
				// HFA: check if elements fit in float registers
				fieldsNeeded := arg.NumField()
				if numFloats+fieldsNeeded <= numOfFloatRegisters {
					numFloats += fieldsNeeded
					continue
				}
			} else if hva {
				// HVA: check if elements fit in int registers
				fieldsNeeded := arg.NumField()
				if numInts+fieldsNeeded <= numOfIntegerRegisters() {
					numInts += fieldsNeeded
					continue
				}
			} else if size <= 16 {
				// Non-HFA/HVA small structs use int registers for byte-packing
				slotsNeeded := int((size + align8ByteMask) / align8ByteSize)
				if numInts+slotsNeeded <= numOfIntegerRegisters() {
					numInts += slotsNeeded
					continue
				}
			}
			// Struct doesn't fit in registers, goes to stack
			stackBytes += size
		} else {
			// Handle primitive types
			usesInt := arg.Kind() != reflect.Float32 && arg.Kind() != reflect.Float64
			if usesInt && numInts < numOfIntegerRegisters() {
				numInts++
			} else if !usesInt && numFloats < numOfFloatRegisters {
				numFloats++
			} else {
				// Goes to stack - accumulate total bytes
				stackBytes += size
			}
		}
	}
	// Round total to 8-byte boundary
	if stackBytes > 0 && stackBytes%align8ByteSize != 0 {
		stackBytes = (stackBytes + align8ByteMask) &^ align8ByteMask
	}
	return stackBytes
}
