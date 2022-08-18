package fakecgo

import "unsafe"

var ptr unsafe.Pointer

/* Stub for creating a new thread */
//go:nosplit
func x_cgo_thread_start(arg *ThreadStart) {
	var ts *ThreadStart
	/* Make our own copy that can persist after we return. */
	//	_cgo_tsan_acquire();
	ts = (*ThreadStart)(malloc(unsafe.Sizeof(*ts)))
	println(ts)
	ptr = unsafe.Pointer(ts)
	//	_cgo_tsan_release();
	if ts == nil {
		println("fakecgo: out of memory in thread_start")
		abort()
	}
	memmove(unsafe.Pointer(ts), unsafe.Pointer(arg), unsafe.Sizeof(*ts))
	_cgo_sys_thread_start(ts) /* OS-dependent half */
}
