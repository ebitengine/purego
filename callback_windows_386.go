// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 The Ebitengine Authors

//go:build windows && 386

package purego

import (
	"math"
	"reflect"
	"sync"
	"syscall"
	"unsafe"
)

const (
	win386CallbackResultInteger = iota
	win386CallbackResultFloat32
	win386CallbackResultFloat64
	win386MaxCallbacks = 2000
)

// win386CallbackResult is filled by win386CallbackDispatch on the C stack.
// callbackasm1 then moves integer results to EDX:EAX or floating-point results
// to x87 ST(0), as required by the 32-bit Windows ABI.
type win386CallbackResult struct {
	low      uintptr
	high     uintptr
	float64  uint64
	kind     uintptr
	stackPop uintptr
}

type win386Callback struct {
	fn       reflect.Value
	argSlots int
	isCDecl  bool
}

var win386Callbacks struct {
	sync.RWMutex
	funcs [win386MaxCallbacks]win386Callback
	n     int
}

var (
	win386CallbackBridgeOnce sync.Once
	win386CallbackBridge     uintptr
)

// callbackasm is implemented by zcallback_386.s.
//
//go:linkname win386Callbackasm callbackasm
var win386Callbackasm byte

func win386CallbackasmAddr(index int) uintptr {
	// Each entry is MOVL $index, CX followed by JMP callbackasm1.
	return uintptr(unsafe.Pointer(&win386Callbackasm)) + uintptr(index*10)
}

func newCallback(fn any, isCDecl bool) uintptr {
	value := reflect.ValueOf(fn)
	if value.Kind() != reflect.Func {
		panic("purego: the type must be a function")
	}
	if value.IsNil() {
		panic("purego: function must not be nil")
	}

	typeOfFn := value.Type()
	argSlots := 0
	for i := 0; i < typeOfFn.NumIn(); i++ {
		in := typeOfFn.In(i)
		if i == 0 && in.AssignableTo(reflect.TypeFor[CDecl]()) {
			continue
		}
		switch in.Kind() {
		case reflect.Bool,
			reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr,
			reflect.Float32, reflect.Float64, reflect.Pointer, reflect.UnsafePointer:
		default:
			panic("purego: unsupported argument type: " + in.Kind().String())
		}
		argSlots += int((in.Size() + unsafe.Sizeof(uintptr(0)) - 1) / unsafe.Sizeof(uintptr(0)))
	}
	if argSlots > maxArgs {
		panic("purego: too many callback argument slots")
	}

	if typeOfFn.NumOut() > 1 {
		panic("purego: callbacks can only have one return")
	}
	if typeOfFn.NumOut() == 1 {
		switch out := typeOfFn.Out(0); out.Kind() {
		case reflect.Bool,
			reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr,
			reflect.Float32, reflect.Float64, reflect.Pointer, reflect.UnsafePointer:
		default:
			panic("purego: unsupported return type: " + out.String())
		}
	}

	win386CallbackBridgeOnce.Do(func() {
		// The native trampoline always cleans these three private arguments.
		// The original callback's cdecl/stdcall cleanup is handled separately.
		win386CallbackBridge = syscall.NewCallback(win386CallbackDispatch)
	})

	win386Callbacks.Lock()
	defer win386Callbacks.Unlock()
	if win386Callbacks.n >= len(win386Callbacks.funcs) {
		panic("purego: the maximum number of callbacks has been reached")
	}
	index := win386Callbacks.n
	win386Callbacks.funcs[index] = win386Callback{fn: value, argSlots: argSlots, isCDecl: isCDecl}
	win386Callbacks.n++
	return win386CallbackasmAddr(index)
}

// win386CallbackDispatch has only pointer-sized arguments and return value, so
// it can use the standard library's restricted Windows/386 callback adapter.
// The original callback's wider typed ABI is decoded here by purego.
func win386CallbackDispatch(index uintptr, stack unsafe.Pointer, result *win386CallbackResult) uintptr {
	*result = win386CallbackResult{}
	win386Callbacks.RLock()
	callback := win386Callbacks.funcs[index]
	win386Callbacks.RUnlock()

	fnType := callback.fn.Type()
	values := make([]reflect.Value, fnType.NumIn())
	slots := (*[maxArgs]uintptr)(stack)
	stackSlot := 0
	for i := range values {
		in := fnType.In(i)
		if i == 0 && in.AssignableTo(reflect.TypeFor[CDecl]()) {
			values[i] = reflect.Zero(in)
			continue
		}
		values[i] = reflect.NewAt(in, unsafe.Pointer(&slots[stackSlot])).Elem()
		stackSlot += int((in.Size() + unsafe.Sizeof(uintptr(0)) - 1) / unsafe.Sizeof(uintptr(0)))
	}

	if !callback.isCDecl {
		result.stackPop = uintptr(callback.argSlots) * unsafe.Sizeof(uintptr(0))
	}
	returned := callback.fn.Call(values)
	if len(returned) == 0 {
		return 0
	}

	value := returned[0]
	switch value.Kind() {
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		bits := value.Uint()
		result.low = uintptr(bits)
		result.high = uintptr(bits >> 32)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		bits := uint64(value.Int())
		result.low = uintptr(bits)
		result.high = uintptr(bits >> 32)
	case reflect.Bool:
		if value.Bool() {
			result.low = 1
		}
	case reflect.Pointer, reflect.UnsafePointer:
		result.low = value.Pointer()
	case reflect.Float32:
		result.float64 = uint64(math.Float32bits(float32(value.Float())))
		result.kind = win386CallbackResultFloat32
	case reflect.Float64:
		result.float64 = math.Float64bits(value.Float())
		result.kind = win386CallbackResultFloat64
	default:
		panic("purego: unsupported callback return kind: " + value.Kind().String())
	}
	return 0
}
