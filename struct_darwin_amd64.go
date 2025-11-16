// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 The Ebitengine Authors

//go:build darwin && amd64

package purego

import (
	"reflect"
)

// shouldBundleStackArgs always returns false on Darwin AMD64
// since C-style stack argument bundling is only needed on Darwin ARM64.
func shouldBundleStackArgs(v reflect.Value, numInts, numFloats int) bool {
	return false
}

// structFitsInRegisters is not used on Darwin AMD64.
// This stub exists for compilation but should never be called.
func structFitsInRegisters(val reflect.Value, tempNumInts, tempNumFloats int) (bool, int, int) {
	panic("purego: structFitsInRegisters should not be called on Darwin AMD64")
}

// collectStackArgs is not used on Darwin AMD64.
// This stub exists for compilation but should never be called.
func collectStackArgs(args []reflect.Value, startIdx int, numInts, numFloats int,
	keepAlive []any, addInt, addFloat, addStack func(uintptr),
	pNumInts, pNumFloats, pNumStack *int) ([]reflect.Value, []any) {
	panic("purego: collectStackArgs should not be called on Darwin AMD64")
}

// bundleStackArgs is not used on Darwin AMD64.
// This stub exists for compilation but should never be called.
func bundleStackArgs(stackArgs []reflect.Value, addStack func(uintptr)) {
	panic("purego: bundleStackArgs should not be called on Darwin AMD64")
}
