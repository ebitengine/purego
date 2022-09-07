// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2022 The Ebitengine Authors

//go:build go1.16
// +build go1.16

package purego

// callbackWrapPicker gets called with whatever is on the stack and in the first register.
// Depending on which version of Go that uses stack or register-based
// calling it passes the respective argument to the real calbackWrap function.
// The other argument is therefore invalid and points to undefined memory so don't use it.
// This function is necessary since we can't use the ABIInternal selector which is only
// valid in the runtime.
//
// the double indirection is because Go 1.15 will do that in runtime.cgocallback
// but in 1.16+ it doesn't so we do it here.
func callbackWrapPicker(stack, register **callbackArgs) {
	if stackCallingConvention {
		callbackWrap(*stack)
	} else {
		callbackWrap(*register)
	}
}
