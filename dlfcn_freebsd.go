// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2022 The Ebitengine Authors

package purego

// Constants as defined in https://github.com/freebsd/freebsd-src/blob/main/include/dlfcn.h
const (
	RTLD_DEFAULT = 0xffffffff_fffffffd // Pseudo-handle for dlsym so search for any loaded symbol
	RTLD_LAZY    = 0x00000001          // Relocations are performed at an implementation-dependent time.
	RTLD_NOW     = 0x00000002          // Relocations are performed when the object is loaded.
	RTLD_LOCAL   = 0x00000000          // All symbols are not made available for relocation processing by other modules.
	RTLD_GLOBAL  = 0x00000100          // All symbols are available for relocation processing of other modules.
)
