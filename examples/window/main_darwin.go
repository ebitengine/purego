// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024 The Ebitengine Authors

package main

import (
	"runtime"

	"github.com/ebitengine/purego"
	"github.com/ebitengine/purego/objc"
)

const (
	NSApplicationActivationPolicyRegular = 0
	NSWindowStyleMaskTitled              = 1 << 0
	NSBackingStoreBuffered               = 2
)

type NSPoint struct {
	X, Y float64
}

type NSSize struct {
	Width, Height float64
}

type NSRect struct {
	Origin NSPoint
	Size   NSSize
}

func init() {
	runtime.LockOSThread()
}

func main() {
	if _, err := purego.Dlopen("/System/Library/Frameworks/Cocoa.framework/Cocoa", purego.RTLD_GLOBAL|purego.RTLD_LAZY); err != nil {
		panic(err)
	}
	nsApp := objc.ID(objc.GetClass("NSApplication")).Send(objc.RegisterName("sharedApplication"))
	nsApp.Send(objc.RegisterName("setActivationPolicy:"), NSApplicationActivationPolicyRegular)
	wnd := objc.ID(objc.GetClass("NSWindow")).Send(objc.RegisterName("alloc"))
	wnd = wnd.Send(objc.RegisterName("initWithContentRect:styleMask:backing:defer:"),
		NSMakeRect(0, 0, 320, 240),
		NSWindowStyleMaskTitled,
		NSBackingStoreBuffered,
		false,
	)

	title := objc.ID(objc.GetClass("NSString")).Send(objc.RegisterName("stringWithUTF8String:"), "My Title")
	wnd.Send(objc.RegisterName("setTitle:"), title)
	wnd.Send(objc.RegisterName("makeKeyAndOrderFront:"), objc.ID(0))
	wnd.Send(objc.RegisterName("center"))
	nsApp.Send(objc.RegisterName("activateIgnoringOtherApps:"), true)
	nsApp.Send(objc.RegisterName("run"))
}

func NSMakeRect(x, y, width, height float64) NSRect {
	return NSRect{Origin: NSPoint{x, y}, Size: NSSize{width, height}}
}
