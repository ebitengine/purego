// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 The Ebitengine Authors

//go:build !darwin

package purego

import (
	"reflect"
)

func placeRegisters(v reflect.Value, addFloat func(uintptr), addInt func(uintptr)) {
	placeRegistersArm64(v, addFloat, addInt)
}
