// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 The Ebitengine Authors

package purego

// Source for constants: https://github.com/openbsd/src/blob/master/include/dlfcn.h
//
// Read from /usr/include/dlfcn.h on OpenBSD 7.9/arm64 rather than copied from
// another BSD: they happen to agree with NetBSD's, and checking is cheaper
// than finding out they had stopped agreeing.
const (
	intSize      = 32 << (^uint(0) >> 63) // 32 or 64
	RTLD_DEFAULT = 1<<intSize - 2         // Pseudo-handle for dlsym so search for any loaded symbol
	RTLD_LAZY    = 0x00000001             // Relocations are performed at an implementation-dependent time.
	RTLD_NOW     = 0x00000002             // Relocations are performed when the object is loaded.
	RTLD_LOCAL   = 0x00000000             // All symbols are not made available for relocation processing by other modules.
	RTLD_GLOBAL  = 0x00000100             // All symbols are available for relocation processing of other modules.
)
