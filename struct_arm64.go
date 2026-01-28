// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024 The Ebitengine Authors

package purego

import (
	"math"
	"reflect"
	"runtime"
	"strconv"
	"sync"
	"unsafe"

	"github.com/ebitengine/purego/internal/strings"
)

// structTypeCache caches reflect.Type for bundled stack arg structs, keyed by
// a string built from the constituent arg types.
var structTypeCache sync.Map // map[string]cachedStructInfo


// structInstancePool pools reflect.Value instances for each cached struct type.
var structInstancePool sync.Map // map[string]*sync.Pool

// fieldNames are pre-computed field name strings to avoid per-call strconv.Itoa.
var fieldNames [maxArgs * 2]string

func init() {
	for i := range fieldNames {
		fieldNames[i] = strconv.Itoa(i)
	}
}

func getStruct(outType reflect.Type, syscall syscall15Args) (v reflect.Value) {
	outSize := outType.Size()
	switch {
	case outSize == 0:
		return reflect.New(outType).Elem()
	case outSize <= 8:
		r1 := syscall.a1
		if isAllFloats, numFields := isAllSameFloat(outType); isAllFloats {
			r1 = syscall.f1
			if numFields == 2 {
				r1 = syscall.f2<<32 | syscall.f1
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
	if runtime.GOOS == "darwin" {
		placeRegistersDarwin(v, addFloat, addInt)
		return
	}
	placeRegistersArm64(v, addFloat, addInt)
}

func placeRegistersArm64(v reflect.Value, addFloat func(uintptr), addInt func(uintptr)) {
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
				val = 0
				class = _NO_CLASS
			}
			switch f.Type().Kind() {
			case reflect.Struct:
				place(f)
			case reflect.Bool:
				if f.Bool() {
					val |= 1 << shift
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
	first := t.Field(0)
	switch first.Type.Kind() {
	case reflect.Float32, reflect.Float64:
		firstKind := first.Type.Kind()
		for i := 0; i < t.NumField(); i++ {
			if t.Field(i).Type.Kind() != firstKind {
				return false
			}
		}
		return true
	case reflect.Array:
		switch first.Type.Elem().Kind() {
		case reflect.Float32, reflect.Float64:
			return true
		default:
			return false
		}
	case reflect.Struct:
		for i := 0; i < first.Type.NumField(); i++ {
			if !isHFA(first.Type) {
				return false
			}
		}
		return true
	default:
		return false
	}
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

// copyStruct8ByteChunks copies struct memory in 8-byte chunks to the provided callback.
// This is used for Darwin ARM64's byte-level packing of non-HFA/HVA structs.
func copyStruct8ByteChunks(ptr unsafe.Pointer, size uintptr, addChunk func(uintptr)) {
	if runtime.GOOS != "darwin" {
		panic("purego: should only be called on darwin")
	}
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
func placeRegistersDarwin(v reflect.Value, addFloat func(uintptr), addInt func(uintptr)) {
	if runtime.GOOS != "darwin" {
		panic("purego: placeRegistersDarwin should only be called on darwin")
	}
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
	if runtime.GOOS != "darwin" {
		return false
	}

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
	if runtime.GOOS != "darwin" {
		panic("purego: structFitsInRegisters should only be called on darwin")
	}
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
// It returns the stack arguments (using the provided buffer to avoid allocation) and processes
// register arguments through addValue.
func collectStackArgs(args []reflect.Value, startIdx int, numInts, numFloats int,
	keepAlive []any, addInt, addFloat, addStack func(uintptr),
	pNumInts, pNumFloats, pNumStack *int, stackBuf []reflect.Value) ([]reflect.Value, []any) {
	if runtime.GOOS != "darwin" {
		panic("purego: collectStackArgs should only be called on darwin")
	}

	stackArgs := stackBuf[:0]
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

// buildStructCacheKey builds a cache key string from the types of stack arguments.
func buildStructCacheKey(stackArgs []reflect.Value) string {
	// Use a simple concatenation of type strings
	var buf []byte
	for i, val := range stackArgs {
		if i > 0 {
			buf = append(buf, ',')
		}
		buf = append(buf, val.Type().String()...)
	}
	return string(buf)
}

// getOrCreateStructInfo returns a cached struct type and field indices for the
// given stack argument types, creating and caching them if necessary.
func getOrCreateStructInfo(stackArgs []reflect.Value, key string) cachedStructInfo {
	if v, ok := structTypeCache.Load(key); ok {
		return v.(cachedStructInfo)
	}

	// Build struct fields with proper C alignment and padding
	var fields []reflect.StructField
	var valueIndices []int
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
				Name: paddingFieldPrefix + fieldNames[fieldIndex],
				Type: reflect.ArrayOf(int(paddingNeeded), reflect.TypeOf(byte(0))),
			})
			currentOffset += paddingNeeded
			fieldIndex++
		}

		fields = append(fields, reflect.StructField{
			Name: "X" + fieldNames[j],
			Type: val.Type(),
		})
		valueIndices = append(valueIndices, fieldIndex)
		currentOffset += valSize
		fieldIndex++
	}

	info := cachedStructInfo{
		typ:          reflect.StructOf(fields),
		valueIndices: valueIndices,
	}
	structTypeCache.Store(key, info)
	return info
}

// getPooledStructInstance returns a pooled struct instance for the given cache key and type.
func getPooledStructInstance(key string, info cachedStructInfo) reflect.Value {
	pool, _ := structInstancePool.LoadOrStore(key, &sync.Pool{
		New: func() any {
			return reflect.New(info.typ)
		},
	})
	return pool.(*sync.Pool).Get().(reflect.Value).Elem()
}

// returnPooledStructInstance returns a struct instance to the pool.
func returnPooledStructInstance(key string, v reflect.Value) {
	if pool, ok := structInstancePool.Load(key); ok {
		pool.(*sync.Pool).Put(v.Addr())
	}
}

// bundleStackArgs bundles remaining arguments for Darwin ARM64 C-style stack packing.
// It creates a packed struct with proper alignment and copies it to the stack in 8-byte chunks.
// When pre-computed bundle info is provided (non-nil), it skips cache key construction and lookup.
func bundleStackArgs(stackArgs []reflect.Value, addStack func(uintptr)) {
	bundleStackArgsWithInfo(stackArgs, addStack, nil)
}

// bundleStackArgsWithInfo is the implementation of bundleStackArgs that optionally accepts
// pre-computed bundle info to skip per-call cache key construction and sync.Map lookups.
func bundleStackArgsWithInfo(stackArgs []reflect.Value, addStack func(uintptr), pre *preBundleInfo) {
	if runtime.GOOS != "darwin" {
		panic("purego: bundleStackArgs should only be called on darwin")
	}
	if len(stackArgs) == 0 {
		return
	}

	var info cachedStructInfo
	var key string
	if pre != nil {
		info = pre.info
		key = pre.key
	} else {
		key = buildStructCacheKey(stackArgs)
		info = getOrCreateStructInfo(stackArgs, key)
	}
	structInstance := getPooledStructInstance(key, info)

	// Set values using pre-computed indices
	for i, idx := range info.valueIndices {
		structInstance.Field(idx).Set(stackArgs[i])
	}

	ptr := unsafe.Pointer(structInstance.Addr().Pointer())
	size := info.typ.Size()
	copyStruct8ByteChunks(ptr, size, addStack)

	returnPooledStructInstance(key, structInstance)
}


// precomputeBundleInfo simulates register assignment for the given function type
// and pre-computes the struct cache key and info for stack arguments.
// Returns nil if no args spill to stack.
func precomputeBundleInfo(ty reflect.Type) *preBundleInfo {
	if runtime.GOOS != "darwin" || runtime.GOARCH != "arm64" {
		return nil
	}

	// Simulate register assignment to find which args spill.
	// String args are converted to *byte by collectStackArgs before bundling,
	// so we must use the post-conversion type here.
	numInts, numFloats := 0, 0
	var spillTypes []reflect.Type
	ptrByteType := reflect.TypeOf((*byte)(nil))
	for i := 0; i < ty.NumIn(); i++ {
		arg := ty.In(i)
		switch arg.Kind() {
		case reflect.Float32, reflect.Float64:
			if numFloats < numOfFloatRegisters {
				numFloats++
			} else {
				spillTypes = append(spillTypes, arg)
			}
		case reflect.Struct:
			// Structs are complex; skip pre-computation for now
			// (they go through structFitsInRegisters at runtime)
			return nil
		case reflect.Slice:
			// Variadic []any — can't pre-compute
			if arg.Elem().Kind() == reflect.Interface {
				return nil
			}
			if numInts < numOfIntegerRegisters() {
				numInts++
			} else {
				spillTypes = append(spillTypes, arg)
			}
		default:
			if numInts < numOfIntegerRegisters() {
				numInts++
			} else {
				// Strings become *byte after CString conversion
				if arg.Kind() == reflect.String {
					spillTypes = append(spillTypes, ptrByteType)
				} else {
					spillTypes = append(spillTypes, arg)
				}
			}
		}
	}

	if len(spillTypes) == 0 {
		return nil
	}

	// Build cache key from spill types
	var buf []byte
	for i, t := range spillTypes {
		if i > 0 {
			buf = append(buf, ',')
		}
		buf = append(buf, t.String()...)
	}
	key := string(buf)

	// Build dummy values to compute struct info
	dummyArgs := make([]reflect.Value, len(spillTypes))
	for i, t := range spillTypes {
		dummyArgs[i] = reflect.New(t).Elem()
	}
	info := getOrCreateStructInfo(dummyArgs, key)

	// Warm the pool
	pool, _ := structInstancePool.LoadOrStore(key, &sync.Pool{
		New: func() any {
			return reflect.New(info.typ)
		},
	})

	return &preBundleInfo{
		key:  key,
		info: info,
		pool: pool.(*sync.Pool),
	}
}
