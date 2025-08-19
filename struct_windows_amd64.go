// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024 The Ebitengine Authors

//go:build windows && amd64

package purego

import (
	"reflect"
)

func addStruct(v reflect.Value, numInts, numFloats, numStack *int, addInt, addFloat, addStack func(uintptr), keepAlive []any) []any {
	panic("not implemented")
}

func getStruct(outType reflect.Type, syscall syscall15Args) (v reflect.Value) {
	panic("not implemented")
}

func placeRegisters(v reflect.Value, addInt, _ func(uintptr), keepAlive []any) []any {
	panic("not implemented")
}

func placeStack(v reflect.Value, addInt, addStack func(uintptr), keepAlive []any) []any {
	panic("not implemented")
}
