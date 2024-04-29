// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024 The Ebitengine Authors

package load

import "golang.org/x/sys/windows"

func OpenLibrary(name string) (uintptr, error) {
	handle, err := windows.LoadLibrary(name)
	return uintptr(handle), err
}

func OpenSymbol(lib uintptr, name string) (uintptr, error) {
	return windows.GetProcAddress(windows.Handle(lib), name)
}
