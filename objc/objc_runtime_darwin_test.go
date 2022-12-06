// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2022 The Ebitengine Authors

package objc_test

import (
	"fmt"

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
	type barObject struct {
		isa objc.Class
		bar int
	}
	var class = objc.AllocateClassPair(objc.GetClass("NSObject"), "BarObject", 0)
	class.AddIvar("bar", int(0), "q")
	class.AddMethod(objc.RegisterName("bar"), objc.NewIMP(func(self *barObject, _cmd objc.SEL) int {
		return self.bar
	}), "q@:")
	class.AddMethod(objc.RegisterName("setBar:"), objc.NewIMP(func(self *barObject, _cmd objc.SEL, bar int) {
		self.bar = bar
	}), "v@:q")
	class.Register()

	var object = objc.ID(class).Send(objc.RegisterName("new"))
	object.Send(objc.RegisterName("setBar:"), 123)
	var bar = int(object.Send(objc.RegisterName("bar")))
	fmt.Println(bar)
	// Output: 123
}

func ExampleIMP() {
	imp := objc.NewIMP(func(self objc.ID, _cmd objc.SEL, a3, a4, a5, a6, a7, a8, a9 int,
		f1, f2, f3, f4, f5, f6, f7, f8 float64) {
		fmt.Println("IMP:", self, _cmd, a3, a4, a5, a6, a7, a8, a9, f1, f2, f3, f4, f5, f6, f7, f8)
	})
	var fn func(self objc.ID, _cmd objc.SEL, a3, a4, a5, a6, a7, a8, a9 int,
		f1, f2, f3, f4, f5, f6, f7, f8 float64)
	purego.RegisterFunc(&fn, uintptr(imp))
	fn(105, 567, 9, 2, 3, -5, 4, 8, 9,
		1.23, 2.45, 5.43, 6.12, 7.8, 1.99, 4.32, 12.5)

	// Output: IMP: 105 567 9 2 3 -5 4 8 9 1.23 2.45 5.43 6.12 7.8 1.99 4.32 12.5
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
