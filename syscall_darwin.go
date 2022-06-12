package purego

import "unsafe"

func callc(fn uintptr, args unsafe.Pointer) {
	runtime_entersyscall()
	runtime_libcCall(unsafe.Pointer(fn), args)
	runtime_exitsyscall()
}

//go:nosplit
func syscall_syscallX3(fn, a1, a2, a3, _, _, _, _, _, _ uintptr) (r1, r2, err uintptr) {
	return syscall_syscall6X(fn, a1, a2, a3, 0, 0, 0)
}

//go:nosplit
func syscall_syscallX6(fn, a1, a2, a3, a4, a5, a6, _, _, _ uintptr) (r1, r2, err uintptr) {
	return syscall_syscall6X(fn, a1, a2, a3, a4, a5, a6)
}

var syscall9XABI0 uintptr

func syscall9X() // implemented in assembly

//go:nosplit
func syscall_syscall9X(fn, a1, a2, a3, a4, a5, a6, a7, a8, a9 uintptr) (r1, r2, err uintptr) {
	args := struct{ fn, a1, a2, a3, a4, a5, a6, a7, a8, a9, r1, r2, err uintptr }{
		fn, a1, a2, a3, a4, a5, a6, a7, a8, a9, r1, r2, err}
	callc(syscall9XABI0, unsafe.Pointer(&args))
	return args.r1, args.r2, args.err
}
