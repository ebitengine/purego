// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2022 The Ebitengine Authors

//go:build !go1.16
// +build !go1.16

package purego

// callbackWrapPicker is here because Go 1.16+ needs it
// to decide between stack and register arguments.
// This only has a single pointer because Go 1.15 runtime.cgocallback
// dereferences the argument once before passing it to this
// function.
func callbackWrapPicker(stack, _ *callbackArgs) {
	callbackWrap(stack)
}
