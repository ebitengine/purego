// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 The Ebitengine Authors

//go:build windows

package purego

import "syscall"

func openEmbeddedHandle(path string, mode int) (uintptr, func(uintptr) error, error) {
	_ = mode
	handle, err := syscall.LoadLibrary(path)
	if err != nil {
		return 0, nil, err
	}
	closeFn := func(h uintptr) error {
		return syscall.FreeLibrary(syscall.Handle(h))
	}
	return uintptr(handle), closeFn, nil
}
