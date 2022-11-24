// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2022 The Ebiten Authors

package purego

// Source for constants: https://codebrowser.dev/glibc/glibc/bits/dlfcn.h.html

const (
	RTLD_LAZY   = 0x00001 // Relocations are performed at an implementation-dependent time.
	RTLD_NOW    = 0x00002 // Relocations are performed when the object is loaded.
	RTLD_LOCAL  = 0x00000 // All symbols are not made available for relocation processing by other modules.
	RTLD_GLOBAL = 0x00100 // All symbols are available for relocation processing of other modules.
)