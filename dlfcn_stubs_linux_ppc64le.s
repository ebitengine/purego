// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 The Ebitengine Authors

//go:build !cgo && !faketime

#include "textflag.h"

// func dlopen(path *byte, mode int) (ret uintptr)
TEXT dlopen(SB), NOSPLIT, $0-0
	CALL purego_dlopen(SB)
	// The NOP is the TOC restore slot. The linker rewrites it to
	// MOVD 24(R1), R2 when linking PIC code.
	WORD $0x60000000
	RET

// func dlsym(handle uintptr, symbol *byte) (ret uintptr)
TEXT dlsym(SB), NOSPLIT, $0-0
	CALL purego_dlsym(SB)
	// The NOP is the TOC restore slot. The linker rewrites it to
	// MOVD 24(R1), R2 when linking PIC code.
	WORD $0x60000000
	RET

// func dlerror() (ret *byte)
TEXT dlerror(SB), NOSPLIT, $0-0
	CALL purego_dlerror(SB)
	// The NOP is the TOC restore slot. The linker rewrites it to
	// MOVD 24(R1), R2 when linking PIC code.
	WORD $0x60000000
	RET

// func dlclose(handle uintptr) (ret int)
TEXT dlclose(SB), NOSPLIT, $0-0
	CALL purego_dlclose(SB)
	// The NOP is the TOC restore slot. The linker rewrites it to
	// MOVD 24(R1), R2 when linking PIC code.
	WORD $0x60000000
	RET
