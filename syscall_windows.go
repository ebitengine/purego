package purego

import (
	"syscall"
)

//go:nosplit
func syscall_syscallX3(fn, a1, a2, a3, _, _, _, _, _, _ uintptr) (r1, r2, err uintptr) {
	r1, r2, errno := syscall.Syscall(fn, 3, a1, a2, a3)
	return r1, r2, uintptr(errno)
}

//go:nosplit
func syscall_syscallX6(fn, a1, a2, a3, a4, a5, a6, _, _, _ uintptr) (r1, r2, err uintptr) {
	r1, r2, errno := syscall.Syscall6(fn, 6, a1, a2, a3, a4, a5, a6)
	return r1, r2, uintptr(errno)
}

//go:nosplit
func syscall_syscall9X(fn, a1, a2, a3, a4, a5, a6, a7, a8, a9 uintptr) (r1, r2, err uintptr) {
	r1, r2, errno := syscall.Syscall9(fn, 9, a1, a2, a3, a4, a5, a6, a7, a8, a9)
	return r1, r2, uintptr(errno)
}
