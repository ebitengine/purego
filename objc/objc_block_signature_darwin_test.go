// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 The Ebitengine Authors

package objc

import (
	"reflect"
	"slices"
	"structs"
	"testing"
)

func TestSplitSignature(t *testing.T) {
	tests := []struct {
		sig  string
		want []string
	}{
		{"v24@?0q8d16", []string{"v", "@?", "q", "d"}},
		{"q28@?0i8c12s16q20", []string{"q", "@?", "i", "c", "s", "q"}},
		{`v32@?0@"NSString"8@"NSError"16q24`, []string{"v", "@?", `@"NSString"`, `@"NSError"`, "q"}},
		{"v24@?0@?<v@?q>8q16", []string{"v", "@?", "@?<v@?q>", "q"}},
		{"v16@?0@?<v@?@?<v@?>>8", []string{"v", "@?", "@?<v@?@?<v@?>>"}},
		{`v16@?0@"<NSCopying>"8`, []string{"v", "@?", `@"<NSCopying>"`}},
		{"{Big=qqqq}16@?0q8", []string{"{Big=qqqq}", "@?", "q"}},
		{"v16@?0^{S=[2{T=i}]}8", []string{"v", "@?", "^{S=[2{T=i}]}"}},
		{"v24@?0r*8Vv16", []string{"v", "@?", "*", "v"}},
		{"v16@?0^r^v8", []string{"v", "@?", "^r^v"}},
		{`{P="x"d"y"d}8@?0`, []string{`{P="x"d"y"d}`, "@?"}},
	}
	for _, tt := range tests {
		got, err := splitSignature(tt.sig)
		if err != nil {
			t.Errorf("splitSignature(%q): %v", tt.sig, err)
			continue
		}
		if !slices.Equal(got, tt.want) {
			t.Errorf("splitSignature(%q) = %q, want %q", tt.sig, got, tt.want)
		}
	}

	for _, sig := range []string{"v8@?0{S=i", `v8@?0@"NSString`, "v8@?0@?<v"} {
		if got, err := splitSignature(sig); err == nil {
			t.Errorf("splitSignature(%q) = %q, want an error", sig, got)
		}
	}
}

func TestEncodingLayout(t *testing.T) {
	tests := []struct {
		enc         string
		size, align uintptr
		scalars     []abiScalar
	}{
		{"v", 0, 0, nil},
		{"B", 1, 1, []abiScalar{{0, '1'}}},
		{"S", 2, 2, []abiScalar{{0, '2'}}},
		{"L", 4, 4, []abiScalar{{0, '4'}}},
		{"^v", 8, 8, []abiScalar{{0, '8'}}},
		{"@?<v@?q>", 8, 8, []abiScalar{{0, '8'}}},
		{`@"NSString"`, 8, 8, []abiScalar{{0, '8'}}},
		{"r*", 8, 8, []abiScalar{{0, '8'}}},
		{"f", 4, 4, []abiScalar{{0, 'f'}}},
		{"{Big=qqqq}", 32, 8, []abiScalar{{0, '8'}, {8, '8'}, {16, '8'}, {24, '8'}}},
		{"{CGRect={CGPoint=dd}{CGSize=dd}}", 32, 8, []abiScalar{{0, 'd'}, {8, 'd'}, {16, 'd'}, {24, 'd'}}},
		{"{BoolFloat=Bf}", 8, 4, []abiScalar{{0, '1'}, {4, 'f'}}},
		{`{P="x"d"y"i}`, 16, 8, []abiScalar{{0, 'd'}, {8, '4'}}},
		{"{S=c[3s]q}", 16, 8, []abiScalar{{0, '1'}, {2, '2'}, {4, '2'}, {6, '2'}, {8, '8'}}},
		{"{S=[2{T=cd}]c}", 40, 8, []abiScalar{{0, '1'}, {8, 'd'}, {16, '1'}, {24, 'd'}, {32, '1'}}},
	}
	for _, tt := range tests {
		got, err := encodingLayout(tt.enc)
		if err != nil {
			t.Errorf("encodingLayout(%q): %v", tt.enc, err)
			continue
		}
		if got.size != tt.size || got.align != tt.align || !slices.Equal(got.scalars, tt.scalars) {
			t.Errorf("encodingLayout(%q) = size %d, align %d, %v; want size %d, align %d, %v",
				tt.enc, got.size, got.align, got.scalars, tt.size, tt.align, tt.scalars)
		}
	}

	for _, enc := range []string{"(U=id)", "{S=b3}", "D", "{Opaque}"} {
		if got, err := encodingLayout(enc); err == nil {
			t.Errorf("encodingLayout(%q) = %v, want an error", enc, got)
		}
	}
}

func TestGoLayoutUnsupported(t *testing.T) {
	for _, typ := range []reflect.Type{reflect.TypeFor[func()](), reflect.TypeFor[any](), reflect.TypeFor[[]int](), reflect.TypeFor[map[int]int]()} {
		if got, err := goLayout(typ); err == nil {
			t.Errorf("goLayout(%v) = %v, want an error", typ, got)
		}
	}
}

func TestLayoutMatches(t *testing.T) {
	type point struct {
		_    structs.HostLayout
		X, Y float64
	}
	type rect struct {
		_      structs.HostLayout
		Origin point
		Size   point
	}
	type mixed struct {
		_ structs.HostLayout
		A [3]float32
		B bool
		C *int
	}
	// padding written out, as the RegisterFunc documentation asks for.
	type boolFloat struct {
		_ structs.HostLayout
		B bool
		_ [3]byte
		F float32
	}
	type boolFloatImplicit struct {
		_ structs.HostLayout
		B bool
		F float32
	}
	// a blank field standing in for a member the caller does not need.
	type skipInt struct {
		_ structs.HostLayout
		A int32
		_ int32
		C int64
	}
	type skipInts struct {
		_ structs.HostLayout
		_ [2]int32
		C int64
	}
	type skipFloat struct {
		_ structs.HostLayout
		A float32
		_ float32
	}
	type intOverFloat struct {
		_ structs.HostLayout
		A float32
		_ int32
	}
	type floatOverInt struct {
		_ structs.HostLayout
		A int32
		_ float32
	}
	type tailPadding struct {
		_ structs.HostLayout
		A int64
		B int8
		_ [7]byte
	}
	type namedPadding struct {
		_   structs.HostLayout
		B   bool
		Pad [3]byte
		F   float32
	}

	tests := []struct {
		enc  string
		typ  reflect.Type
		want bool
	}{
		{"B", reflect.TypeFor[bool](), true},
		{"c", reflect.TypeFor[bool](), true},
		{"i", reflect.TypeFor[int32](), true},
		{"i", reflect.TypeFor[int](), false},
		{"q", reflect.TypeFor[int](), true},
		{"q", reflect.TypeFor[int32](), false},
		{"q", reflect.TypeFor[float64](), false},
		{"d", reflect.TypeFor[float64](), true},
		{"d", reflect.TypeFor[float32](), false},
		{"@", reflect.TypeFor[ID](), true},
		{"@?", reflect.TypeFor[Block](), true},
		{"v", reflect.TypeFor[int](), false},
		{"{CGRect={CGPoint=dd}{CGSize=dd}}", reflect.TypeFor[rect](), true},
		{"{CGPoint=dd}", reflect.TypeFor[rect](), false},
		{"{CGRect={CGPoint=dd}{CGSize=dd}}", reflect.TypeFor[point](), false},
		{"{S=[3f]B^i}", reflect.TypeFor[mixed](), true},
		{"{S=[3f]i^i}", reflect.TypeFor[mixed](), false},
		{"{BoolFloat=Bf}", reflect.TypeFor[boolFloat](), true},
		{"{BoolFloat=Bf}", reflect.TypeFor[boolFloatImplicit](), true},
		{"{S=Bi}", reflect.TypeFor[boolFloat](), false},
		{"{S=iiq}", reflect.TypeFor[skipInt](), true},
		{"{S=issq}", reflect.TypeFor[skipInt](), true},
		{"{S=ifq}", reflect.TypeFor[skipInt](), false},
		{"{S=qq}", reflect.TypeFor[skipInt](), false},
		{"{S=iiq}", reflect.TypeFor[skipInts](), true},
		{"{S=qq}", reflect.TypeFor[skipInts](), true},
		{"{S=ff}", reflect.TypeFor[skipFloat](), true},
		{"{S=fi}", reflect.TypeFor[skipFloat](), false},
		{"{S=f}", reflect.TypeFor[skipFloat](), false},
		{"{S=ff}", reflect.TypeFor[intOverFloat](), false},
		{"{S=fi}", reflect.TypeFor[intOverFloat](), true},
		{"{S=ii}", reflect.TypeFor[floatOverInt](), false},
		{"{S=qc}", reflect.TypeFor[tailPadding](), true},
		{"{BoolFloat=Bf}", reflect.TypeFor[namedPadding](), false},
	}
	for _, tt := range tests {
		c, err := encodingLayout(tt.enc)
		if err != nil {
			t.Errorf("encodingLayout(%q): %v", tt.enc, err)
			continue
		}
		g, err := goLayout(tt.typ)
		if err != nil {
			t.Errorf("goLayout(%v): %v", tt.typ, err)
			continue
		}
		if got := c.matches(g); got != tt.want {
			t.Errorf("%q matches %v = %v, want %v", tt.enc, tt.typ, got, tt.want)
		}
	}
}
