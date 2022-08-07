package objc

import (
	"fmt"
	"math"
	"runtime"
	"unsafe"
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

func ExampleClass_AddIvar() {
	var class = AllocateClassPair(GetClass("NSObject\x00"), "FooObject\x00", 0)
	class.AddIvar("bar", unsafe.Sizeof(int(0)), uint8(math.Log2(float64(unsafe.Sizeof(int(0))))), "q")
	var barOffset = class.InstanceVariable("bar\x00").Offset()
	class.AddMethod(RegisterName("foo\x00"), IMP(func(self ID, _cmd SEL) int {
		return *(*int)(unsafe.Pointer(uintptr(unsafe.Pointer(self)) + barOffset))
	}), "q@:\x00")
	class.AddMethod(RegisterName("setFoo:\x00"), IMP(func(self ID, _cmd SEL, foo int) {
		*(*int)(unsafe.Pointer(uintptr(unsafe.Pointer(self)) + barOffset)) = foo
	}), "v@:q")
	class.Register()

	var FooObject = ID(class).Send(RegisterName("new\x00"))
	FooObject.Send(RegisterName("setFoo:\x00"), 123)
	var foo = int(FooObject.Send(RegisterName("foo\x00")))
	fmt.Println(foo)
	// Output: 123
}
