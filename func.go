// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2022 The Ebitengine Authors

//go:build darwin || freebsd || linux || windows

package purego

import (
	"math"
	"reflect"
	"runtime"
	"unsafe"

	"github.com/ebitengine/purego/internal/strings"
)

// RegisterLibFunc is a wrapper around RegisterFunc that uses the C function returned from Dlsym(handle, name).
// It panics if it can't find the name symbol.
func RegisterLibFunc(fptr interface{}, handle uintptr, name string) {
	sym, err := loadSymbol(handle, name)
	if err != nil {
		panic(err)
	}
	RegisterFunc(fptr, sym)
}

func Symbol(handle uintptr, name string) uintptr {
	sym, err := loadSymbol(handle, name)
	if err != nil {
		panic(err)
	}
	return sym
}

// RegisterFunc takes a pointer to a Go function representing the calling convention of the C function.
// fptr will be set to a function that when called will call the C function given by cfn with the
// parameters passed in the correct registers and stack.
//
// A panic is produced if the type is not a function pointer or if the function returns more than 1 value.
//
// These conversions describe how a Go type in the fptr will be used to call
// the C function. It is important to note that there is no way to verify that fptr
// matches the C function. This also holds true for struct types where the padding
// needs to be ensured to match that of C; RegisterFunc does not verify this.
//
// # Type Conversions (Go <=> C)
//
//	string <=> char*
//	bool <=> _Bool
//	uintptr <=> uintptr_t
//	uint <=> uint32_t or uint64_t
//	uint8 <=> uint8_t
//	uint16 <=> uint16_t
//	uint32 <=> uint32_t
//	uint64 <=> uint64_t
//	int <=> int32_t or int64_t
//	int8 <=> int8_t
//	int16 <=> int16_t
//	int32 <=> int32_t
//	int64 <=> int64_t
//	float32 <=> float (WIP)
//	float64 <=> double (WIP)
//	struct <=> struct (WIP)
//	func <=> C function
//	unsafe.Pointer, *T <=> void*
//	[]T => void*
//
// There is a special case when the last argument of fptr is a variadic interface (or []interface}
// it will be expanded into a call to the C function as if it had the arguments in that slice.
// This means that using arg ...interface{} is like a cast to the function with the arguments inside arg.
// This is not the same as C variadic.
//
// # Memory
//
// In general it is not possible for purego to guarantee the lifetimes of objects returned or received from
// calling functions using RegisterFunc. For arguments to a C function it is important that the C function doesn't
// hold onto a reference to Go memory. This is the same as the [Cgo rules].
//
// However, there are some special cases. When passing a string as an argument if the string does not end in a null
// terminated byte (\x00) then the string will be copied into memory maintained by purego. The memory is only valid for
// that specific call. Therefore, if the C code keeps a reference to that string it may become invalid at some
// undefined time. However, if the string does already contain a null-terminated byte then no copy is done.
// It is then the responsibility of the caller to ensure the string stays alive as long as it's needed in C memory.
// This can be done using runtime.KeepAlive or allocating the string in C memory using malloc. When a C function
// returns a null-terminated pointer to char a Go string can be used. Purego will allocate a new string in Go memory
// and copy the data over. This string will be garbage collected whenever Go decides it's no longer referenced.
// This C created string will not be freed by purego. If the pointer to char is not null-terminated or must continue
// to point to C memory (because it's a buffer for example) then use a pointer to byte and then convert that to a slice
// using unsafe.Slice. Doing this means that it becomes the responsibility of the caller to care about the lifetime
// of the pointer
//
// # Example
//
// All functions below call this C function:
//
//	char *foo(char *str);
//
//	// Let purego convert types
//	var foo func(s string) string
//	goString := foo("copied")
//	// Go will garbage collect this string
//
//	// Manually, handle allocations
//	var foo2 func(b string) *byte
//	mustFree := foo2("not copied\x00")
//	defer free(mustFree)
//
// [Cgo rules]: https://pkg.go.dev/cmd/cgo#hdr-Go_references_to_C
func RegisterFunc(fptr interface{}, cfn uintptr) {
	fn := reflect.ValueOf(fptr).Elem()
	ty := fn.Type()
	if ty.Kind() != reflect.Func {
		panic("purego: fptr must be a function pointer")
	}
	if ty.NumOut() > 1 {
		panic("purego: function can only return zero or one values")
	}
	if cfn == 0 {
		panic("purego: cfn is nil")
	}
	{
		// this code checks how many registers and stack this function will use
		// to avoid crashing with too many arguments
		var ints int
		var floats int
		var stack int
		for i := 0; i < ty.NumIn(); i++ {
			arg := ty.In(i)
			switch arg.Kind() {
			case reflect.String, reflect.Uintptr, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
				reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Ptr, reflect.UnsafePointer, reflect.Slice,
				reflect.Func, reflect.Bool:
				if ints < numOfIntegerRegisters() {
					ints++
				} else {
					stack++
				}
			case reflect.Float32, reflect.Float64:
				if floats < numOfFloats {
					floats++
				} else {
					stack++
				}
			default:
				panic("purego: unsupported kind " + arg.Kind().String())
			}
		}
		sizeOfStack := maxArgs - numOfIntegerRegisters()
		if stack > sizeOfStack {
			panic("purego: too many arguments")
		}
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
		var sysargs [maxArgs]uintptr
		stack := sysargs[numOfIntegerRegisters():]
		var floats [numOfFloats]uintptr
		var numInts int
		var numFloats int
		var numStack int
		var addStack, addInt, addFloat func(x uintptr)
		if runtime.GOARCH == "arm64" || runtime.GOOS != "windows" {
			// Windows arm64 uses the same calling convention as macOS and Linux
			addStack = func(x uintptr) {
				stack[numStack] = x
				numStack++
			}
			addInt = func(x uintptr) {
				if numInts >= numOfIntegerRegisters() {
					addStack(x)
				} else {
					sysargs[numInts] = x
					numInts++
				}
			}
			addFloat = func(x uintptr) {
				if numFloats < len(floats) {
					floats[numFloats] = x
					numFloats++
				} else {
					addStack(x)
				}
			}
		} else {
			// On Windows amd64 the arguments are passed in the numbered registered.
			// So the first int is in the first integer register and the first float
			// is in the second floating register if there is already a first int.
			// This is in contrast to how macOS and Linux pass arguments which
			// tries to use as many registers as possible in the calling convention.
			addStack = func(x uintptr) {
				sysargs[numStack] = x
				numStack++
			}
			addInt = addStack
			addFloat = addStack
		}

		var keepAlive []interface{}
		defer func() {
			runtime.KeepAlive(keepAlive)
			runtime.KeepAlive(args)
		}()
		for _, v := range args {
			switch v.Kind() {
			case reflect.String:
				ptr := strings.CString(v.String())
				keepAlive = append(keepAlive, ptr)
				addInt(uintptr(unsafe.Pointer(ptr)))
			case reflect.Uintptr, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				addInt(uintptr(v.Uint()))
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				addInt(uintptr(v.Int()))
			case reflect.Ptr, reflect.UnsafePointer, reflect.Slice:
				// There is no need to keepAlive this pointer separately because it is kept alive in the args variable
				addInt(v.Pointer())
			case reflect.Func:
				addInt(NewCallback(v.Interface()))
			case reflect.Bool:
				if v.Bool() {
					addInt(1)
				} else {
					addInt(0)
				}
			case reflect.Float32:
				addFloat(uintptr(math.Float32bits(float32(v.Float()))))
			case reflect.Float64:
				addFloat(uintptr(math.Float64bits(v.Float())))
			default:
				panic("purego: unsupported kind: " + v.Kind().String())
			}
		}
		// TODO: support structs
		var r1, r2 uintptr
		if runtime.GOARCH == "arm64" || runtime.GOOS != "windows" {
			// Use the normal arm64 calling convention even on Windows
			syscall := syscall9Args{
				cfn,
				sysargs[0], sysargs[1], sysargs[2], sysargs[3], sysargs[4], sysargs[5], sysargs[6], sysargs[7], sysargs[8],
				floats[0], floats[1], floats[2], floats[3], floats[4], floats[5], floats[6], floats[7],
				0, 0, 0,
			}
			runtime_cgocall(syscall9XABI0, unsafe.Pointer(&syscall))
			r1, r2 = syscall.r1, syscall.r2
		} else {
			// This is a fallback for Windows amd64, 386, and arm. Note this may not support floats
			r1, r2, _ = syscall_syscall9X(cfn, sysargs[0], sysargs[1], sysargs[2], sysargs[3], sysargs[4], sysargs[5], sysargs[6], sysargs[7], sysargs[8])
		}
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
			// It is safe to have the address of r1 not escape because it is immediately dereferenced with .Elem()
			v = reflect.NewAt(outType, runtime_noescape(unsafe.Pointer(&r1))).Elem()
		case reflect.Func:
			// wrap this C function in a nicely typed Go function
			v = reflect.New(outType)
			RegisterFunc(v.Interface(), r1)
		case reflect.String:
			v.SetString(strings.GoString(r1))
		case reflect.Float32, reflect.Float64:
			// NOTE: r2 is only the floating return value on 64bit platforms.
			// On 32bit platforms r2 is the upper part of a 64bit return.
			v.SetFloat(math.Float64frombits(uint64(r2)))
		default:
			panic("purego: unsupported return kind: " + outType.Kind().String())
		}
		return []reflect.Value{v}
	})
	fn.Set(v)
}

func numOfIntegerRegisters() int {
	switch runtime.GOARCH {
	case "arm64":
		return 8
	case "amd64":
		return 6
		// TODO: figure out why 386 tests are not working
		/*case "386":
			return 0
		case "arm":
			return 4*/
	default:
		panic("purego: unknown GOARCH (" + runtime.GOARCH + ")")
	}
}

// WIP: Less reflection below

type syscallStack interface {
	SysArgs() []uintptr
	Floats() []uintptr

	addStack(x uintptr)
	addInt(x uintptr)
	addFloat(x uintptr)
}

type syscallStackArm64NoWin [1 + maxArgs + numOfFloats]uintptr

func (ss *syscallStackArm64NoWin) numStack() uintptr {
	return ss[0] & 0b1111
}

func (ss *syscallStackArm64NoWin) numInts() uintptr {
	return (ss[0] >> 4) & 0b1111
}

func (ss *syscallStackArm64NoWin) numFloats() uintptr {
	return (ss[0] >> 8) & 0b1111
}

func (ss *syscallStackArm64NoWin) addStack(x uintptr) {
	n := ss.numStack()
	ss[1+n] = x
	ss[0] = (ss[0] - n) | (n + 1)
}

func (ss *syscallStackArm64NoWin) addInt(x uintptr) {
	n := ss.numInts()
	if int(n) >= numOfIntegerRegisters() {
		ss.addStack(x)
	} else {
		ss[1+n] = x
		ss[0] = (ss[0] - (n << 4)) | ((n + 1) << 4)
	}
}

func (ss *syscallStackArm64NoWin) addFloat(x uintptr) {
	n := ss.numFloats()
	if int(n) < numOfFloats {
		ss[1+maxArgs+n] = x
		ss[0] = (ss[0] - (n << 8)) | ((n + 1) << 8)
	} else {
		ss.addStack(x)
	}
}

func (ss *syscallStackArm64NoWin) SysArgs() []uintptr {
	return ss[1:]
}

func (ss *syscallStackArm64NoWin) Floats() []uintptr {
	return ss[1+maxArgs:]
}

type syscallStackAmd64OrWin syscallStackArm64NoWin

func (ss *syscallStackAmd64OrWin) numStack() uintptr {
	return ss[0] & 0b1111
}

func (ss *syscallStackAmd64OrWin) numInts() uintptr {
	return (ss[0] >> 4) & 0b1111
}

func (ss *syscallStackAmd64OrWin) numFloats() uintptr {
	return (ss[0] >> 8) & 0b1111
}

func (ss *syscallStackAmd64OrWin) addStack(x uintptr) {
	n := ss.numStack()
	ss[1+n] = x
	ss[0] = (ss[0] - n) | (n + 1)
}

func (ss *syscallStackAmd64OrWin) addInt(x uintptr) {
	ss.addStack(x)
}

func (ss *syscallStackAmd64OrWin) addFloat(x uintptr) {
	ss.addStack(x)
}

func (ss *syscallStackAmd64OrWin) SysArgs() []uintptr {
	return ss[1:]
}

func (ss *syscallStackAmd64OrWin) Floats() []uintptr {
	return ss[1+maxArgs:]
}

func newSyscallStack() syscallStack {
	if runtime.GOARCH == "arm64" || runtime.GOOS != "windows" {
		return &syscallStackArm64NoWin{}
	}
	return &syscallStackAmd64OrWin{}
}

func uintToPtr[T ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr](v T) uintptr {
	return uintptr(uint64(v))
}

func intToPtr[T ~int | ~int8 | ~int16 | ~int32 | ~int64](v T) uintptr {
	return uintptr(int64(v))
}

func getAddFunc[T any]() func(syscallStack, T) {
	// TODO: support structs
	var v T
	switch any(v).(type) {
	case int:
		return func(r syscallStack, x T) {
			r.addInt(intToPtr(any(x).(int)))
		}
	case int8:
		return func(r syscallStack, x T) {
			r.addInt(intToPtr(any(x).(int8)))
		}
	case int16:
		return func(r syscallStack, x T) {
			r.addInt(intToPtr(any(x).(int16)))
		}
	case int32:
		return func(r syscallStack, x T) {
			r.addInt(intToPtr(any(x).(int32)))
		}
	case int64:
		return func(r syscallStack, x T) {
			r.addInt(intToPtr(any(x).(int64)))
		}
	case uint:
		return func(r syscallStack, x T) {
			r.addInt(uintToPtr(any(x).(uint)))
		}
	case uint8:
		return func(r syscallStack, x T) {
			r.addInt(uintToPtr(any(x).(uint8)))
		}
	case uint16:
		return func(r syscallStack, x T) {
			r.addInt(uintToPtr(any(x).(uint16)))
		}
	case uint32:
		return func(r syscallStack, x T) {
			r.addInt(uintToPtr(any(x).(uint32)))
		}
	case uint64:
		return func(r syscallStack, x T) {
			r.addInt(uintToPtr(any(x).(uint64)))
		}
	case uintptr:
		return func(r syscallStack, x T) {
			r.addInt(uintToPtr(any(x).(uintptr)))
		}
	case float32:
		return func(r syscallStack, x T) {
			r.addFloat(uintptr(math.Float32bits(any(x).(float32))))
		}
	case float64:
		return func(r syscallStack, x T) {
			r.addFloat(uintptr(math.Float64bits(any(x).(float64))))
		}
	case bool:
		return func(r syscallStack, x T) {
			if any(x).(bool) {
				r.addInt(1)
			} else {
				r.addInt(0)
			}
		}
	case string:
		return func(r syscallStack, x T) {
			ptr := strings.CString(any(x).(string))
			r.addInt(uintptr(unsafe.Pointer(ptr)))
		}

	default:
		return func(r syscallStack, x T) {
			rv := reflect.ValueOf(x)
			switch rv.Kind() {
			case reflect.Ptr, reflect.UnsafePointer, reflect.Slice:
				// There is no need to keepAlive this pointer separately because it is kept alive in the args variable
				r.addInt(rv.Pointer())
			case reflect.Func:
				r.addInt(NewCallback(rv.Interface()))
			default:
				panic("purego: unsupported kind: " + rv.Kind().String())
			}
		}
	}
}

func retInts[T int | int8 | int16 | int32 | int64 | uint | uint8 | uint16 | uint32 | uint64 | uintptr, V any](r1, r2 uintptr) V {
	return any(T(r1)).(V)
}

func retBool[V any](r1, r2 uintptr) V {
	return any(r1 != 0).(V)
}

func getReturnFunc[T any]() func(r1, r2 uintptr) T {
	var v T
	switch any(v).(type) {
	case int:
		return retInts[int, T]
	case int8:
		return retInts[int8, T]
	case int16:
		return retInts[int16, T]
	case int32:
		return retInts[int32, T]
	case int64:
		return retInts[int64, T]
	case uint:
		return retInts[uint, T]
	case uint8:
		return retInts[uint8, T]
	case uint16:
		return retInts[uint16, T]
	case uint32:
		return retInts[uint32, T]
	case uint64:
		return retInts[uint64, T]
	case uintptr:
		return retInts[uintptr, T]
	case float32:
		return func(r1, r2 uintptr) T {
			return any(math.Float32frombits(uint32(r2))).(T)
		}
	case float64:
		return func(r1, r2 uintptr) T {
			return any(math.Float64frombits(uint64(r2))).(T)
		}
	case bool:
		return retBool[T]
	case string:
		return func(r1, r2 uintptr) T {
			return any(strings.GoString(r1)).(T)
		}
	case unsafe.Pointer:
		// We take the address and then dereference it to trick go vet from creating a possible miss-use of unsafe.Pointer
		return func(r1, r2 uintptr) T {
			return any(*(*unsafe.Pointer)(unsafe.Pointer(&r1))).(T)
		}
	// Note: funcs and ptrs handled via reflect
	default:
		// TODO: below
		/*u := reflect.ValueOf(v)
		switch v.Elem().Type().Kind() {
		case reflect.Ptr:
			// It is safe to have the address of r1 not escape because it is immediately dereferenced with .Elem()
			v.Set(reflect.NewAt(v.Type(), runtime_noescape(unsafe.Pointer(&r1))).Elem())
		case reflect.Func:
			// wrap this C function in a nicely typed Go function
			fv := reflect.New(v.Type())
			// Note: cannot use a generic one unfortunately
			RegisterFunc(fv.Interface(), r1)
			v.Set(fv)
		default:
			panic("purego: unsupported return kind: " + v.Type().Kind().String())
		}*/
	}

	return nil
}

func argsCheck(fptr any, cfn uintptr) {
	if cfn == 0 {
		panic("purego: cfn is nil")
	}
	// this code checks how many registers and stack this function will use
	// to avoid crashing with too many arguments
	var ints, floats, stack int

	ty := reflect.ValueOf(fptr).Elem().Type()
	for i := 0; i < ty.NumIn(); i++ {
		arg := ty.In(i)
		switch arg.Kind() {
		case reflect.String, reflect.Uintptr, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
			reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Ptr, reflect.UnsafePointer, reflect.Slice,
			reflect.Func, reflect.Bool:
			if ints < numOfIntegerRegisters() {
				ints++
			} else {
				stack++
			}
		case reflect.Float32, reflect.Float64:
			if floats < numOfFloats {
				floats++
			} else {
				stack++
			}
		default:
			panic("purego: unsupported kind " + arg.Kind().String())
		}
	}
	sizeOfStack := maxArgs - numOfIntegerRegisters()
	if stack > sizeOfStack {
		panic("purego: too many arguments")
	}
}

// Convenience to avoid code repetition in all instances of RegisterFuncI_O
func runtime_call(ss syscallStack, cfn uintptr) (uintptr, uintptr) {
	var r1, r2 uintptr
	sysargs, floats := ss.SysArgs(), ss.Floats()
	if runtime.GOARCH == "arm64" || runtime.GOOS != "windows" {
		// Use the normal arm64 calling convention even on Windows
		syscall := syscall9Args{
			cfn,
			sysargs[0], sysargs[1], sysargs[2], sysargs[3], sysargs[4], sysargs[5], sysargs[6], sysargs[7], sysargs[8],
			floats[0], floats[1], floats[2], floats[3], floats[4], floats[5], floats[6], floats[7],
			0, 0, 0,
		}
		runtime_cgocall(syscall9XABI0, unsafe.Pointer(&syscall))
		r1, r2 = syscall.r1, syscall.r2
	} else {
		// This is a fallback for Windows amd64, 386, and arm. Note this may not support floats
		r1, r2, _ = syscall_syscall9X(cfn, sysargs[0], sysargs[1], sysargs[2], sysargs[3], sysargs[4], sysargs[5], sysargs[6], sysargs[7], sysargs[8])
	}

	return r1, r2
}

// No return value

func RegisterFunc1_0[I0 any](fptr *func(I0), cfn uintptr) {
	// Prevent too many registers and check func address is okay
	argsCheck(fptr, cfn)
	func0 := getAddFunc[I0]()
	// Create new function
	*fptr = func(i0 I0) {
		// Create new syscall stack
		ss := newSyscallStack()
		// Add inputs in registers
		func0(ss, i0)
		// Function call
		runtime_call(ss, cfn)
	}
}

func RegisterFunc2_0[I0, I1 any](fptr *func(I0, I1), cfn uintptr) {
	// Prevent too many registers and check func address is okay
	argsCheck(fptr, cfn)
	func0 := getAddFunc[I0]()
	func1 := getAddFunc[I1]()
	// Create new function
	*fptr = func(i0 I0, i1 I1) {
		// Create new syscall stack
		ss := newSyscallStack()
		// Add inputs in registers
		func0(ss, i0)
		func1(ss, i1)
		// Function call
		runtime_call(ss, cfn)
	}
}

func RegisterFunc3_0[I0, I1, I2 any](fptr *func(I0, I1, I2), cfn uintptr) {
	// Prevent too many registers and check func address is okay
	argsCheck(fptr, cfn)
	func0 := getAddFunc[I0]()
	func1 := getAddFunc[I1]()
	func2 := getAddFunc[I2]()
	// Create new function
	*fptr = func(i0 I0, i1 I1, i2 I2) {
		// Create new syscall stack
		ss := newSyscallStack()
		// Add inputs in registers
		func0(ss, i0)
		func1(ss, i1)
		func2(ss, i2)
		// Function call
		runtime_call(ss, cfn)
	}
}

func RegisterFunc4_0[I0, I1, I2, I3 any](fptr *func(I0, I1, I2, I3), cfn uintptr) {
	// Prevent too many registers and check func address is okay
	argsCheck(fptr, cfn)
	func0 := getAddFunc[I0]()
	func1 := getAddFunc[I1]()
	func2 := getAddFunc[I2]()
	func3 := getAddFunc[I3]()
	// Create new function
	*fptr = func(i0 I0, i1 I1, i2 I2, i3 I3) {
		// Create new syscall stack
		ss := newSyscallStack()
		// Add inputs in registers
		func0(ss, i0)
		func1(ss, i1)
		func2(ss, i2)
		func3(ss, i3)
		// Function call
		runtime_call(ss, cfn)
	}
}

func RegisterFunc5_0[I0, I1, I2, I3, I4 any](fptr *func(I0, I1, I2, I3, I4), cfn uintptr) {
	// Prevent too many registers and check func address is okay
	argsCheck(fptr, cfn)
	func0 := getAddFunc[I0]()
	func1 := getAddFunc[I1]()
	func2 := getAddFunc[I2]()
	func3 := getAddFunc[I3]()
	func4 := getAddFunc[I4]()
	// Create new function
	*fptr = func(i0 I0, i1 I1, i2 I2, i3 I3, i4 I4) {
		// Create new syscall stack
		ss := newSyscallStack()
		// Add inputs in registers
		func0(ss, i0)
		func1(ss, i1)
		func2(ss, i2)
		func3(ss, i3)
		func4(ss, i4)
		// Function call
		runtime_call(ss, cfn)
	}
}

// .. so on

// 1 return value

func RegisterFunc1_1[I0, O any](fptr *func(I0) O, cfn uintptr) {
	// Prevent too many registers and check func address is okay
	argsCheck(fptr, cfn)
	returnFunc := getReturnFunc[O]()
	func0 := getAddFunc[I0]()
	// Create new function
	*fptr = func(i0 I0) O {
		// Create new syscall stack
		ss := newSyscallStack()
		// Add inputs in registers
		func0(ss, i0)
		// Function call
		r1, r2 := runtime_call(ss, cfn)

		return returnFunc(r1, r2)
	}
}

func RegisterFunc2_1[I0, I1, O any](fptr *func(I0, I1) O, cfn uintptr) {
	// Prevent too many registers and check func address is okay
	argsCheck(fptr, cfn)
	returnFunc := getReturnFunc[O]()
	func0 := getAddFunc[I0]()
	func1 := getAddFunc[I1]()
	// Create new function
	*fptr = func(i0 I0, i1 I1) O {
		// Create new syscall stack
		ss := newSyscallStack()
		// Add inputs in registers
		func0(ss, i0)
		func1(ss, i1)
		// Function call
		r1, r2 := runtime_call(ss, cfn)

		return returnFunc(r1, r2)
	}
}

func RegisterFunc3_1[I0, I1, I2, O any](fptr *func(I0, I1, I2) O, cfn uintptr) {
	// Prevent too many registers and check func address is okay
	argsCheck(fptr, cfn)
	returnFunc := getReturnFunc[O]()
	func0 := getAddFunc[I0]()
	func1 := getAddFunc[I1]()
	func2 := getAddFunc[I2]()
	// Create new function
	*fptr = func(i0 I0, i1 I1, i2 I2) O {
		// Create new syscall stack
		ss := newSyscallStack()
		// Add inputs in registers
		func0(ss, i0)
		func1(ss, i1)
		func2(ss, i2)
		// Function call
		r1, r2 := runtime_call(ss, cfn)

		return returnFunc(r1, r2)
	}
}

func RegisterFunc4_1[I0, I1, I2, I3, O any](fptr *func(I0, I1, I2, I3) O, cfn uintptr) {
	// Prevent too many registers and check func address is okay
	argsCheck(fptr, cfn)
	returnFunc := getReturnFunc[O]()
	func0 := getAddFunc[I0]()
	func1 := getAddFunc[I1]()
	func2 := getAddFunc[I2]()
	func3 := getAddFunc[I3]()
	// Create new function
	*fptr = func(i0 I0, i1 I1, i2 I2, i3 I3) O {
		// Create new syscall stack
		ss := newSyscallStack()
		// Add inputs in registers
		func0(ss, i0)
		func1(ss, i1)
		func2(ss, i2)
		func3(ss, i3)
		// Function call
		r1, r2 := runtime_call(ss, cfn)

		return returnFunc(r1, r2)
	}
}

func RegisterFunc5_1[I0, I1, I2, I3, I4, O any](fptr *func(I0, I1, I2, I3, I4) O, cfn uintptr) {
	// Prevent too many registers and check func address is okay
	argsCheck(fptr, cfn)
	returnFunc := getReturnFunc[O]()
	func0 := getAddFunc[I0]()
	func1 := getAddFunc[I1]()
	func2 := getAddFunc[I2]()
	func3 := getAddFunc[I3]()
	func4 := getAddFunc[I4]()
	// Create new function
	*fptr = func(i0 I0, i1 I1, i2 I2, i3 I3, i4 I4) O {
		// Create new syscall stack
		ss := newSyscallStack()
		// Add inputs in registers
		func0(ss, i0)
		func1(ss, i1)
		func2(ss, i2)
		func3(ss, i3)
		func4(ss, i4)
		// Function call
		r1, r2 := runtime_call(ss, cfn)

		return returnFunc(r1, r2)
	}
}

// TODO: missing 6-8

func RegisterFunc9_1[I0, I1, I2, I3, I4, I5, I6, I7, I8, O any](fptr *func(I0, I1, I2, I3, I4, I5, I6, I7, I8) O, cfn uintptr) {
	// Prevent too many registers and check func address is okay
	argsCheck(fptr, cfn)
	returnFunc := getReturnFunc[O]()
	func0 := getAddFunc[I0]()
	func1 := getAddFunc[I1]()
	func2 := getAddFunc[I2]()
	func3 := getAddFunc[I3]()
	func4 := getAddFunc[I4]()
	func5 := getAddFunc[I5]()
	func6 := getAddFunc[I6]()
	func7 := getAddFunc[I7]()
	func8 := getAddFunc[I8]()
	// Create new function
	*fptr = func(i0 I0, i1 I1, i2 I2, i3 I3, i4 I4, i5 I5, i6 I6, i7 I7, i8 I8) O {
		// Create new syscall stack
		ss := newSyscallStack()
		// Add inputs in registers
		func0(ss, i0)
		func1(ss, i1)
		func2(ss, i2)
		func3(ss, i3)
		func4(ss, i4)
		func5(ss, i5)
		func6(ss, i6)
		func7(ss, i7)
		func8(ss, i8)
		// Function call
		r1, r2 := runtime_call(ss, cfn)

		return returnFunc(r1, r2)
	}
}
