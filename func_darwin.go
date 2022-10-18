// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2022 The Ebitengine Authors

package purego

import (
	"reflect"
	"runtime"
	"unsafe"

	"github.com/ebitengine/purego/internal/strings"
)

// RegisterLibFunc takes a pointer to a Go function representing the calling convention of the C function named name
// found in the shared object provided by handle.
// fptr will be set to a function that when called will call the C function given by name with the
// parameters passed in the correct registers and stack.
//
// A panic is produced if the name symbol cannot be found in handle or if the type is not a function
// pointer or if the function returns more than 1 value.
//
// These conversions describe how a Go type in the fptr will be used to call
// the C function. It is important to note that there is no way to verify that fptr
// matches the C function. This also holds true for struct types where the padding
// needs to be ensured to match that of C; RegisterLibFunc does not verify this.
//
// Type Conversions (Go => C)
//
//	string => char*
//	bool => _Bool
//	uintptr => uintptr_t
//	uint => System Dependent
//	uint8 => uint8_t
//	uint16 => uint16_t
//	uint32 => uint32_t
//	uint64 => uint64_t
//	int => System Dependent
//	int8 => int8_t
//	int16 => int16_t
//	int32 => int32_t
//	int64 => int64_t
//	float32 => float
//	float64 => double
//	struct => struct
//	func => C function
//	[]T, unsafe.Pointer, *T => void*
//
// There is a special case when the last argument of fptr is a variadic interface (or []interface}
// it will be expanded into a call to the C function as if it had the arguments in that slice.
func RegisterLibFunc(fptr interface{}, handle uintptr, name string) {
	sym := Dlsym(handle, name)
	if sym == 0 {
		panic("purego: couldn't find symbol: " + Dlerror())
	}
	registerFunc(fptr, sym)
}

// registerFunc takes a C function ptr and a pointer to a Go function which
// will be set to a function calling the C function with those arguments.
func registerFunc(fptr interface{}, cfn uintptr) {
	fn := reflect.ValueOf(fptr).Elem()
	ty := fn.Type()
	if ty.Kind() != reflect.Func {
		panic("purego: fptr must be a function pointer")
	}
	if ty.NumOut() > 1 {
		panic("purego: function can only return zero or one values")
	}
	v := reflect.MakeFunc(ty, func(args []reflect.Value) (results []reflect.Value) {
		if len(args) > 0 {
			if variadic, ok := args[len(args)-1].Interface().([]interface{}); ok {
				// subtract one from args bc the last argument in args is []interface{}
				// which we are currently expanding
				tmp := make([]reflect.Value, len(args)-1+len(variadic))
				n := copy(tmp, args[:len(args)-1])
				for i, v := range variadic {
					tmp[n+i] = reflect.ValueOf(v)
				}
				args = tmp
			}
		}
		var sysargs = make([]uintptr, len(args))
		var keepAlive []interface{}
		defer func() {
			runtime.KeepAlive(keepAlive)
		}()
		for i, v := range args {
			switch v.Kind() {
			case reflect.String:
				ptr := strings.CString(v.String())
				keepAlive = append(keepAlive, ptr)
				sysargs[i] = uintptr(unsafe.Pointer(ptr))
			case reflect.Uintptr, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				sysargs[i] = uintptr(v.Uint())
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				sysargs[i] = uintptr(v.Int())
			case reflect.Ptr, reflect.UnsafePointer, reflect.Slice:
				keepAlive = append(keepAlive, v.Pointer())
				sysargs[i] = v.Pointer()
			case reflect.Func:
				sysargs[i] = NewCallback(v.Interface())
			case reflect.Bool:
				if v.Bool() {
					sysargs[i] = 1
				} else {
					sysargs[i] = 0
				}
			default:
				panic("purego: unsupported kind: " + v.Kind().String())
			}
		}
		r1, _, _ := SyscallN(cfn, sysargs...) //TODO: handle float32/64 and struct types
		if ty.NumOut() == 0 {
			return nil
		}
		outType := ty.Out(0)
		v := reflect.New(outType).Elem()
		switch outType.Kind() {
		case reflect.Uintptr, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			v.SetUint(uint64(r1))
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			v.SetInt(int64(r1))
		case reflect.Bool:
			v.SetBool(r1 != 0)
		case reflect.UnsafePointer:
			// We take the address and then dereference it to trick go vet from creating a possible miss-use of unsafe.Pointer
			v.SetPointer(*(*unsafe.Pointer)(unsafe.Pointer(&r1)))
		case reflect.Ptr:
			v = reflect.NewAt(outType, unsafe.Pointer(&r1)).Elem()
		case reflect.Func:
			// wrap this C function in a nicely typed Go function
			v = reflect.New(outType)
			registerFunc(v.Interface(), r1)
		case reflect.String:
			v.SetString(strings.GoString(r1))
		default:
			panic("purego: unsupported return kind: " + outType.Kind().String())
		}
		return []reflect.Value{v}
	})
	fn.Set(v)
}
