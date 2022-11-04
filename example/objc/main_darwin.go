// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2022 The Ebitengine Authors

package main

import (
	"fmt"
	"unsafe"

	"github.com/ebitengine/purego/objc"
)

type barObject struct {
	isa objc.Class `objc:"BarObject:NSObject <NSDelegateWindow>"`
	bar int
}

func (b *barObject) Init(_cmd objc.SEL) objc.ID {
	return objc.ID(unsafe.Pointer(b)).SendSuper(_cmd)
}

func (b *barObject) Bar(_cmd objc.SEL) int {
	return b.bar
}

func (b *barObject) SetBar(_cmd objc.SEL, bar int) {
	b.bar = bar
}

func (_ *barObject) Selector(metName string) objc.SEL {
	switch metName {
	case "Init":
		return objc.RegisterName("init")
	case "SetBar":
		return objc.RegisterName("setBar:")
	case "Bar":
		return objc.RegisterName("bar")
	default:
		return 0
	}
}

func main() {
	class, err := objc.RegisterClass(&barObject{})
	if err != nil {
		panic(err)
	}

	var object = objc.ID(class).Send(objc.RegisterName("new"))
	object.Send(objc.RegisterName("setBar:"), 123)
	var bar = int(object.Send(objc.RegisterName("bar")))
	fmt.Println(bar)
	// Output: 123
}
