//go:build go1.25

package xreflect

import "reflect"

func TypeAssert[T any](v reflect.Value) (T, bool) {
	return reflect.TypeAssert[T](v)
}
