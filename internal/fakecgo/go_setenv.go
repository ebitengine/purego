package fakecgo

import "unsafe"

//go:nosplit
func x_cgo_setenv(arg *[2]*byte) {
	setenv(arg[0], arg[1], 1)
}

//go:nosplit
func x_cgo_unsetenv(arg *[1]*byte) {
	print("out: ")
	println(string(unsafe.Slice(arg[0], len("IGNORE_ME\x00"))))
	unsetenv(arg[0])
}
