// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 The Ebitengine Authors

//go:build !windows

package main

import "github.com/ebitengine/purego"

func defaultMode() int {
	return purego.RTLD_NOW | purego.RTLD_GLOBAL
}
