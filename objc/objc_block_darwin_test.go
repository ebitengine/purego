// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024 The Ebitengine Authors

package objc_test

import (
	"fmt"
	"reflect"
	"testing"
	"unsafe"

	"github.com/ebitengine/purego"
	"github.com/ebitengine/purego/objc"
)

func ExampleNewBlock() {
	_, err := purego.Dlopen("/System/Library/Frameworks/Foundation.framework/Foundation", purego.RTLD_GLOBAL)
	if err != nil {
		panic(err)
	}

	var count = 0
	block := objc.NewBlock(
		func(block objc.Block, line objc.ID, stop *bool) {
			count++
			fmt.Printf("LINE %d: %s\n", count, objc.Send[string](line, objc.RegisterName("UTF8String")))
			(*stop) = (count == 3)
		},
	)
	defer block.Release()

	lines := objc.ID(objc.GetClass("NSString")).Send(objc.RegisterName("stringWithUTF8String:"), "Alpha\nBeta\nGamma\nDelta\nEpsilon")
	defer lines.Send(objc.RegisterName("release"))

	lines.Send(objc.RegisterName("enumerateLinesUsingBlock:"), block)
	// Output:
	// LINE 1: Alpha
	// LINE 2: Beta
	// LINE 3: Gamma
}

func ExampleInvokeBlock() {
	type vector struct {
		X, Y, Z float64
	}

	block := objc.NewBlock(
		func(block objc.Block, v1, v2 *vector) *vector {
			return &vector{
				X: v1.Y*v2.Z - v1.Z*v2.Y,
				Y: v1.Z*v2.X - v1.X*v2.Z,
				Z: v1.X*v2.Y - v1.Y*v2.X,
			}
		},
	)
	defer block.Release()

	fmt.Println(*objc.InvokeBlock[*vector](
		block,
		&vector{X: 0.1, Y: 2.3, Z: 4.5},
		&vector{X: 6.7, Y: 8.9, Z: 0.1},
	))
	// Output: {-39.82 30.14 -14.52}
}

func TestNewBlockAndBlockGetImplementation(t *testing.T) {
	t.Parallel()
	values := [14]reflect.Value{
		reflect.ValueOf(true),
		reflect.ValueOf(2),
		reflect.ValueOf(int8(3)),
		reflect.ValueOf(int16(4)),
		reflect.ValueOf(int32(5)),
		reflect.ValueOf(int64(6)),
		reflect.ValueOf(&[]uint8("seven\x00")[0]),
		reflect.ValueOf(uint(8)),
		reflect.ValueOf(uint8(9)),
		reflect.ValueOf(uint16(10)),
		reflect.ValueOf(uint32(11)),
		reflect.ValueOf(uint64(12)),
		reflect.ValueOf(objc.GetClass("NSObject")),
		reflect.ValueOf(unsafe.Pointer(objc.GetProtocol("NSObject"))),
	}

	var argumentRecurse func([]reflect.Value, []reflect.Type, func([]reflect.Value, []reflect.Type))
	argumentRecurse = func(argumentValues []reflect.Value, argumentTypes []reflect.Type, execute func([]reflect.Value, []reflect.Type)) {
		if len(argumentValues) == cap(argumentValues) {
			execute(argumentValues, argumentTypes)
			return
		}

		argumentValues = append(argumentValues, reflect.Value{})
		argumentTypes = append(argumentTypes, nil)
		for index := 0; index < len(values); index++ {
			argumentValues[len(argumentValues)-1] = values[index]
			argumentTypes[len(argumentTypes)-1] = values[index].Type()
			argumentRecurse(argumentValues, argumentTypes, execute)
		}
	}

	for out := 0; out <= len(values); out++ {
		returnValues := make([]reflect.Value, 0, 1)
		returnTypes := make([]reflect.Type, 0, 1)
		if out < len(values) {
			returnValues = append(returnValues, values[out])
			returnTypes = append(returnTypes, returnValues[0].Type())
		}

		for in := 1; in < 3; in++ {
			argumentValues := make([]reflect.Value, 1, in)
			argumentTypes := make([]reflect.Type, 1, len(argumentValues))
			argumentValues[0] = reflect.ValueOf(objc.Block(0))
			argumentTypes[0] = argumentValues[0].Type()

			argumentRecurse(argumentValues, argumentTypes, func(argumentValues []reflect.Value, argumentTypes []reflect.Type) {
				functionType := reflect.FuncOf(argumentTypes, returnTypes, false)
				block := objc.NewBlock(
					reflect.MakeFunc(
						functionType,
						func(args []reflect.Value) (results []reflect.Value) {
							for index, argumentValue := range args {
								if argumentValue.Interface() != argumentValues[index].Interface() {
									t.Fatalf("%v: arg[%d]: %v != %v", functionType, index, argumentValue.Interface(), argumentValues[index].Interface())
								}
							}
							return returnValues
						},
					).Interface(),
				)
				defer block.Release()
				argumentValues[0] = reflect.ValueOf(block)

				fptr := reflect.New(functionType)
				block.GetImplementation(fptr.Interface())

				for index, returnValue := range fptr.Elem().Call(argumentValues) {
					if returnValue.Interface() != returnValues[index].Interface() {
						t.Fatalf("%v: return: %v != %v", functionType, returnValue.Interface(), returnValues[index].Interface())
					}
				}
			})
		}
	}
}

func TestBlockCopyAndBlockRelease(t *testing.T) {
	t.Parallel()

	refCount := 0
	block := objc.NewBlock(
		func(objc.Block) {
			refCount++
		},
	)
	defer block.Release()
	refCount++

	copies := make([]objc.Block, 17)
	copies[0] = block
	for index := 1; index < len(copies); index++ {
		if refCount != index {
			t.Fatalf("refCount: %d != %d", refCount, index)
		}

		copies[index] = copies[index-1].Copy()
		if copies[index] != block {
			t.Fatalf("Block.Copy(): %v != %v", copies[index], block)
		}
		copies[index].Invoke()
	}

	for _, copy := range copies[1:] {
		copy.Release()
		refCount--
	}
	refCount--

	block.Invoke()
	if refCount != 1 {
		t.Fatalf("refCount: %d != 1", refCount)
	}
}
