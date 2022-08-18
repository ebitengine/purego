package fakecgo

import "unsafe"

//go:nosplit
func _cgo_sys_thread_start(ts *ThreadStart) {
	var attr pthread_attr_t
	var ign, oset sigset_t
	var p pthread_t
	var size size_t
	var err int

	//print("fakecgo: _cgo_sys_thread_start: fn=")
	//print(ts.fn)
	//print(", g=")
	//println(ts.g)
	//fprintf(stderr, "runtime/cgo: _cgo_sys_thread_start: fn=%p, g=%p\n", ts->fn, ts->g); // debug
	sigfillset(&ign)
	pthread_sigmask(SIG_SETMASK, &ign, &oset)

	pthread_attr_init(&attr)
	pthread_attr_getstacksize(&attr, &size)
	// Leave stacklo=0 and set stackhi=size; mstart will do the rest.
	ts.g.stackhi = uintptr(size)
	print("threadentry: ")
	println(threadentry_trampolineABI0)
	err = _cgo_try_pthread_create(&p, &attr, threadentry_trampolineABI0, ts)

	pthread_sigmask(SIG_SETMASK, &oset, nil)

	if err != 0 {
		print("fakecgo: pthread_create failed: ")
		println(err)
		abort()
	}
}

// threadentry_trampolineABI0 maps the C ABI to Go ABI then calls the Go functions
var threadentry_trampolineABI0 uintptr

func threadentry(v unsafe.Pointer) unsafe.Pointer {
	println("made it here")
	ts := *(*ThreadStart)(v)
	free(v)

	// TODO: support ios
	//#if TARGET_OS_IPHONE
	//	darwin_arm_init_thread_exception_port();
	//#endif
	_ = ts
	//	crosscall1(ts.fn, setg_gcc, (void*)ts.g);
	return nil
}

// here we will store a pointer to the provided setg func
var setg_func uintptr

//go:nosplit
func x_cgo_init(g *G, setg uintptr) {
	var size size_t
	var attr pthread_attr_t

	setg_func = setg
	pthread_attr_init(&attr)
	pthread_attr_getstacksize(&attr, &size)
	g.stacklo = uintptr(unsafe.Pointer(&size)) - uintptr(size) + 4096
	pthread_attr_destroy(&attr)

	//TODO: support ios
	//#if TARGET_OS_IPHONE
	//	darwin_arm_init_mach_exception_handler();
	//	darwin_arm_init_thread_exception_port();
	//	init_working_dir();
	//#endif
}
