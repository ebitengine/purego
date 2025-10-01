// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2022 The Ebitengine Authors

//go:build !cgo && linux

package fakecgo

import "unsafe"

//go:linkname _cgo_libc_setegid syscall.cgo_libc_setegid
//go:linkname _cgo_libc_seteuid syscall.cgo_libc_seteuid
//go:linkname _cgo_libc_setgid syscall.cgo_libc_setgid
//go:linkname _cgo_libc_setregid syscall.cgo_libc_setregid
//go:linkname _cgo_libc_setresgid syscall.cgo_libc_setresgid
//go:linkname _cgo_libc_setresuid syscall.cgo_libc_setresuid
//go:linkname _cgo_libc_setreuid syscall.cgo_libc_setreuid
//go:linkname _cgo_libc_setuid syscall.cgo_libc_setuid
//go:linkname _cgo_libc_setgroups syscall.cgo_libc_setgroups

var _cgo_libc_setegid = &_cgo_purego_setegid_trampoline
var _cgo_libc_seteuid = &_cgo_purego_seteuid_trampoline
var _cgo_libc_setgid = &_cgo_purego_setgid_trampoline
var _cgo_libc_setregid = &_cgo_purego_setregid_trampoline
var _cgo_libc_setresgid = &_cgo_purego_setresgid_trampoline
var _cgo_libc_setresuid = &_cgo_purego_setresuid_trampoline
var _cgo_libc_setreuid = &_cgo_purego_setreuid_trampoline
var _cgo_libc_setuid = &_cgo_purego_setuid_trampoline
var _cgo_libc_setgroups = &_cgo_purego_setgroups_trampoline

func errno() int32 {
	// this indirection is to avoid go vet complaining about possible misuse of unsafe.Pointer
	loc := __errno_location()
	return **(**int32)(unsafe.Pointer(&loc))
}
