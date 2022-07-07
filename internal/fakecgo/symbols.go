package fakecgo

import _ "unsafe"

// let's pretend we have cgo:
//go:linkname _iscgo runtime.iscgo
var _iscgo = false

// Now all the symbols we need to import from various libraries to implement the above functions:
// (just using one variable and taking the address in libFuncs.go works with amd64 - but the two variable dance is needed for 386, where we get an unknown symbol relocation otherwise :/)

//go:linkname libc_pthread_attr_init_x libc_pthread_attr_init_x
var libc_pthread_attr_init_x uintptr
var libc_pthread_attr_init = &libc_pthread_attr_init_x

//go:linkname libc_pthread_attr_getstacksize_x libc_pthread_attr_getstacksize_x
var libc_pthread_attr_getstacksize_x uintptr
var libc_pthread_attr_getstacksize = &libc_pthread_attr_getstacksize_x

//go:linkname libc_pthread_attr_destroy_x libc_pthread_attr_destroy_x
var libc_pthread_attr_destroy_x uintptr
var libc_pthread_attr_destroy = &libc_pthread_attr_destroy_x

//go:linkname libc_pthread_sigmask_x libc_pthread_sigmask_x
var libc_pthread_sigmask_x uintptr
var libc_pthread_sigmask = &libc_pthread_sigmask_x

//go:linkname libc_pthread_create_x libc_pthread_create_x
var libc_pthread_create_x uintptr
var libc_pthread_create = &libc_pthread_create_x

//go:linkname libc_pthread_detach_x libc_pthread_detach_x
var libc_pthread_detach_x uintptr
var libc_pthread_detach = &libc_pthread_detach_x

//go:linkname libc_setenv_x libc_setenv_x
var libc_setenv_x uintptr
var libc_setenv = &libc_setenv_x

//go:linkname libc_unsetenv_x libc_unsetenv_x
var libc_unsetenv_x uintptr
var libc_unsetenv = &libc_unsetenv_x

//go:linkname libc_malloc_x libc_malloc_x
var libc_malloc_x uintptr
var libc_malloc = &libc_malloc_x

//go:linkname libc_free_x libc_free_x
var libc_free_x uintptr
var libc_free = &libc_free_x

//go:linkname libc_nanosleep_x libc_nanosleep_x
var libc_nanosleep_x uintptr
var libc_nanosleep = &libc_nanosleep_x

//go:linkname libc_sigfillset_x libc_sigfillset_x
var libc_sigfillset_x uintptr
var libc_sigfillset = &libc_sigfillset_x

//go:linkname libc_abort_x libc_abort_x
var libc_abort_x uintptr
var libc_abort = &libc_abort_x

//go:linkname libc_dprintf_x libc_dprintf_x
var libc_dprintf_x uintptr
var libc_dprintf = &libc_dprintf_x

//go:linkname libc_strerror_x libc_strerror_x
var libc_strerror_x uintptr
var libc_strerror = &libc_strerror_x
