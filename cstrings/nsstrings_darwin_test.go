// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 The Ebitengine Authors

package cstrings_test

import (
	"testing"

	"github.com/ebitengine/purego/cstrings"
	"github.com/ebitengine/purego/objc"
)

func TestNSStringToString(t *testing.T) {
	t.Run("nil ID returns empty string", func(t *testing.T) {
		result := cstrings.NSStringToString(0)
		if result != "" {
			t.Errorf("expected empty string for nil ID, got %q", result)
		}
	})

	t.Run("valid NSString returns correct string", func(t *testing.T) {
		sel := objc.RegisterName("stringWithUTF8String:")
		nsString := objc.ID(objc.GetClass("NSString")).Send(sel, "Hello, World!\x00")

		result := cstrings.NSStringToString(nsString)
		if result != "Hello, World!" {
			t.Errorf("expected %q, got %q", "Hello, World!", result)
		}
	})

	t.Run("empty NSString returns empty string", func(t *testing.T) {
		sel := objc.RegisterName("stringWithUTF8String:")
		nsString := objc.ID(objc.GetClass("NSString")).Send(sel, "\x00")

		result := cstrings.NSStringToString(nsString)
		if result != "" {
			t.Errorf("expected empty string, got %q", result)
		}
	})

	t.Run("non-NSString panics", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("expected panic for non-NSString ID")
			}
		}()

		classNSNumber := objc.GetClass("NSNumber")
		sel := objc.RegisterName("numberWithInt:")
		nsNumber := objc.ID(classNSNumber).Send(sel, 42)

		cstrings.NSStringToString(nsNumber)
	})
}
