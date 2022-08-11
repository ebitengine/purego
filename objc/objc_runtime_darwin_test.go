// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2022 The Ebitengine Authors

package objc

import (
	"fmt"
	"math"
	"runtime"
	"unsafe"

	"github.com/ebitengine/purego"
)

func init() {
	// this is here so that the program doesn't crash with:
	// 		fatal error: schedule: in cgo
	// this can be removed after purego#12 is solved.
	// this is due to moving goroutines to different threads while in C code
	runtime.LockOSThread()
}

func ExampleAllocateClassPair() {
	var class = AllocateClassPair(GetClass("NSObject\x00"), "FooObject\x00", 0)
	class.AddMethod(RegisterName("run\x00"), IMP(func(self ID, _cmd SEL) {
		fmt.Println("Hello World!")
	}), "v@:\x00")
	class.Register()

	var fooObject = ID(class).Send(RegisterName("new\x00"))
	fooObject.Send(RegisterName("run\x00"))
	// Output: Hello World!
}

func ExampleClass_AddIvar() {
	var class = AllocateClassPair(GetClass("NSObject\x00"), "BarObject\x00", 0)
	class.AddIvar("bar\x00", unsafe.Sizeof(int(0)), uint8(math.Log2(float64(unsafe.Alignof(int(0))))), "q\x00")
	var barOffset = class.InstanceVariable("bar\x00").Offset()
	class.AddMethod(RegisterName("bar\x00"), IMP(func(self ID, _cmd SEL) int {
		return *(*int)(unsafe.Pointer(uintptr(unsafe.Pointer(self)) + barOffset))
	}), "q@:\x00")
	class.AddMethod(RegisterName("setBar:\x00"), IMP(func(self ID, _cmd SEL, bar int) {
		*(*int)(unsafe.Pointer(uintptr(unsafe.Pointer(self)) + barOffset)) = bar
	}), "v@:q\x00")
	class.Register()

	var barObject = ID(class).Send(RegisterName("new\x00"))
	barObject.Send(RegisterName("setBar:\x00"), 123)
	var bar = int(barObject.Send(RegisterName("bar\x00")))
	fmt.Println(bar)
	// Output: 123
}

func ExampleIMP() {
	imp := IMP(func(self ID, _cmd SEL) {
		fmt.Println("IMP:", self, _cmd)
	})

	purego.SyscallN(uintptr(imp), 105, 567)

	// Output: IMP: 105 567
}
