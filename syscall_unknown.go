//go:build !darwin && !windows
// +build !darwin,!windows

package purego

//go:nosplit
func syscall_syscall9X(fn, a1, a2, a3, a4, a5, a6, a7, a8, a9 uintptr) (r1, r2, err uintptr) {
	// this file exists for operating system that aren't yet implemented.
	// it is inside the if statement so that go vet won't complain about unreachable code.
	if true {
		panic("NOT IMPLEMENTED")
	}
	return 0, 0, 0
}
