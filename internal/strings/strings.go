// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2022 The Ebitengine Authors

package strings

import (
	"reflect"
	"unsafe"
)

// hasSuffix tests whether the string s ends with suffix.
func hasSuffix(s, suffix string) bool {
	return len(s) >= len(suffix) && s[len(s)-len(suffix):] == suffix
}

// CString converts a go string to *byte that can be passed to C code.
func CString(name string) *byte {
	if hasSuffix(name, "\x00") {
		return &(*(*[]byte)(unsafe.Pointer(&name)))[0]
	}
	var b = make([]byte, len(name)+1)
	copy(b, name)
	return &b[0]
}

// GoString copies a char* to a Go string.
func GoString(c uintptr) string {
	// We take the address and then dereference it to trick go vet from creating a possible misuse of unsafe.Pointer
	ptr := *(*unsafe.Pointer)(unsafe.Pointer(&c))
	if ptr == nil {
		return ""
	}
	var length int
	for {
		// use unsafe.Add once we reach 1.17
		if *(*byte)(unsafe.Pointer(uintptr(ptr) + uintptr(length))) == '\x00' {
			break
		}
		length++
	}
	// use unsafe.Slice once we reach 1.17
	s := make([]byte, length)
	var src []byte
	h := (*reflect.SliceHeader)(unsafe.Pointer(&src))
	h.Data = uintptr(ptr)
	h.Len = length
	h.Cap = length
	copy(s, src)
	return string(s)
}
