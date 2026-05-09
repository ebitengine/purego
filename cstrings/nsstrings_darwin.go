// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 The Ebitengine Authors

package cstrings

import (
	"fmt"

	"github.com/ebitengine/purego"
	"github.com/ebitengine/purego/internal/strings"
	"github.com/ebitengine/purego/objc"
)

var (
	sel_isKindOf   objc.SEL
	sel_UTF8String objc.SEL

	class_NSString objc.Class
)

func init() {
	// Must pull in Foundation to get the NSString class.
	_, err := purego.Dlopen("/System/Library/Frameworks/Foundation.framework/Foundation", purego.RTLD_GLOBAL|purego.RTLD_NOW)
	if err != nil {
		panic(fmt.Errorf("cstrings: %w", err))
	}
	sel_isKindOf = objc.RegisterName("isKindOfClass:")
	sel_UTF8String = objc.RegisterName("UTF8String")
	class_NSString = objc.GetClass("NSString")
}

// NSStringToString returns a copy of the NSString contents as a Go string.
// If the ID is 0 then an empty string is returned.
// If the ID is not an NSString class then the function panics.
// This function is only available on darwin.
func NSStringToString(str objc.ID) string {
	if str == 0 {
		return ""
	}
	if str.Send(sel_isKindOf, class_NSString) == 0 {
		panic("cstrings: provided ID is not an NSString")
	}
	return strings.GoString(uintptr(str.Send(sel_UTF8String)))
}
