// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build !cgo && amd64

package fakecgo

import "unsafe"

//go:nosplit
func _cgo_sys_thread_start(ts *ThreadStart) {
	var attr pthread_attr_t
	var ign, oset sigset_t
	var p pthread_t
	var size size_t
	var err int

	sigfillset(&ign)
	pthread_sigmask(SIG_SETMASK, &ign, &oset)

	pthread_attr_init(&attr)
	pthread_attr_getstacksize(&attr, &size)
	// illumos reports 0 for the default stack size, fix it.
	if size == 0 {
		size = 2 << 20
		if pthread_attr_setstacksize(&attr, size) != 0 {
			println("fakecgo: pthread_attr_setstacksize failed")
		}
	}
	// Leave stacklo=0 and set stackhi=size; mstart will do the rest.
	ts.g.stackhi = uintptr(size)

	err = _cgo_try_pthread_create(&p, &attr, unsafe.Pointer(threadentry_trampolineABI0), ts)

	// pthread_attr_t holds a pointer to memory allocated by pthread_attr_init.
	pthread_attr_destroy(&attr)

	pthread_sigmask(SIG_SETMASK, &oset, nil)

	if err != 0 {
		print("fakecgo: pthread_create failed: ")
		println(err)
		abort()
	}
}

// threadentry_trampolineABI0 maps the C ABI to Go ABI then calls the Go function
//
//go:linkname x_threadentry_trampoline threadentry_trampoline
var x_threadentry_trampoline byte
var threadentry_trampolineABI0 = &x_threadentry_trampoline

//go:nosplit
func threadentry(v unsafe.Pointer) unsafe.Pointer {
	ts := *(*ThreadStart)(v)
	free(v)

	setg_trampoline(setg_func, uintptr(unsafe.Pointer(ts.g)))

	callThreadEntryFn(ts.fn)

	return nil
}

// setg_func is the setg_gcc function passed to x_cgo_init.
var setg_func uintptr

// x_cgo_init(G *g, void (*setg)(void*)) (runtime/cgo/gcc_unix.c)
// is called during startup. It saves setg_gcc and sets g0's stacklo. Because
// _cgo_init is non-nil, the runtime skips its own TLS setup. It can't be
// go:systemstack because the runtime is not set up far enough for that check.
//
// The stack bounds come from pthread_attr_get_np, like x_cgo_getstackbound
// in runtime/cgo/pthread_unix.c, because the default attributes report a
// stack size of 0 on illumos.
//
//go:nosplit
func x_cgo_init(g *G, setg uintptr) {
	var attr pthread_attr_t
	var addr uintptr
	var size size_t

	setg_func = setg

	pthread_attr_init(&attr)
	// pthread_t is 32 bits wide; the callee ignores the upper half of the register.
	pthread_attr_get_np(pthread_self(), &attr)
	pthread_attr_getstack(&attr, &addr, &size)
	pthread_attr_destroy(&attr)

	// libc derives the main thread's stack from RLIMIT_STACK, which gives a
	// bogus address when the limit is unlimited. runtime/cgo aborts in that
	// case; keep the conservative stacklo set by rt0_go instead.
	if addr != 0 && addr < g.stackhi {
		g.stacklo = addr
	}
}
