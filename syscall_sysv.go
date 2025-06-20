// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2022 The Ebitengine Authors

//go:build darwin || freebsd || (linux && (amd64 || arm64)) || netbsd

package purego

import (
	"math"
	"reflect"
	"runtime"
	"sync"
	"unsafe"
)

var syscall15XABI0 uintptr

//go:nosplit
func syscall_syscall15X(fn, a1, a2, a3, a4, a5, a6, a7, a8, a9, a10, a11, a12, a13, a14, a15 uintptr) (r1, r2, err uintptr) {
	args := syscall15Args{
		fn, a1, a2, a3, a4, a5, a6, a7, a8, a9, a10, a11, a12, a13, a14, a15,
		a1, a2, a3, a4, a5, a6, a7, a8,
		0,
	}
	runtime_cgocall(syscall15XABI0, unsafe.Pointer(&args))
	return args.a1, args.a2, 0
}

// NewCallback converts a Go function to a function pointer conforming to the C calling convention.
// This is useful when interoperating with C code requiring callbacks. The argument is expected to be a
// function with zero or one uintptr-sized result. The function must not have arguments with size larger than the size
// of uintptr. Only a limited number of callbacks may be created in a single Go process, and any memory allocated
// for these callbacks is never released. At least 2000 callbacks can always be created. Although this function
// provides similar functionality to windows.NewCallback it is distinct.
func NewCallback(fn any) uintptr {
	ty := reflect.TypeOf(fn)
	for i := 0; i < ty.NumIn(); i++ {
		in := ty.In(i)
		if !in.AssignableTo(reflect.TypeOf(CDecl{})) {
			continue
		}
		if i != 0 {
			panic("purego: CDecl must be the first argument")
		}
	}
	return compileCallback(fn)
}

// maxCb is the maximum number of callbacks
// only increase this if you have added more to the callbackasm function
const maxCB = 2000

var cbs struct {
	lock  sync.Mutex
	numFn int                  // the number of functions currently in cbs.funcs
	funcs [maxCB]reflect.Value // the saved callbacks
	// largeStructRet tracks whether each callback returns a struct > 16 bytes
	// This is needed by assembly to handle the hidden pointer parameter
	largeStructRet [maxCB]bool
}

type callbackArgs struct {
	index uintptr
	// args points to the argument block.
	//
	// The structure of the arguments goes
	// float registers followed by the
	// integer registers followed by the stack.
	//
	// This variable is treated as a continuous
	// block of memory containing all of the arguments
	// for this callback.
	args unsafe.Pointer
	// Below are out-args from callbackWrap
	result uintptr
	// structRetPtr is the pointer where large struct returns should be written
	// AMD64: hidden first parameter for structs > 16 bytes
	// ARM64: passed in X8 register for structs > 16 bytes
	structRetPtr uintptr
}

func compileCallback(fn any) uintptr {
	val := reflect.ValueOf(fn)
	if val.Kind() != reflect.Func {
		panic("purego: the type must be a function but was not")
	}
	if val.IsNil() {
		panic("purego: function must not be nil")
	}
	ty := val.Type()
	for i := 0; i < ty.NumIn(); i++ {
		in := ty.In(i)
		switch in.Kind() {
		case reflect.Struct:
			if i == 0 && in.AssignableTo(reflect.TypeOf(CDecl{})) {
				continue
			}
			// Allow structs as callback arguments
			continue
		case reflect.Interface, reflect.Func, reflect.Slice,
			reflect.Chan, reflect.Complex64, reflect.Complex128,
			reflect.String, reflect.Map, reflect.Invalid:
			panic("purego: unsupported argument type: " + in.Kind().String())
		}
	}
output:
	switch {
	case ty.NumOut() == 1:
		switch ty.Out(0).Kind() {
		case reflect.Pointer, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr,
			reflect.Bool, reflect.UnsafePointer, reflect.Float32, reflect.Float64, reflect.Struct:
			break output
		}
		panic("purego: unsupported return type: " + ty.String())
	case ty.NumOut() > 1:
		panic("purego: callbacks can only have one return")
	}
	cbs.lock.Lock()
	defer cbs.lock.Unlock()
	if cbs.numFn >= maxCB {
		panic("purego: the maximum number of callbacks has been reached")
	}
	cbs.funcs[cbs.numFn] = val

	// Check if this callback returns a large struct
	if ty.NumOut() == 1 && ty.Out(0).Kind() == reflect.Struct && ty.Out(0).Size() > 16 {
		cbs.largeStructRet[cbs.numFn] = true
	}

	cbs.numFn++
	return callbackasmAddr(cbs.numFn - 1)
}

const ptrSize = unsafe.Sizeof((*int)(nil))

const callbackMaxFrame = 64 * ptrSize

// callbackasm is implemented in zcallback_GOOS_GOARCH.s
//
//go:linkname __callbackasm callbackasm
var __callbackasm byte
var callbackasmABI0 = uintptr(unsafe.Pointer(&__callbackasm))

// callbackWrap_call allows the calling of the ABIInternal wrapper
// which is required for runtime.cgocallback without the
// <ABIInternal> tag which is only allowed in the runtime.
// This closure is used inside sys_darwin_GOARCH.s
var callbackWrap_call = callbackWrap

// callbackWrap is called by assembly code which determines which Go function to call.
// This function takes the arguments and passes them to the Go function and returns the result.
func callbackWrap(a *callbackArgs) {
	cbs.lock.Lock()
	fn := cbs.funcs[a.index]
	cbs.lock.Unlock()
	fnType := fn.Type()
	args := make([]reflect.Value, fnType.NumIn())
	frame := (*[callbackMaxFrame]uintptr)(a.args)
	var floatsN int // floatsN represents the number of float arguments processed
	var intsN int   // intsN represents the number of integer arguments processed
	// stack points to the index into frame of the current stack element.
	// The stack begins after the float and integer registers.
	stack := numOfIntegerRegisters() + numOfFloatRegisters
	for i := range args {
		var pos int
		switch fnType.In(i).Kind() {
		case reflect.Float32, reflect.Float64:
			if floatsN >= numOfFloatRegisters {
				pos = stack
				stack++
			} else {
				pos = floatsN
			}
			floatsN++
		case reflect.Struct:
			if fnType.In(i).AssignableTo(reflect.TypeOf(CDecl{})) {
				// This is the CDecl field
				args[i] = reflect.Zero(fnType.In(i))
				continue
			}
			// Parse actual struct argument
			args[i] = parseCallbackStruct(fnType.In(i), frame, &floatsN, &intsN, &stack)
			continue
		default:

			if intsN >= numOfIntegerRegisters() {
				pos = stack
				stack++
			} else {
				// the integers begin after the floats in frame
				pos = intsN + numOfFloatRegisters
			}
			intsN++
		}
		args[i] = reflect.NewAt(fnType.In(i), unsafe.Pointer(&frame[pos])).Elem()
	}
	ret := fn.Call(args)
	if len(ret) > 0 {
		switch k := ret[0].Kind(); k {
		case reflect.Uint, reflect.Uint64, reflect.Uint32, reflect.Uint16, reflect.Uint8, reflect.Uintptr:
			a.result = uintptr(ret[0].Uint())
		case reflect.Int, reflect.Int64, reflect.Int32, reflect.Int16, reflect.Int8:
			a.result = uintptr(ret[0].Int())
		case reflect.Bool:
			if ret[0].Bool() {
				a.result = 1
			} else {
				a.result = 0
			}
		case reflect.Pointer:
			a.result = ret[0].Pointer()
		case reflect.UnsafePointer:
			a.result = ret[0].Pointer()
		case reflect.Float32:
			a.result = uintptr(math.Float32bits(float32(ret[0].Float())))
		case reflect.Float64:
			a.result = uintptr(math.Float64bits(ret[0].Float()))
		case reflect.Struct:
			handleStructReturn(ret[0], a)
		default:
			panic("purego: unsupported kind: " + k.String())
		}
	}
}

// callbackasmAddr returns address of runtime.callbackasm
// function adjusted by i.
// On x86 and amd64, runtime.callbackasm is a series of CALL instructions,
// and we want callback to arrive at
// correspondent call instruction instead of start of
// runtime.callbackasm.
// On ARM, runtime.callbackasm is a series of mov and branch instructions.
// R12 is loaded with the callback index. Each entry is two instructions,
// hence 8 bytes.
func callbackasmAddr(i int) uintptr {
	var entrySize int
	switch runtime.GOARCH {
	default:
		panic("purego: unsupported architecture")
	case "386", "amd64":
		entrySize = 5
	case "arm", "arm64":
		// On ARM and ARM64, each entry is a MOV instruction
		// followed by a branch instruction
		entrySize = 8
	}
	return callbackasmABI0 + uintptr(i*entrySize)
}
