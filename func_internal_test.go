// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 The Ebitengine Authors

package purego

import (
	"reflect"
	"testing"
)

func TestAddValue_ArrayArgumentCanAddr(t *testing.T) {
	addr := [6]byte{'h', 'e', 'l', 'l', 'o', 0}

	tests := []struct {
		name             string
		value            reflect.Value
		wantCanAddr      bool
		wantKeepAliveLen int
	}{
		{
			name:             "non-addressable",
			value:            reflect.ValueOf([6]byte{'h', 'e', 'l', 'l', 'o', 0}),
			wantCanAddr:      false,
			wantKeepAliveLen: 1,
		},
		{
			name:             "addressable",
			value:            reflect.ValueOf(&addr).Elem(),
			wantCanAddr:      true,
			wantKeepAliveLen: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.value.CanAddr(); got != tt.wantCanAddr {
				t.Fatalf("CanAddr() = %v, want %v", got, tt.wantCanAddr)
			}

			var ints []uintptr
			addInt := func(x uintptr) { ints = append(ints, x) }
			addFloat := func(uintptr) {}
			addStack := func(uintptr) {}
			numInts, numFloats, numStack := 0, 0, 0

			keepAlive := addValue(tt.value, nil, addInt, addFloat, addStack, &numInts, &numFloats, &numStack)
			if len(ints) != 1 {
				t.Fatalf("len(ints) = %d, want 1", len(ints))
			}
			if ints[0] == 0 {
				t.Fatal("array pointer argument was zero")
			}
			if len(keepAlive) != tt.wantKeepAliveLen {
				t.Fatalf("len(keepAlive) = %d, want %d", len(keepAlive), tt.wantKeepAliveLen)
			}
		})
	}
}
