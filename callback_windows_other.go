// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 The Ebitengine Authors

//go:build windows && !386

package purego

import "syscall"

func newCallback(fn any, isCDecl bool) uintptr {
	if isCDecl {
		return syscall.NewCallbackCDecl(fn)
	}
	return syscall.NewCallback(fn)
}
