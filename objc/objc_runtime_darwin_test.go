// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2022 The Ebitengine Authors

package objc_test

import (
	"fmt"
	"runtime"
	"unsafe"

	"github.com/ebitengine/purego"
	"github.com/ebitengine/purego/objc"
)

func init() {
	// this is here so that the program doesn't crash with:
	// 		fatal error: schedule: in cgo
	// this can be removed after purego#12 is solved.
	// this is due to moving goroutines to different threads while in C code
	runtime.LockOSThread()
}

func ExampleAllocateClassPair() {
	var class = objc.AllocateClassPair(objc.GetClass("NSObject\x00"), "FooObject\x00", 0)
	class.AddMethod(objc.RegisterName("run\x00"), objc.IMP(func(self objc.ID, _cmd objc.SEL) {
		fmt.Println("Hello World!")
	}), "v@:\x00")
	class.Register()

	var fooObject = objc.ID(class).Send(objc.RegisterName("new\x00"))
	fooObject.Send(objc.RegisterName("run\x00"))
	// Output: Hello World!
}

func ExampleClass_AddIvar() {
	var class = objc.AllocateClassPair(objc.GetClass("NSObject\x00"), "BarObject\x00", 0)
	class.AddIvar("bar\x00", int(0), "q\x00")
	var barOffset = class.InstanceVariable("bar\x00").Offset()
	class.AddMethod(objc.RegisterName("bar\x00"), objc.IMP(func(self objc.ID, _cmd objc.SEL) int {
		return *(*int)(unsafe.Pointer(uintptr(unsafe.Pointer(self)) + barOffset))
	}), "q@:\x00")
	class.AddMethod(objc.RegisterName("setBar:\x00"), objc.IMP(func(self objc.ID, _cmd objc.SEL, bar int) {
		*(*int)(unsafe.Pointer(uintptr(unsafe.Pointer(self)) + barOffset)) = bar
	}), "v@:q\x00")
	class.Register()

	var barObject = objc.ID(class).Send(objc.RegisterName("new\x00"))
	barObject.Send(objc.RegisterName("setBar:\x00"), 123)
	var bar = int(barObject.Send(objc.RegisterName("bar\x00")))
	fmt.Println(bar)
	// Output: 123
}

func ExampleIMP() {
	imp := objc.IMP(func(self objc.ID, _cmd objc.SEL) {
		fmt.Println("IMP:", self, _cmd)
	})

	purego.SyscallN(uintptr(imp), 105, 567)

	// Output: IMP: 105 567
}
