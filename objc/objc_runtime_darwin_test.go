// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2022 The Ebitengine Authors

package objc_test

import (
	"fmt"
	"unsafe"

	"github.com/ebitengine/purego"
	"github.com/ebitengine/purego/objc"
)

func ExampleAllocateClassPair() {
	var class = objc.AllocateClassPair(objc.GetClass("NSObject"), "FooObject", 0)
	class.AddMethod(objc.RegisterName("run"), objc.NewIMP(func(self objc.ID, _cmd objc.SEL) {
		fmt.Println("Hello World!")
	}), "v@:")
	class.Register()

	var fooObject = objc.ID(class).Send(objc.RegisterName("new"))
	fooObject.Send(objc.RegisterName("run"))
	// Output: Hello World!
}

func ExampleClass_AddIvar() {
	var class = objc.AllocateClassPair(objc.GetClass("NSObject"), "BarObject", 0)
	class.AddIvar("bar", int(0), "q")
	var barOffset = class.InstanceVariable("bar").Offset()
	class.AddMethod(objc.RegisterName("bar"), objc.NewIMP(func(self objc.ID, _cmd objc.SEL) int {
		return *(*int)(unsafe.Pointer(uintptr(unsafe.Pointer(self)) + barOffset))
	}), "q@:")
	class.AddMethod(objc.RegisterName("setBar:"), objc.NewIMP(func(self objc.ID, _cmd objc.SEL, bar int) {
		*(*int)(unsafe.Pointer(uintptr(unsafe.Pointer(self)) + barOffset)) = bar
	}), "v@:q")
	class.Register()

	var barObject = objc.ID(class).Send(objc.RegisterName("new"))
	barObject.Send(objc.RegisterName("setBar:"), 123)
	var bar = int(barObject.Send(objc.RegisterName("bar")))
	fmt.Println(bar)
	// Output: 123
}

func ExampleIMP() {
	imp := objc.NewIMP(func(self objc.ID, _cmd objc.SEL) {
		fmt.Println("IMP:", self, _cmd)
	})

	purego.SyscallN(uintptr(imp), 105, 567)

	// Output: IMP: 105 567
}

func ExampleID_SendSuper() {
	super := objc.AllocateClassPair(objc.GetClass("NSObject"), "SuperObject", 0)
	super.AddMethod(objc.RegisterName("doSomething"), objc.NewIMP(func(self objc.ID, _cmd objc.SEL) {
		fmt.Println("In Super!")
	}), "v@:")
	super.Register()

	child := objc.AllocateClassPair(super, "ChildObject", 0)
	child.AddMethod(objc.RegisterName("doSomething"), objc.NewIMP(func(self objc.ID, _cmd objc.SEL) {
		fmt.Println("In Child")
		self.SendSuper(_cmd)
	}), "v@:")
	child.Register()

	objc.ID(child).Send(objc.RegisterName("new")).Send(objc.RegisterName("doSomething"))

	// Output: In Child
	// In Super!
}
