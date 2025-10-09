// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 The Ebitengine Authors

//go:build darwin || freebsd || linux || netbsd

package purego

func openEmbeddedHandle(path string, mode int) (uintptr, func(uintptr) error, error) {
	handle, err := Dlopen(path, mode)
	if err != nil {
		return 0, nil, err
	}
	return handle, Dlclose, nil
}
