// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2022 The Ebitengine Authors

package main

import (
	"fmt"
	"reflect"

	"github.com/ebitengine/purego/objc"
)

var (
	sel_new    = objc.RegisterName("new")
	sel_init   = objc.RegisterName("init")
	sel_setBar = objc.RegisterName("setBar:")
	sel_bar    = objc.RegisterName("bar")
)

func BarInit(id objc.ID, _cmd objc.SEL) objc.ID {
	return id.SendSuper(_cmd)
}

func main() {
	/*
		This struct is equivalent to the following Objective-C definition.

		@interface BarObject : NSObject <NSDelegateWindow>
		@property (readwrite) bar int
		@end
	*/
	class, err := objc.RegisterClass(
		"BarObject",
		objc.GetClass("NSObject"),
		[]*objc.Protocol{
			objc.GetProtocol("NSDelegateWindow"),
		},
		[]objc.IvarDef{
			{
				Name:      "bar",
				Type:      reflect.TypeOf(int(0)),
				Attribute: objc.ReadWrite,
			},
		},
		[]objc.MethodDef{
			{sel_init, objc.NewIMP(BarInit)},
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
