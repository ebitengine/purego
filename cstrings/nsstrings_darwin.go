package cstrings

import (
	"fmt"

	"github.com/ebitengine/purego"
	"github.com/ebitengine/purego/internal/strings"
	"github.com/ebitengine/purego/objc"
)

var (
	selIsKindOf   objc.SEL
	selUTF8String objc.SEL

	classNSString objc.Class
)

func init() {
	// Must pull in Foundation to get the NSString class.
	_, err := purego.Dlopen("/System/Library/Frameworks/Foundation.framework/Foundation", purego.RTLD_GLOBAL|purego.RTLD_NOW)
	if err != nil {
		panic(fmt.Errorf("nsstrings: %w", err))
	}
	selIsKindOf = objc.RegisterName("isKindOfClass:")
	selUTF8String = objc.RegisterName("UTF8String")
	classNSString = objc.GetClass("NSString")
}

// NSStringToString returns a copy of the NSString contents as a Go string.
// If the ID is 0 then an empty string is returned.
// If the ID is not an NSString class then the function panics.
// This function is only available on darwin.
func NSStringToString(str objc.ID) string {
	if str == 0 {
		return ""
	}
	if str.Send(selIsKindOf, classNSString) == 0 {
		panic("nsstrings: provided ID is not an NSString")
	}
	return strings.GoString(uintptr(str.Send(selUTF8String)))
}
