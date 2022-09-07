// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2022 The Ebitengine Authors

//go:build !go1.16
// +build !go1.16

package purego

import (
	"reflect"
	"sync"
	"unsafe"
)

// from runtime2.go
// describes how to handle callback
type callbackcontext struct {
	gobody       unsafe.Pointer // go function to call
	argsize      uintptr        // callback arguments size (in bytes)
	restorestack uintptr        // adjust stack on return by (in bytes) (386 only)
	cleanstack   bool
}

type callbacks struct {
	lock sync.Mutex
	ctxt [maxCB]*callbackcontext
	n    int
}

var (
	cbs     callbacks
	cbctxts **callbackcontext = &cbs.ctxt[0] // to simplify access to cbs.ctxt in sys_windows_*.s
)

func compileCallback(fn interface{}) (code uintptr) {
	val := reflect.ValueOf(fn)
	if val.Kind() != reflect.Func {
		panic("purego: type is not a function")
	}
	ty := val.Type()
	var argSize uintptr
	for i := 0; i < ty.NumIn(); i++ {
		in := ty.In(i)
		switch in.Kind() {
		case reflect.Struct, reflect.Float32, reflect.Float64,
			reflect.Interface, reflect.Func, reflect.Slice,
			reflect.Chan, reflect.Complex64, reflect.Complex128,
			reflect.String, reflect.Map, reflect.Invalid:
			panic("purego: unsupported argument type: " + in.Kind().String())
		}
		argSize += ptrSize
	}
	if ty.NumOut() > 1 || ty.NumOut() == 1 && ty.Out(0).Size() != ptrSize {
		panic("purego: callbacks can only have one pointer-sized return")
	}
	cbs.lock.Lock() // We don't unlock this in a defer because this is used from the system stack.

	n := cbs.n
	if n >= maxCB {
		cbs.lock.Unlock()
		panic("purego: the maximum number of callbacks has been reached")
	}

	c := new(callbackcontext)
	c.gobody = unsafe.Pointer(val.Pointer())
	c.argsize = argSize
	cbs.ctxt[n] = c
	cbs.n++
	r := callbackasmAddr(n)
	cbs.lock.Unlock()
	return r
}
