// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 The Ebitengine Authors

package purego

import (
	"reflect"
	"unsafe"
)

// loongLeaf is a scalar member of an aggregate after flattening nested structs
// and arrays. offset is the byte offset of the member within the aggregate.
type loongLeaf struct {
	isFloat bool
	kind    reflect.Kind
	offset  uintptr
	size    uintptr
}

// loongFlatten appends the scalar leaves of t to leaves, offsetting each member
// by base. Nested structs and arrays are expanded into their members.
func loongFlatten(t reflect.Type, base uintptr, leaves *[]loongLeaf) {
	switch t.Kind() {
	case reflect.Struct:
		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)
			loongFlatten(f.Type, base+f.Offset, leaves)
		}
	case reflect.Array:
		elem := t.Elem()
		for i := 0; i < t.Len(); i++ {
			loongFlatten(elem, base+uintptr(i)*elem.Size(), leaves)
		}
	default:
		k := t.Kind()
		*leaves = append(*leaves, loongLeaf{
			isFloat: k == reflect.Float32 || k == reflect.Float64,
			kind:    k,
			offset:  base,
			size:    t.Size(),
		})
	}
}

// loongClassify flattens t and reports whether it is passed and returned through
// the floating-point calling convention. Under the LoongArch hard-float ABI an
// aggregate uses FP registers only when, after flattening, it has one or two
// floating-point members and nothing else, or exactly one floating-point member
// together with one integer member.
func loongClassify(t reflect.Type) (leaves []loongLeaf, useFP bool) {
	loongFlatten(t, 0, &leaves)
	var floats, ints int
	for _, l := range leaves {
		if l.isFloat {
			floats++
		} else {
			ints++
		}
	}
	switch {
	case ints == 0 && (floats == 1 || floats == 2):
		return leaves, true
	case ints == 1 && floats == 1:
		return leaves, true
	default:
		return leaves, false
	}
}

func getStruct(outType reflect.Type, syscall syscallArgs) reflect.Value {
	outSize := outType.Size()
	if outSize == 0 {
		return reflect.New(outType).Elem()
	}
	if outSize > 16 {
		// Returned indirectly through a pointer in the first integer register.
		return reflect.NewAt(outType, *(*unsafe.Pointer)(unsafe.Pointer(&syscall.a1))).Elem()
	}

	var buf [16]byte
	base := unsafe.Pointer(&buf[0])
	if leaves, useFP := loongClassify(outType); useFP {
		floatRegs := [2]uintptr{syscall.f1, syscall.f2}
		intRegs := [2]uintptr{syscall.a1, syscall.a2}
		var fi, ii int
		for _, l := range leaves {
			dst := unsafe.Add(base, l.offset)
			if l.isFloat {
				r := floatRegs[fi]
				fi++
				if l.kind == reflect.Float32 {
					// A single-precision value is NaN-boxed in the register.
					*(*uint32)(dst) = uint32(r)
				} else {
					*(*uint64)(dst) = uint64(r)
				}
			} else {
				loongStoreInt(dst, l.size, intRegs[ii])
				ii++
			}
		}
	} else {
		*(*uintptr)(base) = syscall.a1
		if outSize > 8 {
			*(*uintptr)(unsafe.Add(base, 8)) = syscall.a2
		}
	}
	return reflect.NewAt(outType, base).Elem()
}

func loongStoreInt(dst unsafe.Pointer, size uintptr, r uintptr) {
	switch size {
	case 1:
		*(*uint8)(dst) = uint8(r)
	case 2:
		*(*uint16)(dst) = uint16(r)
	case 4:
		*(*uint32)(dst) = uint32(r)
	default:
		*(*uint64)(dst) = uint64(r)
	}
}

func addStruct(v reflect.Value, numInts, numFloats, numStack *int, addInt, addFloat, addStack func(uintptr), keepAlive []any) []any {
	size := v.Type().Size()
	if size == 0 {
		return keepAlive
	}
	if size > 16 {
		return placeStack(v, keepAlive, addInt)
	}

	var ptr unsafe.Pointer
	if v.CanAddr() {
		ptr = v.Addr().UnsafePointer()
	} else {
		tmp := reflect.New(v.Type())
		tmp.Elem().Set(v)
		ptr = tmp.UnsafePointer()
		keepAlive = append(keepAlive, tmp.Interface())
	}

	if leaves, useFP := loongClassify(v.Type()); useFP {
		for _, l := range leaves {
			src := unsafe.Add(ptr, l.offset)
			switch {
			case l.isFloat && l.kind == reflect.Float32:
				// NaN-box the single-precision value in the 64-bit FP register.
				addFloat(uintptr(*(*uint32)(src)) | 0xFFFFFFFF_00000000)
			case l.isFloat:
				addFloat(uintptr(*(*uint64)(src)))
			default:
				addInt(loongLoadInt(src, l))
			}
		}
		return keepAlive
	}

	// Integer calling convention: pass the raw aggregate in one or two GARs.
	var words [16]byte
	copy(words[:], unsafe.Slice((*byte)(ptr), size))
	addInt(*(*uintptr)(unsafe.Pointer(&words[0])))
	if size > 8 {
		addInt(*(*uintptr)(unsafe.Pointer(&words[8])))
	}
	return keepAlive
}

// loongLoadInt reads an integer leaf into a register value, sign-extending
// signed members and zero-extending the rest.
func loongLoadInt(src unsafe.Pointer, l loongLeaf) uintptr {
	switch l.kind {
	case reflect.Int8:
		return uintptr(int64(*(*int8)(src)))
	case reflect.Int16:
		return uintptr(int64(*(*int16)(src)))
	case reflect.Int32:
		return uintptr(int64(*(*int32)(src)))
	case reflect.Int, reflect.Int64:
		return uintptr(*(*int64)(src))
	case reflect.Bool, reflect.Uint8:
		return uintptr(*(*uint8)(src))
	case reflect.Uint16:
		return uintptr(*(*uint16)(src))
	case reflect.Uint32:
		return uintptr(*(*uint32)(src))
	default:
		return uintptr(*(*uint64)(src))
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

// shouldBundleStackArgs always returns false on loong64
// since C-style stack argument bundling is only needed on Darwin ARM64.
func shouldBundleStackArgs(v reflect.Value, numInts, numFloats int) bool {
	return false
}

// structFitsInRegisters is not used on loong64.
func structFitsInRegisters(val reflect.Value, tempNumInts, tempNumFloats int) (bool, int, int) {
	panic("purego: structFitsInRegisters should not be called on loong64")
}

// collectStackArgs is not used on loong64.
func collectStackArgs(args []reflect.Value, startIdx int, numInts, numFloats int,
	keepAlive []any, addInt, addFloat, addStack func(uintptr),
	pNumInts, pNumFloats, pNumStack *int) ([]reflect.Value, []any) {
	panic("purego: collectStackArgs should not be called on loong64")
}

// bundleStackArgs is not used on loong64.
func bundleStackArgs(stackArgs []reflect.Value, addStack func(uintptr)) {
	panic("purego: bundleStackArgs should not be called on loong64")
}

func getCallbackStruct(inType reflect.Type, frame unsafe.Pointer, floatsN *int, intsN *int, stackSlot *int, stackByteOffset *uintptr) reflect.Value {
	panic("purego: struct callback arguments are not supported on loong64")
}

func setStruct(a *callbackArgs, ret reflect.Value) {
	panic("purego: struct returns are not supported on loong64")
}
