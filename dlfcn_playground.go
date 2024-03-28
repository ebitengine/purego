// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024 The Ebitengine Authors

//go:build faketime

package purego

import _ "unsafe"

// The playground doesn't support dynamic linking so just stub out the addresses
var (
	//go:linkname purego_dlopen purego_dlopen
	purego_dlopen uintptr
	//go:linkname purego_dlsym purego_dlsym
	purego_dlsym uintptr
	//go:linkname purego_dlerror purego_dlerror
	purego_dlerror uintptr
	//go:linkname purego_dlclose purego_dlclose
	purego_dlclose uintptr
)
