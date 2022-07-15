// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2022 The Ebiten Authors

package strings

import "unsafe"

// hasSuffix tests whether the string s ends with suffix.
func hasSuffix(s, suffix string) bool {
	return len(s) >= len(suffix) && s[len(s)-len(suffix):] == suffix
}

// CString converts a go string to *byte that can be passed to C code.
// if requireNil is true it will panic if the passed string doesn't have
// a null byte at the end.
func CString(name string, requireNil bool) *byte {
	if hasSuffix(name, "\x00") {
		return &(*(*[]byte)(unsafe.Pointer(&name)))[0]
	}
	if requireNil {
		panic("null byte required")
	}
	var b = make([]byte, len(name)+1)
	copy(b, name)
	return &b[0]
}
