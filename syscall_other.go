// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 The Ebitengine Authors

//go:build linux && !amd64 && !arm64 && !loong64 && !riscv64

package purego

// getCallbackStart returns the start address of the callback region.
// This platform doesn't support the callback detection mechanism.
// TODO: Remove this function once callback tight packing is implemented.
func getCallbackStart() uintptr {
	panic("purego: getCallbackStart not supported on this platform")
}

// getMaxCB returns the maximum number of callbacks.
// This platform doesn't support the callback detection mechanism.
// TODO: Remove this function once callback tight packing is implemented.
func getMaxCB() int {
	panic("purego: getMaxCB not supported on this platform")
}
