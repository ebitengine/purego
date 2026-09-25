// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build !cgo && openbsd

package fakecgo

import _ "unsafe" // for go:linkname

// Supply environ and __progname, because we don't link against the standard
// OpenBSD crt0.o and the libc dynamic library needs them. Both are weak
// symbols in /usr/lib/libc.so.103.0.
//
// NetBSD's equivalent also supplies __ps_strings; OpenBSD's libc does not
// reference it, and libc here exports no such symbol.

//go:linkname _environ environ
//go:linkname _progname __progname

var (
	_environ  uintptr
	_progname uintptr
)
