// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2022 The Ebitengine Authors

//go:build darwin
// +build darwin

package fakecgo

type size_t uintptr

type sigset_t [128]byte      // TODO: figure out how big this should be
type pthread_attr_t [56]byte // TODO: figure out how big this should be
type pthread_t int

// We could take timespec from syscall - but there it uses int32 and int64 for 32 bit and 64 bit arch, which complicates stuff for us
type timespec struct {
	tv_sec  int
	tv_nsec int
}

// for pthread_sigmask:

type sighow int32

const (
	SIG_BLOCK   sighow = 0
	SIG_UNBLOCK sighow = 1
	SIG_SETMASK sighow = 2
)

type G struct {
	stacklo uintptr
	stackhi uintptr
}

type ThreadStart struct {
	g   *G
	tls *uintptr
	fn  uintptr
}
