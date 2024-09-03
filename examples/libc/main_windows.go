// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2023 The Ebitengine Authors

package main

import "syscall"

func openLibrary(name string) (uintptr, error) {
	// Use [syscall.LoadLibrary] here to avoid external dependencies (#270).
	// For actual use cases, [golang.org/x/sys/windows.NewLazySystemDLL] is recommended.
	handle, err := syscall.LoadLibrary(name)
	return uintptr(handle), err
}
