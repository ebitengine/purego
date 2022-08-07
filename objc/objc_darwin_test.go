package objc

import (
	"fmt"
	"runtime"
)

func init() {
	// this is here so that the program doesn't crash with:
	// 		fatal error: schedule: in cgo
	// this can be removed after purego#12 is solved.
	runtime.LockOSThread()
}

func ExampleAllocateClassPair() {
	var class = AllocateClassPair(GetClass("NSObject\x00"), "FooObject\x00", 0)
	class.AddMethod(RegisterName("run\x00"), IMP(func(self ID, _cmd SEL) {
		fmt.Println("Hello World!")
	}), "v@:\x00")
	class.Register()

	var FooObject = ID(class).Send(RegisterName("new\x00"))
	FooObject.Send(RegisterName("run\x00"))
	// Output: Hello World!
}
