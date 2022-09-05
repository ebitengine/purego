//go:build !go1.16
// +build !go1.16

package purego

import (
	"reflect"
	"runtime"
	"sync"
	"unsafe"
)

func NewCallback(fn interface{}) uintptr {
	return compileCallback(fn)
}

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
		panic("unsupported architecture")
	case "386", "amd64":
		entrySize = 5
	case "arm":
		// On ARM, each entry is a MOV instruction
		// followed by a branch instruction
		entrySize = 8
	}
	return callbackasmABI0 + uintptr(i*entrySize)
}

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
	/*
		lock(&cbs.lock) // We don't unlock this in a defer because this is used from the system stack.

		n := cbs.n
		for i := 0; i < n; i++ {
			if cbs.ctxt[i].gobody == fn.data && cbs.ctxt[i].isCleanstack() == cleanstack {
				r := callbackasmAddr(i)
				unlock(&cbs.lock)
				return r
			}
		}
		if n >= cb_max {
			unlock(&cbs.lock)
			throw("too many callback functions")
		}

		c := new(callbackcontext)
		c.gobody = fn.data
		c.argsize = argsize
		c.setCleanstack(cleanstack)
		if cleanstack && argsize != 0 {
			c.restorestack = argsize
		} else {
			c.restorestack = 0
		}
		cbs.ctxt[n] = c
		cbs.n++

		r := callbackasmAddr(n)
		unlock(&cbs.lock)
		return r*/
}
