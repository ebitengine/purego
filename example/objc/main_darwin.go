// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2022 The Ebitengine Authors

package main

import (
	"fmt"
	"unsafe"

	"github.com/ebitengine/purego/objc"
)

var (
	sel_new    = objc.RegisterName("new")
	sel_init   = objc.RegisterName("init")
	sel_setBar = objc.RegisterName("setBar:")
	sel_bar    = objc.RegisterName("bar")
)

type barObject struct {
	isa objc.Class `objc:"BarObject : NSObject <NSDelegateWindow>"`
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
		return sel_init
	case "SetBar":
		return sel_setBar
	case "Bar":
		return sel_bar
	default:
		return 0
	}
}

func main() {
	class, err := objc.RegisterClass(&barObject{})
	if err != nil {
		panic(err)
	}

	var object = objc.ID(class).Send(sel_new)
	object.Send(sel_setBar, 123)
	var bar = int(object.Send(sel_bar))
	fmt.Println(bar)
	// Output: 123
}
