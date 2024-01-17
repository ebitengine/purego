package cgo

/*
 #cgo LDFLAGS: -ldl

#include <dlfcn.h>
*/
import "C"
import _ "unsafe" // for go:linkname

// all that is needed is to assign each dl function because then its
// symbol will then be made available to the linker and linked to inside dlfcn.go
var (
	_ = C.dlopen
	_ = C.dlsym
	_ = C.dlerror
	_ = C.dlclose
)
