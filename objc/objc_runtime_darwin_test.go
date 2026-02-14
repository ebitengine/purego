// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2022 The Ebitengine Authors

package objc_test

import (
	"fmt"
	"log"
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
	_, err := purego.Dlopen("/System/Library/Frameworks/Foundation.framework/Foundation", purego.RTLD_GLOBAL|purego.RTLD_NOW)
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

func ExampleAllocateProtocol() {
	var p *objc.Protocol
	if p = objc.AllocateProtocol("MyCustomProtocol"); p != nil {
		p.AddMethodDescription(objc.RegisterName("isFoo"), "B16@0:8", true, true)
		var adoptedProtocol *objc.Protocol
		adoptedProtocol = objc.GetProtocol("NSObject")
		if adoptedProtocol == nil {
			log.Fatalln("protocol 'NSObject' does not exist")
		}
		p.AddProtocol(adoptedProtocol)
		p.AddProperty("accessibilityElement", []objc.PropertyAttribute{
			{Name: &[]byte("T\x00")[0], Value: &[]byte("B\x00")[0]},
			{Name: &[]byte("G\x00")[0], Value: &[]byte("isBar\x00")[0]},
		}, true, true)
		p.Register()
	}

	p = objc.GetProtocol("MyCustomProtocol")

	for _, protocol := range p.CopyProtocolList() {
		fmt.Println(protocol.Name())
	}
	for _, property := range p.CopyPropertyList(true, true) {
		fmt.Println(property.Name(), property.Attributes())
	}
	for _, method := range p.CopyMethodDescriptionList(true, true) {
		fmt.Println(method.Name(), method.Types())
	}

	// Output:
	// NSObject
	// accessibilityElement TB,GisBar
	// isFoo B16@0:8
}

func TestNSArraySliceReturn(t *testing.T) {
	_, err := purego.Dlopen("/System/Library/Frameworks/Foundation.framework/Foundation", purego.RTLD_GLOBAL|purego.RTLD_NOW)
	if err != nil {
		t.Fatal(err)
	}

	sel_processInfo := objc.RegisterName("processInfo")
	sel_arguments := objc.RegisterName("arguments")
	sel_UTF8String := objc.RegisterName("UTF8String")

	processInfo := objc.ID(objc.GetClass("NSProcessInfo")).Send(sel_processInfo)
	if processInfo == 0 {
		t.Fatal("NSProcessInfo.processInfo returned nil")
	}

	t.Run("send_slice_return", func(t *testing.T) {
		result := objc.Send[[]objc.ID](processInfo, sel_arguments)
		if len(result) == 0 {
			t.Fatal("expected at least 1 argument (the test binary)")
		}
		t.Logf("Send[[]objc.ID] returned %d elements", len(result))
		for i, elem := range result {
			str := objc.Send[string](elem, sel_UTF8String)
			t.Logf("  arg[%d]: %s", i, str)
		}
	})

	t.Run("nsarray_to_slice", func(t *testing.T) {
		arrayID := processInfo.Send(sel_arguments)
		if arrayID == 0 {
			t.Fatal("arguments returned nil")
		}
		result := objc.NSArrayToSlice(arrayID)
		if len(result) == 0 {
			t.Fatal("expected at least 1 argument (the test binary)")
		}
		t.Logf("NSArrayToSlice returned %d elements", len(result))
		for i, elem := range result {
			str := objc.Send[string](elem, sel_UTF8String)
			t.Logf("  arg[%d]: %s", i, str)
		}
	})

	t.Run("unsupported_element_type", func(t *testing.T) {
		defer func() {
			r := recover()
			if r == nil {
				t.Fatal("expected panic but did not get one")
			}
			got, ok := r.(string)
			if !ok {
				t.Fatalf("expected string panic, got %T: %v", r, r)
			}
			want := "objc: Send with slice return only supports []objc.ID, got []string"
			if got != want {
				t.Fatalf("unexpected panic message:\n got: %s\nwant: %s", got, want)
			}
			t.Logf("got expected panic: %s", got)
		}()
		objc.Send[[]string](processInfo, sel_arguments)
	})

	t.Run("empty_nsarray", func(t *testing.T) {
		sel_array := objc.RegisterName("array")
		emptyArray := objc.ID(objc.GetClass("NSArray")).Send(sel_array)
		if emptyArray == 0 {
			t.Fatal("NSArray.array returned nil")
		}
		result := objc.NSArrayToSlice(emptyArray)
		if result != nil {
			t.Fatalf("expected nil for empty array, got len=%d", len(result))
		}
		// Also test via NSMutableArray
		mutableArray := objc.ID(objc.GetClass("NSMutableArray")).Send(objc.RegisterName("new"))
		result2 := objc.NSArrayToSlice(mutableArray)
		if result2 != nil {
			t.Fatalf("expected nil for empty mutable array, got len=%d", len(result2))
		}
	})

	t.Run("single_element_array", func(t *testing.T) {
		sel_arrayWithObject := objc.RegisterName("arrayWithObject:")
		sel_stringWithUTF8String := objc.RegisterName("stringWithUTF8String:")

		str := objc.ID(objc.GetClass("NSString")).Send(sel_stringWithUTF8String, "hello\x00")
		singleArray := objc.ID(objc.GetClass("NSArray")).Send(sel_arrayWithObject, str)
		if singleArray == 0 {
			t.Fatal("arrayWithObject: returned nil")
		}
		result := objc.NSArrayToSlice(singleArray)
		if len(result) != 1 {
			t.Fatalf("expected 1 element, got %d", len(result))
		}
		got := objc.Send[string](result[0], sel_UTF8String)
		if got != "hello" {
			t.Fatalf("expected 'hello', got %q", got)
		}
	})

	t.Run("string_components", func(t *testing.T) {
		sel_stringWithUTF8String := objc.RegisterName("stringWithUTF8String:")
		sel_componentsSeparatedByString := objc.RegisterName("componentsSeparatedByString:")

		full := objc.ID(objc.GetClass("NSString")).Send(sel_stringWithUTF8String, "a,b,c\x00")
		sep := objc.ID(objc.GetClass("NSString")).Send(sel_stringWithUTF8String, ",\x00")

		parts := objc.Send[[]objc.ID](full, sel_componentsSeparatedByString, sep)
		if len(parts) != 3 {
			t.Fatalf("expected 3 parts, got %d", len(parts))
		}
		want := []string{"a", "b", "c"}
		for i, elem := range parts {
			got := objc.Send[string](elem, sel_UTF8String)
			if got != want[i] {
				t.Fatalf("part[%d]: expected %q, got %q", i, want[i], got)
			}
		}
		t.Logf("componentsSeparatedByString: returned %v", want)
	})

	t.Run("dictionary_keys_values", func(t *testing.T) {
		sel_stringWithUTF8String := objc.RegisterName("stringWithUTF8String:")
		sel_allKeys := objc.RegisterName("allKeys")
		sel_allValues := objc.RegisterName("allValues")
		sel_dictionaryWithObject_forKey := objc.RegisterName("dictionaryWithObject:forKey:")

		key := objc.ID(objc.GetClass("NSString")).Send(sel_stringWithUTF8String, "myKey\x00")
		val := objc.ID(objc.GetClass("NSString")).Send(sel_stringWithUTF8String, "myValue\x00")
		dict := objc.ID(objc.GetClass("NSDictionary")).Send(sel_dictionaryWithObject_forKey, val, key)
		if dict == 0 {
			t.Fatal("dictionaryWithObject:forKey: returned nil")
		}

		keys := objc.Send[[]objc.ID](dict, sel_allKeys)
		if len(keys) != 1 {
			t.Fatalf("expected 1 key, got %d", len(keys))
		}
		gotKey := objc.Send[string](keys[0], sel_UTF8String)
		if gotKey != "myKey" {
			t.Fatalf("expected key 'myKey', got %q", gotKey)
		}

		values := objc.Send[[]objc.ID](dict, sel_allValues)
		if len(values) != 1 {
			t.Fatalf("expected 1 value, got %d", len(values))
		}
		gotVal := objc.Send[string](values[0], sel_UTF8String)
		if gotVal != "myValue" {
			t.Fatalf("expected value 'myValue', got %q", gotVal)
		}
		t.Logf("NSDictionary keys=%v values=%v", []string{gotKey}, []string{gotVal})
	})

	t.Run("nil_array_id", func(t *testing.T) {
		// In ObjC, sending messages to nil returns zero. So NSArrayToSlice(0)
		// should call Send[uint](0, sel_count) which returns 0 → nil slice.
		result := objc.NSArrayToSlice(0)
		if result != nil {
			t.Fatalf("expected nil for nil ID, got len=%d", len(result))
		}
	})
}
