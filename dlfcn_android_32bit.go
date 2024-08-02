// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024 The Ebitengine Authors

//go:build android && (386 || arm)

package purego

const (
	RTLD_DEFAULT = 0xffffffff
	RTLD_LAZY    = 0x00000001
	RTLD_NOW     = 0x00000000
	RTLD_LOCAL   = 0x00000000
	RTLD_GLOBAL  = 0x00000002
)
