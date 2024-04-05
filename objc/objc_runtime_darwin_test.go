// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2022 The Ebitengine Authors

package objc_test

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/ebitengine/purego"
	"github.com/ebitengine/purego/objc"
)

func ExampleRegisterClass_helloworld() {
	class, err := objc.RegisterClass(
		"FooObject",
		objc.GetClass("NSObject"),
		nil,
		nil,
		[]objc.MethodDef{
			{
				Cmd: objc.RegisterName("run"),
				Fn: func(self objc.ID, _cmd objc.SEL) {
					fmt.Println("Hello World!")
				},
			},
		},
	)
	if err != nil {
		panic(err)
	}

	object := objc.ID(class).Send(objc.RegisterName("new"))
	object.Send(objc.RegisterName("run"))
	// Output: Hello World!
}

func ExampleRegisterClass() {
	var (
		sel_new    = objc.RegisterName("new")
		sel_init   = objc.RegisterName("init")
		sel_setBar = objc.RegisterName("setBar:")
		sel_bar    = objc.RegisterName("bar")

		BarInit = func(id objc.ID, cmd objc.SEL) objc.ID {
			return id.SendSuper(cmd)
		}
	)

	class, err := objc.RegisterClass(
		"BarObject",
		objc.GetClass("NSObject"),
		[]*objc.Protocol{
			objc.GetProtocol("NSDelegateWindow"),
		},
		[]objc.FieldDef{
			{
				Name:      "bar",
				Type:      reflect.TypeOf(int(0)),
				Attribute: objc.ReadWrite,
			},
			{
				Name:      "foo",
				Type:      reflect.TypeOf(false),
				Attribute: objc.ReadWrite,
			},
		},
		[]objc.MethodDef{
			{
				Cmd: sel_init,
				Fn:  BarInit,
			},
		},
	)
	if err != nil {
		panic(err)
	}

	object := objc.ID(class).Send(sel_new)
	object.Send(sel_setBar, 123)
	bar := int(object.Send(sel_bar))
	fmt.Println(bar)
	// Output: 123
}

func ExampleIMP() {
	imp := objc.NewIMP(func(self objc.ID, _cmd objc.SEL, a3, a4, a5, a6, a7, a8, a9 int) {
		fmt.Println("IMP:", self, _cmd, a3, a4, a5, a6, a7, a8, a9)
	})

	purego.SyscallN(uintptr(imp), 105, 567, 9, 2, 3, ^uintptr(4), 4, 8, 9)
	// Output: IMP: 105 567 9 2 3 -5 4 8 9
}

func ExampleID_SendSuper() {
	super, err := objc.RegisterClass(
		"SuperObject",
		objc.GetClass("NSObject"),
		nil,
		nil,
		[]objc.MethodDef{
			{
				Cmd: objc.RegisterName("doSomething"),
				Fn: func(self objc.ID, _cmd objc.SEL) {
					fmt.Println("In Super!")
				},
			},
		},
	)
	if err != nil {
		panic(err)
	}

	child, err := objc.RegisterClass(
		"ChildObject",
		super,
		nil,
		nil,
		[]objc.MethodDef{
			{
				Cmd: objc.RegisterName("doSomething"),
				Fn: func(self objc.ID, _cmd objc.SEL) {
					fmt.Println("In Child")
					self.SendSuper(_cmd)
				},
			},
		},
	)
	if err != nil {
		panic(err)
	}

	objc.ID(child).Send(objc.RegisterName("new")).Send(objc.RegisterName("doSomething"))
	// Output: In Child
	// In Super!
}

func TestSend(t *testing.T) {
	// NSNumber comes from Foundation so make sure we have linked to that framework.
	_, err := purego.Dlopen("/System/Library/Frameworks/Foundation.framework/Foundation", purego.RTLD_GLOBAL)
	if err != nil {
		t.Fatal(err)
	}
	const double = float64(2.34)
	// Initialize a NSNumber
	NSNumber := objc.ID(objc.GetClass("NSNumber")).Send(objc.RegisterName("numberWithDouble:"), double)
	// Then get that number back using the generic Send function.
	number := objc.Send[float64](NSNumber, objc.RegisterName("doubleValue"))
	if double != number {
		t.Failed()
	}
}

func ExampleSend() {
	type NSRange struct {
		Location, Range uint
	}
	class_NSString := objc.GetClass("NSString")
	sel_stringWithUTF8String := objc.RegisterName("stringWithUTF8String:")

	fullString := objc.ID(class_NSString).Send(sel_stringWithUTF8String, "Hello, World!\x00")
	subString := objc.ID(class_NSString).Send(sel_stringWithUTF8String, "lo, Wor\x00")

	r := objc.Send[NSRange](fullString, objc.RegisterName("rangeOfString:"), subString)
	fmt.Println(r)
	// Output: {3 7}
}

func ExampleSendSuper() {
	super, err := objc.RegisterClass(
		"SuperObject2",
		objc.GetClass("NSObject"),
		nil,
		nil,
		[]objc.MethodDef{
			{
				Cmd: objc.RegisterName("doSomething"),
				Fn: func(self objc.ID, _cmd objc.SEL) int {
					return 16
				},
			},
		},
	)
	if err != nil {
		panic(err)
	}

	child, err := objc.RegisterClass(
		"ChildObject2",
		super,
		nil,
		nil,
		[]objc.MethodDef{
			{
				Cmd: objc.RegisterName("doSomething"),
				Fn: func(self objc.ID, _cmd objc.SEL) int {
					return 24
				},
			},
		},
	)
	if err != nil {
		panic(err)
	}

	res := objc.SendSuper[int](objc.ID(child).Send(objc.RegisterName("new")), objc.RegisterName("doSomething"))
	fmt.Println(res)
	// Output: 16
}
