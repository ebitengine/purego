// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 The Ebitengine Authors

//go:build darwin || freebsd || linux || netbsd || windows

package purego

// MaxArgs re-exports maxArgs for external tests.
const MaxArgs = maxArgs

// StructReturnInMemory re-exports structReturnInMemory for external tests.
func StructReturnInMemory(size uintptr) bool {
	return structReturnInMemory(size)
}
