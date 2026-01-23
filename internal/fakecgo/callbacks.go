// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build !cgo && (darwin || freebsd || linux || netbsd)

package fakecgo

import (
	_ "unsafe"
)

// TODO: decide if we need _runtime_cgo_panic_internal

// Indicates whether a dummy thread key has been created or not.
//
// When calling go exported function from C, we register a destructor
// callback, for a dummy thread key, by using pthread_key_create.

//go:linkname _cgo_pthread_key_created _cgo_pthread_key_created
var x_cgo_pthread_key_created uintptr
var _cgo_pthread_key_created = &x_cgo_pthread_key_created

// Set the x_crosscall2_ptr C function pointer variable point to crosscall2.
// It's for the runtime package to call at init time.
func set_crosscall2() {
	// nothing needs to be done here for fakecgo
	// because it's possible to just call cgocallback directly
}

//go:linkname _set_crosscall2 runtime.set_crosscall2
var _set_crosscall2 = set_crosscall2

// TODO: decide if we need x_cgo_set_context_function
// TODO: decide if we need _cgo_yield
