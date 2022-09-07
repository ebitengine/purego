//go:build !go1.16
// +build !go1.16

package purego

// callbackWrapPicker gets whatever is on the stack and in the first register.
// Depending on which version of Go that uses stack or register-based
// calling it passes the respective argument to the real calbackWrap function.
// The other argument is therefore invalid and points to undefined memory so don't use it.
// This function is necessary since we can't use the ABIInternal selector which is only
// valid in the runtime.
func callbackWrapPicker(stack, register *callbackArgs) {
	callbackWrap(stack)
}
