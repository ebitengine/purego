// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024 The Ebitengine Authors

//go:build android && (386 || arm)

package purego

// Source for constants: https://android.googlesource.com/platform/bionic/+/refs/heads/main/libc/include/dlfcn.h

const (
	RTLD_DEFAULT = 0xffffffff // Pseudo-handle for dlsym so search for any loaded symbol
	RTLD_LAZY    = 0x00000001 // Relocations are performed at an implementation-dependent time.
	RTLD_NOW     = 0x00000000 // Relocations are performed when the object is loaded.
	RTLD_LOCAL   = 0x00000000 // All symbols are not made available for relocation processing by other modules.
	RTLD_GLOBAL  = 0x00000002 // All symbols are available for relocation processing of other modules.
)
